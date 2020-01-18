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
	"io"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"github.com/greenbrew/rest/api"
)

// Target is the function receiving events
type Target func(interface{})

// The Operation interface represents a currently running operation.
type Operation interface {
	AddHandler(function func(api.Operation)) (target Target, err error)
	Cancel() (err error)
	Get() (op api.Operation)
	RemoveHandler(target Target) (err error)
	Refresh() (err error)
	Wait(ctx context.Context) (err error)
}

// The Operations interface represents operations exposed API methods
type Operations interface {
	ListOperationUUIDs() (uuids []string, err error)
	ListOperations() (operations []api.Operation, err error)
	RetrieveOperationByID(uuid string) (op *api.Operation, etag string, err error)
	WaitForOperationToFinish(uuid string, timeout time.Duration) (op *api.Operation, err error)
	DeleteOperation(uuid string) (err error)
}

// The Client interface represents all available REST client operations
type Client interface {
	SetTransportTimeout(timeout time.Duration)

	QueryStruct(method, path string, params QueryParams, header http.Header, body io.Reader, ETag string, target interface{}) (etag string, err error)
	QueryOperation(method, path string, params QueryParams, header http.Header, body io.Reader, ETag string) (operation Operation, etag string, err error)
	CallAPI(method, path string, params QueryParams, header http.Header, body io.Reader, ETag string) (response *api.Response, etag string, err error)

	Websocket(resource string) (conn *websocket.Conn, err error)

	// Event handling functions
	GetEvents() (listener *EventListener, err error)
}
