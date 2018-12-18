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
	"encoding/json"
	"net/http"
	"path/filepath"
	"time"

	"github.com/pborman/uuid"
	"github.com/pkg/errors"
	"github.com/greenbrew/rest/api"
	check "gopkg.in/check.v1"
)

type responseOperationSuite struct {
	err error
}

var _ = check.Suite(&responseOperationSuite{})

func (s *responseOperationSuite) TestNotPendingOperationCannotBeStarted(c *check.C) {
	op := &Operation{
		status: api.Running,
	}

	response := OperationResponse(op)

	w := newBufferedResponseWriter()
	err := response.Render(w)
	c.Assert(err, check.ErrorMatches, "Only pending operations can be started")
}

func (s *responseOperationSuite) TestOperationResponse(c *check.C) {
	id := uuid.NewRandom().String()
	url := filepath.Join("1.0/operations", id)
	createdAt := time.Now()
	updatedAt := createdAt
	resources := map[string][]string{
		"whatever": []string{"my-whatever-id"},
	}

	run := func(*Operation) error {
		// run enough time to detect intermediate states into the test
		time.Sleep(time.Millisecond * 500)
		return nil
	}

	op := &Operation{
		id:          id,
		description: "foo operation",
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		status:      api.Pending,
		url:         url,
		resources:   resources,
		metadata:    nil,
		onRun:       run,
		doneCh:      make(chan error),
		cache:       &cache{operations: make(map[string]*Operation)},
		events:      &eventsManager{listeners: make(map[string]*eventsListener)},
	}

	// Setup the resource URLs
	formattedResources := make(map[string][]string)
	for k, v := range op.resources {
		var vals []string
		for _, c := range v {
			vals = append(vals, api.Path(k, c))
		}
		formattedResources[k] = vals
	}

	// Create the expected body after op run
	expectedStruct := &api.ResponseRaw{
		Response: api.Response{
			Type:       api.ResponseTypeAsync,
			Status:     api.Created.String(),
			StatusCode: int(api.Created),
			Operation:  url,
		},
		Metadata: api.Operation{
			ID:          op.id,
			Description: op.description,
			CreatedAt:   op.createdAt,
			UpdatedAt:   op.updatedAt,
			Status:      api.Running.String(),
			StatusCode:  api.Running,
			Resources:   formattedResources,
			Metadata:    op.metadata,
			Err:         op.errStr,
		},
	}
	expectedBytes, err := json.Marshal(expectedStruct)
	c.Assert(err, check.IsNil)

	response := OperationResponse(op)

	w := newBufferedResponseWriter()
	err = response.Render(w)
	c.Assert(err, check.IsNil)

	// Verify body content:
	// Compare the content of the buffered expected and desired
	var expected interface{}
	err = json.Unmarshal(expectedBytes, &expected)
	c.Assert(err, check.IsNil)

	var result interface{}
	err = json.Unmarshal(w.buffer.Bytes(), &result)
	c.Assert(err, check.IsNil)

	c.Assert(expected, check.DeepEquals, result)

	// Verify status code
	c.Assert(w.statusCode, check.Equals, http.StatusAccepted)

	// Verify Location header
	location, ok := w.headers["Location"]
	c.Assert(ok, check.Equals, true)
	c.Assert(location, check.HasLen, 1)
	c.Assert(location[0], check.Equals, url)

	// Wait for final state
	err = op.WaitFinal(10)
	c.Assert(err, check.IsNil)

	// Verify that operation status after finilizing operation is updated
	_, gotOp, err := op.Render()
	c.Assert(err, check.IsNil)

	expectedOp := &api.Operation{
		ID:          op.id,
		Description: op.description,
		CreatedAt:   op.createdAt,
		UpdatedAt:   op.updatedAt,
		Status:      api.Success.String(),
		StatusCode:  api.Success,
		Resources:   formattedResources,
		Metadata:    op.metadata,
		Err:         op.errStr,
	}

	c.Assert(gotOp, check.DeepEquals, expectedOp)
}

