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
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/greenbrew/rest/api"
)

func operationsGet(d *Daemon, r *http.Request) Response {
	md := jmap{}
	recursion := IsRecursionRequest(r)
	ops := d.cache.getOperationsMap()

	for _, v := range ops {
		status := strings.ToLower(v.status.String())
		_, ok := md[status]
		if !ok {
			if recursion {
				md[status] = make([]*api.Operation, 0)
			} else {
				md[status] = make([]string, 0)
			}
		}

		if !recursion {
			md[status] = append(md[status].([]string), v.url)
			continue
		}

		_, body, err := v.Render()
		if err != nil {
			continue
		}

		md[status] = append(md[status].([]*api.Operation), body)
	}

	return SyncResponse(true, md)
}

func operationGet(d *Daemon, r *http.Request) Response {
	id := mux.Vars(r)["id"]

	op, err := d.cache.getOperationByID(id)
	if err != nil {
		return SmartError(err)
	}

	var body *api.Operation
	_, body, err = op.Render()
	if err != nil {
		return SmartError(err)
	}

	return SyncResponse(true, body)
}

func operationWaitGet(d *Daemon, r *http.Request) Response {
	var err error
	timeout := -1
	t := r.FormValue("timeout")
	if len(t) > 0 {
		timeout, err = strconv.Atoi(t)
		if err != nil {
			return SmartError(err)
		}
	}

	id := mux.Vars(r)["id"]
	op, err := d.cache.getOperationByID(id)
	if err != nil {
		return SmartError(err)
	}

	err = op.WaitFinal(timeout)
	if err != nil {
		return InternalError(err)
	}

	_, body, err := op.Render()
	if err != nil {
		return SmartError(err)
	}

	return SyncResponse(true, body)
}
