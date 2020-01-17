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
	"net/url"
	"time"

	"github.com/greenbrew/rest/api"
)

type operations struct {
	Client
}

// UpgradeToOperationsClient wraps generic client to implement Operations interface operations
func UpgradeToOperationsClient(c Client) Operations {
	return &operations{c}
}

// ListOperationUUIDs returns a list of operation uuids
func (c *operations) ListOperationUUIDs() ([]string, error) {
	urls := []string{}
	resource := APIPath("operations")
	_, err := c.QueryStruct("GET", resource, nil, nil, nil, "", &urls)
	return urls, err
}

// ListOperations returns a list of Operation struct
func (c *operations) ListOperations() ([]api.Operation, error) {
	apiOps := map[string][]api.Operation{}

	params := QueryParams{
		"recursion": "1",
	}
	resource := APIPath("operations")
	_, err := c.QueryStruct("GET", resource, params, nil, nil, "", &apiOps)

	// Turn it into just a list of operations
	ops := []api.Operation{}
	for _, v := range apiOps {
		for _, op := range v {
			ops = append(ops, op)
		}
	}

	return ops, err
}

// RetrieveOperationByID returns a websocket connection for the provided operation id
func (c *operations) RetrieveOperationByID(uuid string) (*api.Operation, string, error) {
	op := &api.Operation{}
	resource := APIPath("operations", url.QueryEscape(uuid))
	etag, err := c.QueryStruct("GET", resource, nil, nil, nil, "", op)
	return op, etag, err
}

// WaitForOperationToFinish blocks until operation is finished or timeout
func (c *operations) WaitForOperationToFinish(uuid string, timeout time.Duration) (*api.Operation, error) {
	op := &api.Operation{}
	resource := APIPath("operations", url.QueryEscape(uuid), "wait")
	var params QueryParams
	if timeout > 0 {
		params := make(map[string]string)
		params["timeout"] = timeout.String()
	}
	_, err := c.QueryStruct("GET", resource, params, nil, nil, "", op)
	return op, err
}

// DeleteOperation deletes (cancels) a running operation
func (c *operations) DeleteOperation(uuid string) error {
	resource := APIPath("operations", url.QueryEscape(uuid))
	_, _, err := c.CallAPI("DELETE", resource, nil, nil, nil, "")
	return err
}
