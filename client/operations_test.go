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
	"encoding/json"
	"time"

	check "gopkg.in/check.v1"

	"github.com/greenbrew/rest/api"
)

type operationsSuite struct {
}

var _ = check.Suite(&operationsSuite{})

func (s *operationsSuite) TestListOperationUUIDs(c *check.C) {
	expected := []string{"uuid1", "uuid2"}

	b, err := json.Marshal(expected)
	c.Assert(err, check.IsNil)

	mock := MockClient{
		Response: &api.Response{
			Metadata: b,
		},
	}

	opsCli := UpgradeToOperationsClient(&mock)
	uuids, err := opsCli.ListOperationUUIDs()
	c.Assert(err, check.IsNil)
	c.Assert(uuids, check.DeepEquals, expected)
}

func (s *operationsSuite) TestListOperations(c *check.C) {
	op1 := api.Operation{ID: "op1", Description: "opdesc", Status: "running", StatusCode: api.Pending}
	op2 := api.Operation{ID: "op2", Description: "opdesc", Status: "running", StatusCode: api.Pending}

	metadata := map[string][]api.Operation{
		"op1": {op1},
		"op2": {op2},
	}

	b, err := json.Marshal(metadata)
	c.Assert(err, check.IsNil)

	mock := MockClient{
		Response: &api.Response{
			Metadata: b,
		},
	}

	opsCli := UpgradeToOperationsClient(&mock)
	ops, err := opsCli.ListOperations()
	c.Assert(err, check.IsNil)
	c.Assert(ops, check.HasLen, 2)
	for _, op := range ops {
		switch op.ID {
		case "op1":
			c.Assert(op, check.DeepEquals, op1)
		case "op2":
			c.Assert(op, check.DeepEquals, op2)
		}
	}
}

func (s *operationsSuite) TestRetrieveOperationByID(c *check.C) {
	expected := &api.Operation{ID: "op1", Description: "opdesc", Status: "running", StatusCode: api.Pending}

	b, err := json.Marshal(expected)
	c.Assert(err, check.IsNil)

	mock := MockClient{
		Response: &api.Response{
			Metadata: b,
		},
	}

	opsCli := UpgradeToOperationsClient(&mock)
	op, etag, err := opsCli.RetrieveOperationByID("op1")
	c.Assert(err, check.IsNil)
	c.Assert(etag, check.Equals, "")
	c.Assert(op, check.DeepEquals, expected)
}

func (s *operationsSuite) TestWaitOperationToFinish(c *check.C) {
	expected := &api.Operation{ID: "op1", Description: "opdesc", Status: "running", StatusCode: api.Pending}

	b, err := json.Marshal(expected)
	c.Assert(err, check.IsNil)

	mock := MockClient{
		Response: &api.Response{
			Metadata: b,
		},
	}

	opsCli := UpgradeToOperationsClient(&mock)
	op, err := opsCli.WaitForOperationToFinish("op1", 30*time.Second)
	c.Assert(err, check.IsNil)
	c.Assert(op, check.DeepEquals, expected)
}

func (s *operationsSuite) TestDeleteOperation(c *check.C) {
	opsCli := UpgradeToOperationsClient(&MockClient{})
	err := opsCli.DeleteOperation("op1")
	c.Assert(err, check.IsNil)
}
