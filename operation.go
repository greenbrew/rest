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

package rest

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/greenbrew/rest/api"
	"github.com/greenbrew/rest/logger"
	"github.com/greenbrew/rest/pool"
)

// Operation struct holding metadata for an API operation, including handlers
// for run, cancel or socket connection; metadata, status or dates it was created, updated, etc..
type Operation struct {
	id          string
	createdAt   time.Time
	updatedAt   time.Time
	status      api.StatusCode
	url         string
	resources   map[string][]string
	metadata    map[string]interface{}
	errStr      string
	description string
	cancel      context.CancelFunc

	// Operation handlers
	onRun    func(*Operation) error
	onCancel func(*Operation) error

	// Channels used for error reporting and state tracking of background actions
	doneCh chan error

	// Locking for concurent access to the operation
	mux sync.RWMutex

	// Reference to the queue where pushing run, cancel, etc.. jobs
	operationsQueue *pool.JobChannel

	// Cached map of in progress operations reference
	cache *cache

	// Reference to the events manager
	events *eventsManager
}

// Render writes in response the operation details, included the list of resources (urls)
func (op *Operation) Render() (string, *api.Operation, error) {
	op.mux.RLock()
	defer op.mux.RUnlock()

	// Setup the resource URLs
	resources := op.resources
	if resources != nil {
		tmpResources := make(map[string][]string)
		for key, value := range resources {
			var values []string
			for _, c := range value {
				values = append(values, api.Path(key, c))
			}
			tmpResources[key] = values
		}
		resources = tmpResources
	}

	return op.url, &api.Operation{
		ID:          op.id,
		Description: op.description,
		CreatedAt:   op.createdAt,
		UpdatedAt:   op.updatedAt,
		Status:      op.status.String(),
		StatusCode:  op.status,
		Resources:   resources,
		Metadata:    op.metadata,
		Err:         op.errStr,
	}, nil
}

// WaitFinal waits for the operation to be completed
func (op *Operation) WaitFinal(timeout int) error {
	// Check current state
	if op.getStatus().IsFinal() {
		return nil
	}

	// Wait indefinitely
	if timeout == -1 {
		<-op.doneCh
		return nil
	}

	// Wait until timeout
	if timeout > 0 {
		timer := time.NewTimer(time.Duration(timeout) * time.Second)
		select {
		case <-op.doneCh:
			return nil
		case <-timer.C:
			return errors.Errorf("Timeout waiting for operation %v", op.getID())
		}
	}

	return nil
}

// Run executes internal 'onRun' provided handler
func (op *Operation) Run() error {
	if op.getStatus() != api.Pending {
		return errors.New("Only pending operations can be started")
	}

	op.setStatus(api.Running)

	if op.onRun != nil {
		job := func() {
			err := op.onRun(op)
			if err != nil {
				op.setStatus(api.Failure)
				op.setErrStr(SmartError(err).String())
				op.done()

				logger.Errorf("Failure for operation: %s: %s", op.getID(), err)

				_, md, _ := op.Render()
				op.events.send(md)
				return
			}

			op.setStatus(api.Success)
			op.done()

			logger.Debugf("Success for operation: %s", op.getID())
			_, md, _ := op.Render()
			op.events.send(md)
		}

		// Enqueue job if queue is enabled. Execute it now otherwise
		if op.operationsQueue != nil {
			op.operationsQueue.Push(job)
		} else {
			go job()
		}
	}

	logger.Debugf("Started operation: %s", op.getID())
	_, md, _ := op.Render()
	op.events.send(md)

	return nil
}

// Cancel calls internal context cancel() method
func (op *Operation) Cancel() error {
	if op.getStatus() != api.Running {
		return errors.New("Only running operations can be cancelled")
	}

	op.setStatus(api.Cancelling)

	if op.onCancel != nil {
		job := func() {
			err := op.onCancel(op)
			if err != nil {
				op.setStatus(api.Failure)
				op.setErrStr(SmartError(err).String())
				op.done()

				logger.Errorf("Failure for cancelling operation: %s: %s", op.getID(), err)

				_, md, _ := op.Render()
				op.events.send(md)
				return
			}

			op.setStatus(api.Cancelled)
			op.setErrStr("Operation cancelled")
			op.done()

			logger.Debugf("Cancelled operation: %s", op.getID())
			_, md, _ := op.Render()
			op.events.send(md)
		}

		// Enqueue job if queue is enabled. Execute it now otherwise
		if op.operationsQueue != nil {
			op.operationsQueue.Push(job)
		} else {
			go job()
		}
	}

	logger.Debugf("Cancelling operation: %s", op.getID())
	_, md, _ := op.Render()
	op.events.send(md)

	if op.onCancel == nil {
		op.setStatus(api.Cancelled)
		op.setErrStr("Operation cancelled")
		op.done()

		logger.Debugf("Cancelled operation: %s", op.getID())
		_, md, _ := op.Render()
		op.events.send(md)
	}

	return nil
}

func (op *Operation) done() {
	// Ensure that the operation is still enabled
	select {
	case <-op.doneCh:
		return
	default:
	}

	op.mux.Lock()
	defer op.mux.Unlock()

	op.onRun = nil
	op.cancel = nil
	close(op.doneCh)

	// TODO Watch out this. Original code delays 5 secs to do it
	op.cache.deleleOperationByID(op.id)
}

func (op *Operation) read(fn func() interface{}) interface{} {
	op.mux.RLock()
	defer op.mux.RUnlock()
	return fn()
}

func (op *Operation) write(fn func()) {
	op.mux.Lock()
	defer op.mux.Unlock()
	fn()
}

func (op *Operation) getID() string {
	return op.read(func() interface{} {
		return op.id
	}).(string)
}

func (op *Operation) getStatus() api.StatusCode {
	return op.read(func() interface{} {
		return op.status
	}).(api.StatusCode)
}

func (op *Operation) setStatus(status api.StatusCode) {
	op.write(func() {
		op.status = status
	})
}

func (op *Operation) setErrStr(errStr string) {
	op.write(func() {
		op.errStr = errStr
	})
}
