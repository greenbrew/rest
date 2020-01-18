// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2018 Roberto Mier Escandon <rmescandon@gmail.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/greenbrew/rest/api"
)

// Operation wrapper type for operations response allowing certain additional logic
// like blocking current thread until operations completes or cancel it.
type operation struct {
	api.Operation

	c            *operations
	listener     *EventListener
	handlerReady bool
	mux          sync.Mutex
	doneCh       chan bool
}

// AddHandler adds a function to be called whenever an event is received
func (op *operation) AddHandler(function func(api.Operation)) (Target, error) {
	// Make sure we have a listener setup
	err := op.setupListener()
	if err != nil {
		return nil, err
	}

	// Make sure we're not racing with ourselves
	op.mux.Lock()
	defer op.mux.Unlock()

	// If we're done already, just return
	if op.StatusCode.IsFinal() {
		return nil, nil
	}

	// Wrap the function to filter unwanted messages
	wrapped := func(data interface{}) {
		newOp := op.extractOperation(data)
		if newOp == nil {
			return
		}

		function(*newOp)
	}

	return op.listener.AddHandler(wrapped)
}

// Cancel will request that server cancels the operation (if supported)
func (op *operation) Cancel() error {
	return op.c.DeleteOperation(op.ID)
}

// Get returns the API operation struct
func (op *operation) Get() api.Operation {
	return op.Operation
}

// RemoveHandler removes a function to be called whenever an event is received
func (op *operation) RemoveHandler(target Target) error {
	// Make sure we're not racing with ourselves
	op.mux.Lock()
	defer op.mux.Unlock()

	// If the listener is gone, just return
	if op.listener == nil {
		return nil
	}

	return op.listener.RemoveHandler(target)
}

// Refresh pulls the current version of the operation and updates the struct
func (op *operation) Refresh() error {
	// Don't bother with a manual update if we are listening for events
	if op.handlerReady {
		return nil
	}

	// Get the current version of the operation
	newOp, _, err := op.c.RetrieveOperationByID(op.ID)
	if err != nil {
		return err
	}

	// Update the operation struct
	op.Operation = *newOp

	return nil
}

// Wait lets you wait until the operation reaches a final state
func (op *operation) Wait(ctx context.Context) error {
	// Check if not done already
	if op.StatusCode.IsFinal() {
		return op.theError()
	}

	// Make sure we have a listener setup
	err := op.setupListener()
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		if err := op.Cancel(); err != nil {
			return fmt.Errorf("Could not cancel operation %v: %v", op.ID, err)
		}
		switch ctx.Err() {
		case context.Canceled:
			return errors.New("Operation cancelled")
		case context.DeadlineExceeded:
			return errors.New("Operation timeout")
		default:
			return ctx.Err()
		}
	case <-op.doneCh:
	}

	return op.theError()
}

func (op *operation) theError() error {
	if len(op.Err) > 0 {
		return errors.New(op.Err)
	}
	return nil
}

func (op *operation) setupListener() error {
	// Make sure we're not racing with ourselves
	op.mux.Lock()
	defer op.mux.Unlock()

	// We already have a listener setup
	if op.handlerReady {
		return nil
	}

	// Get a new listener
	if op.listener == nil {
		listener, err := op.c.GetEvents()
		if err != nil {
			return err
		}

		op.listener = listener
	}

	// Setup the handler
	readyCh := make(chan struct{})
	_, err := op.listener.AddHandler(func(data interface{}) {
		<-readyCh

		// Get an operation struct out of this data
		newOp := op.extractOperation(data)
		if newOp == nil {
			return
		}

		// We don't want concurrency while processing events
		op.mux.Lock()
		defer op.mux.Unlock()

		// Check if we're done already (because of another event)
		if op.listener == nil {
			return
		}

		// Update the struct
		op.Operation = *newOp

		// And check if we're done
		if op.StatusCode.IsFinal() {
			op.done()
		}
	})
	if err != nil {
		op.done()
		close(readyCh)
		return err
	}

	// Monitor event listener
	go func() {
		<-readyCh

		op.mux.Lock()
		// Check if we're done already (because of another event)
		if op.listener == nil {
			op.mux.Unlock()
			return
		}
		op.mux.Unlock()

		// Wait for the listener or operation to be done
		select {
		case <-op.listener.doneCh:
			op.mux.Lock()
			if op.listener != nil {
				op.Err = fmt.Sprintf("%v", op.listener.err)
				close(op.doneCh)
			}
			op.mux.Unlock()
		case <-op.doneCh:
			return
		}
	}()

	// And do a manual refresh to avoid races
	err = op.Refresh()
	if err != nil {
		op.done()
		close(readyCh)
		return err
	}

	// Check if not done already
	if op.StatusCode.IsFinal() {
		op.done()

		if len(op.Err) > 0 {
			return errors.New(op.Err)
		}
		return nil
	}

	// Start processing background updates
	op.handlerReady = true
	close(readyCh)

	return nil
}

func (op *operation) done() {
	op.mux.Lock()
	defer op.mux.Unlock()

	if op.listener != nil {
		op.listener.Disconnect()
		op.listener = nil
	}
	close(op.doneCh)
}

func (op *operation) extractOperation(data interface{}) *api.Operation {
	// Extract the metadata
	meta, ok := data.(map[string]interface{})["metadata"]
	if !ok {
		return nil
	}

	// And attempt to decode it as JSON operation data
	encoded, err := json.Marshal(meta)
	if err != nil {
		return nil
	}

	newOp := api.Operation{}
	err = json.Unmarshal(encoded, &newOp)
	if err != nil {
		return nil
	}

	// And now check that it's what we want
	if newOp.ID != op.ID {
		return nil
	}

	return &newOp
}