func (s *responseOperationSuite) TestOnRunFails(c *check.C) {
	id := uuid.NewRandom().String()
	url := filepath.Join("1.0/operations", id)
	createdAt := time.Now()
	updatedAt := createdAt
	resources := map[string][]string{
		"whatever": []string{"my-whatever-id"},
	}

	run := func(*Operation) error {
		return errors.New("Runtime error")
	}

	op := &Operation{
		id:          id,
		description: "foo operation",
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		status:      api.Pending,
		url:         url,
		resources:   resources,
		metadata:    nil,
		onRun:       run,
		doneCh:      make(chan error),
		cache:       &cache{operations: make(map[string]*Operation)},
		events:      &eventsManager{listeners: make(map[string]*eventsListener)},
	}

	response := OperationResponse(op)

	w := newBufferedResponseWriter()
	err := response.Render(w)
	c.Assert(err, check.IsNil)

	op.WaitFinal(10)
	_, operation, err := op.Render()
	c.Assert(err, check.IsNil)
	c.Assert(operation.StatusCode, check.Equals, api.Failure)
}

func (s *responseOperationSuite) TestCancel(c *check.C) {
	id := uuid.NewRandom().String()
	url := filepath.Join("1.0/operations", id)
	createdAt := time.Now()
	updatedAt := createdAt
	resources := map[string][]string{
		"whatever": []string{"my-whatever-id"},
	}

	run := func(*Operation) error {
		// run enough time for making possible to cancel
		time.Sleep(time.Millisecond * 500)
		return nil
	}

	cancelCh := make(chan struct{})
	cancel := func(*Operation) error {
		close(cancelCh)
		return nil
	}

	op := &Operation{
		id:          id,
		description: "foo operation",
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		status:      api.Pending,
		url:         url,
		resources:   resources,
		metadata:    nil,
		onRun:       run,
		onCancel:    cancel,
		doneCh:      make(chan error),
		cache:       &cache{operations: make(map[string]*Operation)},
		events:      &eventsManager{listeners: make(map[string]*eventsListener)},
	}

	response := OperationResponse(op)

	w := newBufferedResponseWriter()
	err := response.Render(w)
	c.Assert(err, check.IsNil)

	err = op.Cancel()
	c.Assert(err, check.IsNil)

	err = op.WaitFinal(10)
	c.Assert(err, check.IsNil)

	// Verify that cancel operation has been called
	<-cancelCh

	_, operation, err := op.Render()
	c.Assert(operation.StatusCode, check.Equals, api.Cancelled)
}

func (s *responseOperationSuite) TestCancelOperationWithoutCancelHandler(c *check.C) {
	id := uuid.NewRandom().String()
	url := filepath.Join("1.0/operations", id)
	createdAt := time.Now()
	updatedAt := createdAt
	resources := map[string][]string{
		"whatever": []string{"my-whatever-id"},
	}

	run := func(*Operation) error {
		// run enough time for making possible to cancel
		time.Sleep(time.Millisecond * 500)
		return nil
	}

	op := &Operation{
		id:          id,
		description: "foo operation",
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		status:      api.Pending,
		url:         url,
		resources:   resources,
		metadata:    nil,
		onRun:       run,
		onCancel:    nil,
		doneCh:      make(chan error),
		cache:       &cache{operations: make(map[string]*Operation)},
		events:      &eventsManager{listeners: make(map[string]*eventsListener)},
	}

	response := OperationResponse(op)

	w := newBufferedResponseWriter()
	err := response.Render(w)
	c.Assert(err, check.IsNil)

	err = op.Cancel()
	c.Assert(err, check.IsNil)

	err = op.WaitFinal(10)
	c.Assert(err, check.IsNil)

	_, operation, err := op.Render()
	c.Assert(operation.StatusCode, check.Equals, api.Cancelled)
}
