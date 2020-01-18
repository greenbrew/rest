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

package endpoints

import (
	"context"
	"net/http"
)

// Endpoint represents an endpoint were a REST service starts
type Endpoint interface {
	Name() string
	Init() (interface{}, error)
	Start(data interface{}) error
	Stop() error
}

// A default partial implementation of an endpoint. Valid as a common
// methods implementation for specific endpoints.
type endpoint struct {
	server *http.Server
}

func (e *endpoint) Stop() error {
	return e.server.Shutdown(context.Background())
}
