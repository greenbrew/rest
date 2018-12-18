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
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/pborman/uuid"
	"github.com/greenbrew/rest/logger"
)

// TODO see if this can be included in EventsManager object
var websocketUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type eventsResponse struct {
	req    *http.Request
	events *eventsManager
}

func (r *eventsResponse) Render(w http.ResponseWriter) error {
	typeStr := r.req.FormValue("type")
	if typeStr == "" {
		typeStr = "logging,operation"
	}

	c, err := websocketUpgrader.Upgrade(w, r.req, nil)
	if err != nil {
		return err
	}

	listener := &eventsListener{
		id:         uuid.NewRandom().String(),
		connection: c,
		doneCh:     make(chan struct{}),
	}

	r.events.mux.Lock()
	r.events.listeners[listener.id] = listener
	r.events.mux.Unlock()

	logger.Debugf("New event listener: %s", listener.id)

	<-listener.doneCh

	return nil
}

func (r *eventsResponse) String() string {
	return "event handler"
}

// Return true if this an API request coming from a cluster node that is
// notifying us of some user-initiated API request that needs some action to be
// taken on this node as well.
func isClusterNotification(r *http.Request) bool {
	return r.Header.Get("User-Agent") == "cluster-notifier"
}
