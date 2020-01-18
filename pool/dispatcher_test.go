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

package pool

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	check "gopkg.in/check.v1"
)

func Test(t *testing.T) { check.TestingT(t) }

type dispatcherSuite struct{}

var _ = check.Suite(&dispatcherSuite{})

func (s *dispatcherSuite) TestSizes(c *check.C) {
	queueSize := 50
	poolSize := 5

	d := NewDispatcher(queueSize, poolSize)

	c.Assert(len(d.pool), check.Equals, 0)
	c.Assert(cap(d.pool), check.Equals, 0)

	d.Start()
	defer d.Stop(true)

	c.Assert(len(d.Queue.queue), check.Equals, 0)
	c.Assert(cap(d.Queue.queue), check.Equals, queueSize)
	c.Assert(len(d.pool), check.Equals, poolSize)
	c.Assert(cap(d.pool), check.Equals, poolSize)
}

func (s *dispatcherSuite) TestCorrectJobsOrder(c *check.C) {
	nJobs := 1500
	queueSize := 5
	poolSize := 10

	d := NewDispatcher(queueSize, poolSize)
	d.Start()
	defer d.Stop(true)

	executionsRegister := make(chan uint32, nJobs)

	initialJobNumber := uint32(rand.Intn(100))
	jobNumber := initialJobNumber

	var wg sync.WaitGroup
	var lock sync.Mutex
	for i := 0; i < nJobs; i++ {
		lock.Lock()
		wg.Add(1)

		err := d.Queue.Push(func() {
			executionsRegister <- jobNumber
			atomic.AddUint32(&jobNumber, 1)

			lock.Unlock()
			wg.Done()
		})
		c.Assert(err, check.IsNil)
	}

	wg.Wait()

	close(executionsRegister)

	i := initialJobNumber
	for e := range executionsRegister {
		c.Assert(e, check.DeepEquals, i)
		i++
	}
}

// Test that new jobs added when dispatcher is stopped will result
// in an error
func (s *dispatcherSuite) TestCannotPushWhenStopped(c *check.C) {
	nJobs := 150
	queueSize := 5
	poolSize := 3

	d := NewDispatcher(queueSize, poolSize)
	d.Start()

	// Add jobs to the queue and wait for them to be pushed before stopping
	initialJobNumber := uint32(rand.Intn(100))
	jobNumber := initialJobNumber

	var wg sync.WaitGroup
	var lock sync.Mutex
	for i := 0; i < nJobs; i++ {
		lock.Lock()
		wg.Add(1)

		err := d.Queue.Push(func() {
			atomic.AddUint32(&jobNumber, 1)

			lock.Unlock()
			wg.Done()
		})
		c.Assert(err, check.IsNil)
	}

	wg.Wait()
	d.Stop(true)

	// Check that any try to add a new job returns an error
	err := d.Queue.Push(func() { atomic.AddUint32(&jobNumber, 1) })
	c.Assert(err, check.Equals, ErrJobQueueClosed)
}

func (s *dispatcherSuite) TestCannotPushWhenQueueIsFull(c *check.C) {
	queueSize := 1
	poolSize := 1

	d := NewDispatcher(queueSize, poolSize)
	d.Start()
	defer d.Stop(true)

	for i := 0; i < poolSize+1; i++ {
		err := d.Queue.Push(func() {})
		c.Assert(err, check.IsNil)
	}

	// A third job will received a queue full error
	c.Assert(d.Queue.Push(func() {}), check.Equals, ErrJobQueueFull)
}

func (s *dispatcherSuite) TestUnattendedJobsAfterClosing(c *check.C) {
	queueSize := 5
	poolSize := 1
	nJobs := 4

	d := NewDispatcher(queueSize, poolSize)
	d.Start()

	// Let's add nJobs blocked during execution until the end of the tests
	var wg sync.WaitGroup
	for i := 0; i < nJobs; i++ {
		wg.Add(1)
		err := d.Queue.Push(func() {
			wg.Wait()
			time.Sleep(time.Second)
		})
		c.Assert(err, check.IsNil)
	}

	d.Stop(false)

	// as first job is waiting forever while attended by unique worker,
	// then, there should be in the queue the remaining ones
	c.Assert(len(d.Queue.queue), check.Equals, nJobs-1)

	for i := 0; i < nJobs; i++ {
		wg.Done()
	}
}

func (s *dispatcherSuite) TestDispatcherCanBeRestarted(c *check.C) {
	firstCountOfJobs := 7
	secondCountOfJobs := 8

	queueSize := 15
	poolSize := 10

	d := NewDispatcher(queueSize, poolSize)
	d.Start()

	var wg sync.WaitGroup
	for i := 0; i < firstCountOfJobs; i++ {
		wg.Add(1)
		d.Queue.Push(func() {
			wg.Done()
		})
	}

	wg.Wait()

	// Stop the dispatcher, not deleting the input queue
	d.Stop(false)

	for i := 0; i < secondCountOfJobs; i++ {
		wg.Add(1)
		d.Queue.Push(func() {
			wg.Done()
		})
	}

	// Verify last added jobs are not attended
	c.Assert(d.Queue.queue, check.HasLen, secondCountOfJobs)
	c.Assert(len(d.pool), check.Equals, 0)

	// Restart dispatcher
	d.Start()
	c.Assert(len(d.pool), check.Equals, poolSize)

	// Wait for pool to process the remaining ones
	wg.Wait()
	// Verify no jobs remain in queue
	c.Assert(d.Queue.queue, check.HasLen, 0)
}
