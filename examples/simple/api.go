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

package simple

import "github.com/greenbrew/rest"

// API simple example exposed API
var API = rest.API{
	Version: "1.0",
	Commands: []*rest.Command{
		serviceCmd,
		resourcesCmd,
		resourceCmd,
	},
}

var (
	serviceCmd = &rest.Command{
		Name: "",
		GET:  serviceGet,
	}

	resourcesCmd = &rest.Command{
		Name: "resources",
		GET:  resourcesGet,
		POST: resourcesPost,
	}

	resourceCmd = &rest.Command{
		Name:   "resources/{id:[a-zA-Z0-9-_]+}",
		GET:    resourceGet,
		PUT:    resourcePut,
		DELETE: resourceDelete,
	}
)
