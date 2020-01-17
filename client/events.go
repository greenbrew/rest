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
	"encoding/json"
)

// Event handling functions

// GetEvents connects to the monitoring interface
func (c *client) GetEvents() (*EventListener, error) {
	// Prevent anything else from interacting with the listeners
	c.mux.Lock()
	defer c.mux.Unlock()

	// Setup a new listener
	c.listener = &EventListener{
		c:      c,
		doneCh: make(chan struct{}),
	}

	// Setup a new connection with the server
	resource := APIPath("events")
	conn, err := c.dialWebsocket(c.composeWebsocketPath(resource))
	if err != nil {
		return nil, err
	}

	// And spawn the listener
	go func() {
		for {
			c.mux.Lock()
			if c.listener == nil {
				// We don't need the connection anymore, disconnect
				conn.Close()
				c.mux.Unlock()
				break
			}
			c.mux.Unlock()

			_, data, err := conn.ReadMessage()
			if err != nil {
				// Prevent anything else from interacting with the listeners
				c.mux.Lock()
				defer c.mux.Unlock()

				c.listener.err = err
				c.listener.connected = false
				close(c.listener.doneCh)
				c.listener = nil
				return
			}

			// Attempt to unpack the message
			message := make(map[string]interface{})
			err = json.Unmarshal(data, &message)
			if err != nil {
				continue
			}

			// Send the message to all handlers
			c.listener.mux.Lock()
			for _, target := range c.listener.targets {
				go target(message)
			}
			c.listener.mux.Unlock()
		}
	}()

	return c.listener, nil
}
