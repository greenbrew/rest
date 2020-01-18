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
	"io"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"github.com/greenbrew/rest/api"
)

// MockClient mock of Client interface for testing purpose
type MockClient struct {
	Response      *api.Response
	Operation     *MockOperation
	EventListener *EventListener
	ETag          string
}

// SetTransportTimeout mocked
func (c *MockClient) SetTransportTimeout(timeout time.Duration) {
}

// QueryStruct mocked
func (c *MockClient) QueryStruct(method, path string, params QueryParams, header http.Header, body io.Reader, ETag string, target interface{}) (string, error) {
	err := c.Response.MetadataAsStruct(&target)
	return c.ETag, err
}

// QueryOperation mocked
func (c *MockClient) QueryOperation(method, path string, params QueryParams, header http.Header, body io.Reader, ETag string) (op Operation, etag string, err error) {
	return c.Operation, c.ETag, nil
}

// CallAPI mocked
func (c *MockClient) CallAPI(method, path string, params QueryParams, header http.Header, body io.Reader, ETag string) (response *api.Response, etag string, err error) {
	return c.Response, c.ETag, nil
}

// Websocket mocked
func (c *MockClient) Websocket(resource string) (conn *websocket.Conn, err error) {
	return &websocket.Conn{}, nil
}

// GetEvents mocked
func (c *MockClient) GetEvents() (eventListener *EventListener, err error) {
	return c.EventListener, nil
}
