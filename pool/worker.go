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

// Worker represents the process executing a job
type Worker struct {
	workerPool chan *Worker
	jobChannel chan Job
	stopChan   chan struct{}
	doneChan   chan struct{}
	running    bool
}

// NewWorker returns a new worker instance
func NewWorker(pool chan *Worker) *Worker {
	return &Worker{
		workerPool: pool,
		jobChannel: make(chan Job),
	}
}

// Start method starts the run loop for this worker
func (w *Worker) Start() {
	if w.running {
		return
	}

	w.stopChan = make(chan struct{})
	w.doneChan = make(chan struct{})

	go func() {
		defer close(w.doneChan)
		defer func() { w.running = false }()

		w.running = true

		var job Job
		for {
			// At this point the worker is free. Add it to the pool.
			// The pool will send a job to its jobChannel when the worker
			// is the next in the pool to handle a job
			w.workerPool <- w

			select {
			case job = <-w.jobChannel:
				// we have received a work request.
				job()
			case <-w.stopChan:
				// we have received a signal to stop
				return
			}
		}
	}()
}

// Stop signals the worker to stop listening for work requests and will
// block until the worker is stopped.
func (w *Worker) Stop() {
	if !w.running {
		return
	}

	close(w.stopChan)
	<-w.doneChan
}
