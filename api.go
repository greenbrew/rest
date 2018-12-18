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

	"github.com/greenbrew/rest/api"
)

type responseFunc func(*Daemon, *http.Request) Response

// MiddlewareFunc describes a function uses to process a HTTP request
// as a middleman
type MiddlewareFunc func(inner http.Handler) http.Handler

// API holds all the commands and metadata for a specific version of the API
type API struct {
	Version    string
	Middleware MiddlewareFunc
	Commands   []*Command
}

// Command is the basic structure for every API call.
type Command struct {
	Name       string
	Middleware MiddlewareFunc

	GET    responseFunc
	PUT    responseFunc
	POST   responseFunc
	DELETE responseFunc
	PATCH  responseFunc
}

var builtinAPI = &API{
	Version: api.Version,
	Commands: []*Command{
		eventsCmd,
		operationsCmd,
		operationCmd,
		operationWaitCmd,
	},
}

var (
	eventsCmd = &Command{
		Name: "events",
		GET:  eventsGet,
	}

	operationsCmd = &Command{
		Name: "operations",
		GET:  operationsGet,
	}

	operationCmd = &Command{
		Name: "operations/{id:[a-zA-Z0-9-_:]+}",
		GET:  operationGet,
	}

	operationWaitCmd = &Command{
		Name: "operations/{id:[a-zA-Z0-9-_:]+}/wait",
		GET:  operationWaitGet,
	}
)
