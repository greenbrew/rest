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
	"fmt"
	"sync"
)

// EventListener is the event struct is used to interact with an event stream
type EventListener struct {
	c         *client
	connected bool
	err       error
	targets   []Target
	mux       sync.Mutex
	doneCh    chan struct{}
}

// AddHandler adds a function to be called whenever an event is received
func (e *EventListener) AddHandler(target Target) (Target, error) {
	e.mux.Lock()
	defer e.mux.Unlock()

	e.targets = append(e.targets, target)
	return target, nil
}

// RemoveHandler removes a function to be called whenever an event is received
func (e *EventListener) RemoveHandler(target Target) error {
	// Handle locking
	e.mux.Lock()
	defer e.mux.Unlock()

	for i, entry := range e.targets {
		if &entry == &target {
			e.targets = append(e.targets[:i], e.targets[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("Couldn't find this function and event types combination")
}

// Disconnect must be used once done listening for events
func (e *EventListener) Disconnect() {
	if !e.connected {
		return
	}

	// Handle locking
	e.c.mux.Lock()
	defer e.c.mux.Unlock()

	// remove reference from the client
	if e.c.listener == e {
		e.c.listener = nil
	}

	// Turn off the handler
	e.err = nil
	e.connected = false
	close(e.doneCh)
}

// Wait hangs until the server disconnects the connection or Disconnect() is called
func (e *EventListener) Wait() error {
	<-e.doneCh
	return e.err
}

// IsActive returns true if this listener is still connected, false otherwise.
func (e *EventListener) IsActive() bool {
	select {
	case <-e.doneCh:
		return false // If the chActive channel is closed we got disconnected
	default:
		return true
	}
}
