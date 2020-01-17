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
	"sync"

	"github.com/pkg/errors"
)

var (
	// ErrJobQueueFull happens when jobs queue reaches its maximum size and a new job
	// needs to be attended
	ErrJobQueueFull = errors.New("Jobs queue is full")
	// ErrJobQueueClosed happens when job queue gets closed and a new job arrives
	ErrJobQueueClosed = errors.New("Jobs queue already closed")
)

// Job the function representing the work to be processed by a worker
type Job func()

// JobChannel a channel to read or write jobs
type JobChannel struct {
	queue  chan Job
	closed bool
	mux    sync.Mutex
}

// builds a new JobChannel
func newJobChannel(size int) *JobChannel {
	return &JobChannel{
		queue:  make(chan Job, size),
		closed: false,
	}
}

// Push adds job to queue
func (c *JobChannel) Push(job Job) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.closed {
		return ErrJobQueueClosed
	}

	// If queue is full, default case returns an error
	select {
	case c.queue <- job:
	default:
		return ErrJobQueueFull
	}

	return nil
}

// Close closes the queue
func (c *JobChannel) Close() {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.closed = true
	close(c.queue)
}
