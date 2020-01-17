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
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/greenbrew/rest/logger"
)

type eventsManager struct {
	listeners map[string]*eventsListener
	mux       sync.Mutex
}

type eventsListener struct {
	id         string
	connection *websocket.Conn
	running    bool
	doneCh     chan struct{}
	mux        sync.Mutex
}

func (m *eventsManager) send(eventMessage interface{}) error {
	event := jmap{}
	event["timestamp"] = time.Now()
	event["metadata"] = eventMessage

	return m.broadcast(event)
}

func (m *eventsManager) broadcast(event jmap) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	m.mux.Lock()
	defer m.mux.Unlock()
	for _, listener := range m.listeners {
		go func(listener *eventsListener, body []byte) {
			// Check that the listener still exists
			if listener == nil {
				return
			}

			// Ensure there is only a single even going out at the time
			listener.mux.Lock()
			defer listener.mux.Unlock()

			// Make sure we're not done already
			if !listener.running {
				return
			}

			err = listener.connection.WriteMessage(websocket.TextMessage, body)
			if err != nil {
				// Remove the listener from the list
				m.removeListenerByID(listener.id)

				// Disconnect the listener
				listener.connection.Close()
				listener.running = false
				close(listener.doneCh)
				logger.Debugf("Disconnected event listener: %s", listener.id)
			}
		}(listener, body)
	}

	return nil
}

func (m *eventsManager) removeListenerByID(id string) {
	m.mux.Lock()
	defer m.mux.Unlock()

	if _, ok := m.listeners[id]; !ok {
		return
	}
	delete(m.listeners, id)
}
