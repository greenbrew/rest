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

import (
	"encoding/json"
	"path/filepath"

	"github.com/gorilla/mux"

	"github.com/greenbrew/rest"
	"github.com/greenbrew/rest/random"
)

const (
	success = true
	failure = false
)

var resources map[string]interface{}

func serviceGet(r *rest.Request) rest.Response {
	return rest.SyncResponse(true, "Service Alive and Kicking")
}

func resourcesGet(r *rest.Request) rest.Response {
	list := []interface{}{}
	for k, v := range resources {
		if r.IsRecursionRequest() {
			list = append(list, v)
		} else {
			list = append(list, filepath.Join(r.HTTPRequest.URL.Path, k))
		}
	}

	return rest.SyncResponse(success, list)
}

func resourcesPost(r *rest.Request) rest.Response {
	var resource interface{}
	if err := json.NewDecoder(r.HTTPRequest.Body).Decode(&resource); err != nil {
		return rest.BadRequest(err)
	}

	id := random.New(8)
	new := map[string][]string{}
	new["resources"] = []string{id}

	run := func(op *rest.Operation) error {
		// In this case we simply store the resource in memory.
		// This operation runs asynchronously and should include all the logic for the resource creation
		if resources == nil {
			resources = make(map[string]interface{})
		}
		resources[id] = resource

		return nil
	}

	op, err := r.CreateOperation("Creating resource", new, nil, run, nil)
	if err != nil {
		return rest.SmartError(err)
	}

	return rest.OperationResponse(op)
}

func resourceGet(r *rest.Request) rest.Response {
	id := mux.Vars(r.HTTPRequest)["id"]
	if len(id) == 0 {
		return rest.NotFoundError("Resource")
	}

	res, ok := resources[id]
	if !ok {
		return rest.NotFoundError("Resource")
	}

	return rest.SyncResponse(success, res)
}

func resourcePut(r *rest.Request) rest.Response {
	id := mux.Vars(r.HTTPRequest)["id"]
	if len(id) == 0 {
		return rest.NotFoundError("Resource")
	}

	_, ok := resources[id]
	if !ok {
		return rest.NotFoundError("Resource")
	}

	var resource interface{}
	if err := json.NewDecoder(r.HTTPRequest.Body).Decode(&resource); err != nil {
		return rest.BadRequest(err)
	}

	updated := map[string][]string{}
	updated["resources"] = []string{id}

	run := func(op *rest.Operation) error {
		resources[id] = resource
		return nil
	}

	op, err := r.CreateOperation("Updating resource", updated, nil, run, nil)
	if err != nil {
		return rest.SmartError(err)
	}

	return rest.OperationResponse(op)
}

func resourceDelete(r *rest.Request) rest.Response {
	id := mux.Vars(r.HTTPRequest)["id"]
	if len(id) == 0 {
		return rest.NotFoundError("Resource")
	}

	_, ok := resources[id]
	if !ok {
		return rest.NotFoundError("Resource")
	}

	deleted := map[string][]string{}
	deleted["resources"] = []string{id}

	run := func(op *rest.Operation) error {
		delete(resources, id)
		return nil
	}

	op, err := r.CreateOperation("Deleting resource", deleted, nil, run, nil)
	if err != nil {
		return rest.SmartError(err)
	}

	return rest.OperationResponse(op)
}
