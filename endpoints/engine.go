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

package endpoints

import (
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/greenbrew/rest/logger"
)

const (
	stopTimeout = 120 * time.Second
)

type initFunc func() (interface{}, error)
type startFunc func(interface{}) error

// EndpointEngine represents the engine for a specific endpoint
type EndpointEngine interface {
	Start(wg *sync.WaitGroup) error
	Stop() error
}

type endpointEngine struct {
	endpoint Endpoint
	running  bool
	doneChan chan struct{}
}

// NewEndpointEngine returns a new instance of the endpoint
func NewEndpointEngine(ep Endpoint) EndpointEngine {
	return &endpointEngine{
		endpoint: ep,
		running:  false,
		doneChan: make(chan struct{}),
	}
}

func (e *endpointEngine) Start(wg *sync.WaitGroup) error {
	defer close(e.doneChan)

	e.running = false
	defer func() {
		if !e.running && wg != nil {
			wg.Done()
		}
		e.running = false
	}()

	data, err := e.endpoint.Init()
	if err != nil {
		return err
	}

	e.running = true
	// It is needed to set the 'done' signal after the listener is created and
	// before serving (which is blocking). At that point, even if a request
	// is received before the serve method has finished, it is buffered in the
	// listener until it is possible to be processed. So, no request would be missed
	if wg != nil {
		wg.Done()
	}

	sync := make(chan error)
	go func() {
		sync <- e.endpoint.Start(data)
	}()
	err = <-sync

	// As explained in https://golang.org/pkg/net/http/
	// ErrServerClosed is returned by the Server's Serve, ServeTLS, ListenAndServe, and
	// ListenAndServeTLS methods after a call to Shutdown or Close.
	// It was introduced to ensure that server was indeed closed, so let's not consider
	// is as an error
	if err != nil && err != http.ErrServerClosed {
		return errors.Errorf("Could not start %v endpoint: %v", e.endpoint.Name(), err)
	}
	return nil
}

func (e *endpointEngine) Stop() error {
	if !e.running {
		return nil
	}

	// Let's create a timeout, to force detention if endpoint fails to stop (unlikely)
	ticker := time.NewTicker(stopTimeout)
	defer ticker.Stop()

	// We need to ensure that the endpoint has been stopped. The
	// only way it happens is when doneChan is closed.
	// However here could happen a race if the e.endpoint.Stop() is called
	// before the server is actually started. In that case, e.endpoint.Stop()
	// returns no error but the doneChan is not closed and never will be
	// because the start is happening after the stop. To prevent it
	// we retry the stop until doneChan is eventually closed.
	for {
		err := e.endpoint.Stop()

		// Give some time for endpoint to finish the server.
		// No matter if it is not enough time, because the stop will be retried
		// in such case. This is added, just to prevent that the first time
		// it is not called e.endpoint.Stop() two times always
		time.Sleep(50 * time.Millisecond)

		select {
		case <-e.doneChan:
			logger.Infof("Stop listening %v endpoint", e.endpoint.Name())
			return err
		case <-ticker.C:
			logger.Critf("Could not stop %v endpoint into expected time. Forcing stop service", e.endpoint.Name())
		default:
			time.Sleep(time.Second)
		}
	}
}
