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
	"errors"
	"time"
)

const (
	maxConcurrentJobsDefault = 10
	maxQueuedJobsDefault     = 100
)

// DispatcherStatus type representing the current status of the dispatcher
type DispatcherStatus int

// Different status values for the dispatcher
const (
	DispatcherStatusStopped = iota
	DispatcherStatusRunning
)

const (
	statusChangeTimeout   = 5 * time.Second
	statusChangeCheckStep = 100 * time.Millisecond
)

var errDispatcherStatusTimeout = errors.New("Timeout waiting for dispatcher status change")

// Dispatcher is the pool engine. Holds the queue receiving the jobs
// to be processed and holds the queue of available workers.
// When a job is received, takes the next available worker job channel
// and sends the job to it. Then that channel is read and processed
// by the corresponding worker
type Dispatcher struct {
	Queue *JobChannel
	// A pool of workers that are registered with the dispatcher
	pool      chan *Worker
	queueSize int
	poolSize  int
	running   bool

	stopChan chan bool
	doneChan chan struct{}
}

// NewDispatcher returns a new dispatcher with a new workers pool of the requested size
func NewDispatcher(queueSize, poolSize int) *Dispatcher {
	if queueSize <= 0 {
		queueSize = maxQueuedJobsDefault
	}
	if poolSize <= 0 {
		poolSize = maxConcurrentJobsDefault
	}

	return &Dispatcher{
		queueSize: queueSize,
		poolSize:  poolSize,
		running:   false,
	}
}

// Start creates and starts all the workers until filling the pool
func (d *Dispatcher) Start() {
	if d.running {
		return
	}

	// Build queue only if it nil. Otherwise we are reusing the same queue
	if d.Queue == nil || d.Queue.closed {
		d.Queue = newJobChannel(d.queueSize)
	}
	d.pool = make(chan *Worker, d.poolSize)

	// Initialize our channels as they supposed to be closed at this time
	d.stopChan = make(chan bool)
	d.doneChan = make(chan struct{})

	// starting as much workers as size allows
	for i := 0; i < cap(d.pool); i++ {
		worker := NewWorker(d.pool)
		worker.Start()
	}

	go d.run()

	// Wait until reached the capacity to update the status
	for len(d.pool) < cap(d.pool) {
		time.Sleep(time.Millisecond)
	}

	d.running = true
}

// Stop stops the dispatcher
func (d *Dispatcher) Stop(closeQueue bool) {
	if !d.running {
		return
	}

	// Send close request signal to the stop channel. This
	// is the signal to stop the dispatcher and to close the
	// queue (if true) at the same time
	d.stopChan <- closeQueue
	<-d.doneChan

	d.finalize()
}

func (d *Dispatcher) finalize() {
	// dispatch remaining jobs after closing the input job
	if d.Queue.closed {
		for job := range d.Queue.queue {
			d.dispatch(job)
		}
	}

	// Finish all workers and close the pool
	d.foreachWorker(func(w *Worker) {
		w.Stop()
	})
}

func (d *Dispatcher) run() {
	defer func() {
		d.running = false
		close(d.doneChan)
	}()

	for {
		select {
		case job := <-d.Queue.queue:
			// a job request has been received
			go d.dispatch(job)
		case mustClose := <-d.stopChan:
			if mustClose {
				// Close jobQueue to stop receiving more jobs
				d.Queue.Close()
			}
			return
		}
	}
}

func (d *Dispatcher) dispatch(job Job) {
	// try to obtain a worker job channel that is available.
	// this will block until a worker is idle
	nextWorker := <-d.pool
	// dispatch the job to the worker job channel
	nextWorker.jobChannel <- job
}

func (d *Dispatcher) foreachWorker(f func(w *Worker)) {
	for i := 0; i < cap(d.pool); i++ {
		f(<-d.pool)
	}
}
