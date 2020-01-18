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
	"errors"

	"github.com/gorilla/websocket"

	"github.com/greenbrew/rest/api"
)

// MockOperation mock of Operation interface
type MockOperation struct {
	Operation api.Operation
	Websocket *websocket.Conn

	Targets []Target
}

// AddHandler mocked
func (op *MockOperation) AddHandler(function func(api.Operation)) (target Target, err error) {
	if op.Targets == nil {
		op.Targets = []Target{}
	}

	fn := func(data interface{}) {
		function(op.Operation)
	}

	op.Targets = append(op.Targets, fn)
	return fn, nil
}

// Cancel mocked
func (op *MockOperation) Cancel() error {
	return nil
}

// Get mocked
func (op *MockOperation) Get() api.Operation {
	return op.Operation
}

// GetWebsocket mocked
func (op *MockOperation) GetWebsocket() (conn *websocket.Conn, err error) {
	return op.Websocket, nil
}

// RemoveHandler mocked
func (op *MockOperation) RemoveHandler(target Target) error {
	for i, entry := range op.Targets {
		if &entry == &target {
			op.Targets = append(op.Targets[:i], op.Targets[i+1:]...)
			return nil
		}
	}
	return errors.New("Target not found")
}

// Refresh mocked
func (op *MockOperation) Refresh() error {
	return nil
}

// Wait mocked
func (op *MockOperation) Wait(ctx context.Context) error {
	return nil
}
