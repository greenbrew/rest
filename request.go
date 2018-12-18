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
	"context"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/greenbrew/rest/api"
	"github.com/greenbrew/rest/logger"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
)

// Request represents the context of a REST request
type Request struct {
	HTTPRequest *http.Request
	daemon      *Daemon
	version     string
}

// CreateOperation creates an operation to be executed asynchronously
func (r *Request) CreateOperation(
	description string,
	opResources map[string][]string,
	opMetadata interface{},
	onRun func(*Operation) error,
	cancel context.CancelFunc) (*Operation, error) {

	// Main attributes
	op := &Operation{}
	op.id = uuid.NewRandom().String()
	op.description = description
	op.createdAt = time.Now()
	op.updatedAt = op.createdAt
	op.status = api.Pending
	op.url = filepath.Join(api.Version, "operations", op.id)
	op.resources = opResources
	op.doneCh = make(chan error)

	var err error
	op.metadata, err = parseMetadata(opMetadata)
	if err != nil {
		return nil, err
	}

	op.onRun = onRun
	op.cancel = cancel

	op.version = r.version

	if r.daemon.dispatcher != nil {
		op.operationsQueue = r.daemon.dispatcher.Queue
	}

	if r.daemon.cache == nil {
		return nil, errors.New("Cache not initialized")
	}
	op.cache = r.daemon.cache
	op.cache.addOperation(op)

	logger.Debugf("New operation: %s", op.id)
	_, md, _ := op.Render()

	op.events = r.daemon.events
	op.events.send(md)

	return op, nil
}

// IsRecursionRequest checks whether the given HTTP request is marked with the
// "recursion" flag in its form values.
func (r *Request) IsRecursionRequest() bool {
	if r.HTTPRequest == nil {
		return false
	}

	recursionStr := r.HTTPRequest.FormValue("recursion")
	recursion, err := strconv.Atoi(recursionStr)
	if err != nil {
		return false
	}

	return recursion != 0
}

// IsRecursionRequest checks whether the given HTTP request is marked with the
// "recursion" flag in its form values.
func IsRecursionRequest(r *http.Request) bool {
	recursionStr := r.FormValue("recursion")

	recursion, err := strconv.Atoi(recursionStr)
	if err != nil {
		return false
	}

	return recursion != 0
}
