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
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/coreos/go-systemd/daemon"
)

// Watchdog manager for the systemd wathdog feature
type Watchdog struct {
	keepAliveInterval time.Duration
	pinger            Pinger

	closeChan chan struct{}
	doneChan  chan struct{}

	running bool
}

// New returns a new instance of a watchdog
func New(keepAliveInterval time.Duration, pinger Pinger) *Watchdog {
	return &Watchdog{
		keepAliveInterval: keepAliveInterval,
		pinger:            pinger,
		closeChan:         make(chan struct{}),
		doneChan:          make(chan struct{}),
	}
}

// Run starts watchdogging
func (w *Watchdog) Run(ctx context.Context) error {
	defer close(w.doneChan)

	w.running = true
	defer func() { w.running = false }()

	if !w.running {
		w.running = true
	}

	if w.keepAliveInterval <= 0 {
		return errors.New("Not a valid keep-alive interval for the watchdog")
	}

	ticker := time.NewTicker(w.keepAliveInterval)

	// Notify if this is started in a systemd service of notify type.
	// If the service is not run through systemd, this line is a no-op
	daemon.SdNotify(false, "READY=1")

	// The service sends watchdog keep-alive at regular interval.
	// If it fails to do so, systemd will restart it
	for {
		select {
		case <-ticker.C:
			if !w.pinger.Ping() {
				ticker.Stop()
				return nil
			}
			daemon.SdNotify(false, "WATCHDOG=1")
		case <-w.closeChan:
			return nil
		}
	}
}

// Stop stops watchdogging
func (w *Watchdog) Stop() error {
	if !w.running {
		return nil
	}

	daemon.SdNotify(false, "STOPPING=1")

	close(w.closeChan)
	<-w.doneChan
	return nil
}
