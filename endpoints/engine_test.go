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
	"math/rand"
	"net/http"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	check "gopkg.in/check.v1"
)

var tCtx *testing.T

func Test(t *testing.T) {
	tCtx = t
	check.TestingT(t)
}

type engineSuite struct {
	mockEndpoint *MockEndpoint
	mockCtrl     *gomock.Controller
}

var _ = check.Suite(&engineSuite{})

func (s *engineSuite) SetUpTest(c *check.C) {
	s.mockCtrl = gomock.NewController(tCtx)
	s.mockEndpoint = NewMockEndpoint(s.mockCtrl)
}

func (s *engineSuite) TearDownTest(c *check.C) {
	if s.mockCtrl != nil {
		s.mockCtrl.Finish()
	}
}

func (s *engineSuite) TestEndpointFinishInmediatly(c *check.C) {
	s.mockEndpoint.EXPECT().Init()
	s.mockEndpoint.EXPECT().Start(gomock.Any())

	ee := NewEndpointEngine(s.mockEndpoint)

	// endpoint won't block the start, so it is not needed a goroutine
	err := ee.Start(nil)
	c.Assert(err, check.IsNil)

	// test that stop called in an already stopped endpoint does not fail
	err = ee.Stop()
	c.Assert(err, check.IsNil)
}

func (s *engineSuite) TestLaunchAndStop(c *check.C) {
	runningCh := make(chan struct{})
	readyCh := make(chan struct{})

	s.mockEndpoint.EXPECT().Init()
	s.mockEndpoint.EXPECT().Start(gomock.Any()).Do(func(data interface{}) {
		close(readyCh)
		<-runningCh
	})
	s.mockEndpoint.EXPECT().Stop().Do(func() {
		close(runningCh)
	})
	s.mockEndpoint.EXPECT().Name()

	// Create the engine and start the endpoint in another routine, because it blocks
	ee := NewEndpointEngine(s.mockEndpoint)
	go func() {
		err := ee.Start(nil)
		c.Assert(err, check.IsNil)
	}()

	// wait for endpoint to start before stopping it
	<-readyCh
	err := ee.Stop()
	c.Assert(err, check.IsNil)
}

func (s *engineSuite) TestStartupData(c *check.C) {
	runningCh := make(chan struct{})
	readyCh := make(chan struct{})

	data := struct {
		whatever string
		number   int
	}{
		whatever: "whatever",
		number:   42,
	}

	s.mockEndpoint.EXPECT().Init().DoAndReturn(func() (interface{}, error) {
		return data, nil
	})
	s.mockEndpoint.EXPECT().Start(data).Do(func(data interface{}) {
		close(readyCh)
		<-runningCh
	})
	s.mockEndpoint.EXPECT().Stop().Do(func() {
		close(runningCh)
	})
	s.mockEndpoint.EXPECT().Name()

	// Create the engine and start the endpoint in another routine, because it blocks
	ee := NewEndpointEngine(s.mockEndpoint)
	go func() {
		err := ee.Start(nil)
		c.Assert(err, check.IsNil)
	}()

	// wait for endpoint to start before stopping it
	<-readyCh
	err := ee.Stop()
	c.Assert(err, check.IsNil)
}

func (s *engineSuite) TestInitFails(c *check.C) {
	s.mockEndpoint.EXPECT().Init().DoAndReturn(func() (interface{}, error) {
		return nil, errors.New("whatever")
	})

	// Create the engine and start the endpoint in another routine, because it blocks
	ee := NewEndpointEngine(s.mockEndpoint)
	err := ee.Start(nil)
	c.Assert(err, check.NotNil)
}

func (s *engineSuite) TestStartFails(c *check.C) {
	s.mockEndpoint.EXPECT().Init()
	s.mockEndpoint.EXPECT().Name()
	s.mockEndpoint.EXPECT().Start(gomock.Any()).Return(errors.New("whatever"))

	// Create the engine and start the endpoint in another routine, because it blocks
	ee := NewEndpointEngine(s.mockEndpoint)
	err := ee.Start(nil)
	c.Assert(err, check.NotNil)
}

func (s *engineSuite) TestErrServerClosedNotConsideredAsError(c *check.C) {
	s.mockEndpoint.EXPECT().Init()
	s.mockEndpoint.EXPECT().Start(gomock.Any()).Return(http.ErrServerClosed)

	// Create the engine and start the endpoint in another routine, because it blocks
	ee := NewEndpointEngine(s.mockEndpoint)
	err := ee.Start(nil)
	c.Assert(err, check.IsNil)
}

func (s *engineSuite) TestLaunchAndStopMultipleEndpoints(c *check.C) {
	// laucn a random number of endpoints between 1 and 50
	n := rand.Intn(50) + 1

	type testEndpoint struct {
		runningCh chan struct{}
		readyCh   chan struct{}
		mock      *MockEndpoint
	}

	engines := []EndpointEngine{}
	endpoints := []*testEndpoint{}

	wg := &sync.WaitGroup{}

	for i := 0; i < n; i++ {
		e := &testEndpoint{
			runningCh: make(chan struct{}),
			readyCh:   make(chan struct{}),
			mock:      NewMockEndpoint(s.mockCtrl),
		}

		e.mock.EXPECT().Init()
		e.mock.EXPECT().Start(gomock.Any()).Do(func(data interface{}) {
			close(e.readyCh)
			<-e.runningCh
		})
		e.mock.EXPECT().Stop().Do(func() {
			close(e.runningCh)
		})
		e.mock.EXPECT().Name()

		ee := NewEndpointEngine(e.mock)
		go func() {
			wg.Add(1)
			err := ee.Start(wg)
			c.Assert(err, check.IsNil)
		}()

		engines = append(engines, ee)
		endpoints = append(endpoints, e)
	}

	wg.Wait()

	for i := range engines {
		<-endpoints[i].readyCh
		err := engines[i].Stop()
		c.Assert(err, check.IsNil)
	}
}
