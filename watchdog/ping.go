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

package watchdog

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/greenbrew/rest/client"
	"github.com/greenbrew/rest/system"
)

// Pinger method to ping the REST service
type Pinger interface {
	Ping() bool
}

type pinger struct {
	c client.Client
}

// DefaultPinger returns a default implementation for Pinger interface using
// a REST client pointing to a local unix socket
func DefaultPinger(unixSocketPath string) (Pinger, error) {
	if len(unixSocketPath) == 0 {
		return nil, errors.New("Empty unix socket path provided")
	}

	if !system.PathExists(unixSocketPath) {
		return nil, errors.New("Unix socket path does not exist")
	}

	c, err := client.New(unixSocketPath, nil)
	if err != nil {
		return nil, fmt.Errorf("Could not start watchdog client: %v", err)
	}

	return &pinger{c}, nil
}

func (p *pinger) Ping() bool {
	_, _, err := p.c.CallAPI("GET", client.APIPath("version"), nil, nil, nil, "")
	// The error is not relevant, only the fact that the endpoint
	// is reachable or not
	return err == nil
}
