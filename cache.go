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
	"sync"

	"github.com/pkg/errors"
)

type cache struct {
	operations map[string]*Operation
	mux        sync.Mutex
}

func (c *cache) getOperationsMap() map[string]*Operation {
	c.mux.Lock()
	defer c.mux.Unlock()

	return c.operations
}

func (c *cache) addOperation(op *Operation) {
	c.mux.Lock()
	defer c.mux.Unlock()

	if c.operations == nil {
		c.operations = make(map[string]*Operation)
	}
	c.operations[op.id] = op
}

func (c *cache) getOperationByID(id string) (*Operation, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	op, ok := c.operations[id]
	if !ok {
		return nil, errors.Errorf("Operation '%s' does not exist", id)
	}
	return op, nil
}

func (c *cache) deleleOperationByID(id string) {
	c.mux.Lock()
	defer c.mux.Unlock()

	_, ok := c.operations[id]
	if !ok {
		return
	}

	delete(c.operations, id)
}
