// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2017 Canonical Ltd
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

package progresstest

import (
	"github.com/greenbrew/rest/cli/progress"
)

// Meter implements the progress.Meter interface for testing purposes
type Meter struct {
	Labels   []string
	Totals   []float64
	Values   []float64
	Written  [][]byte
	Notices  []string
	Finishes int
}

// interface check
var _ progress.Meter = (*Meter)(nil)

// Start progress with max "total" steps
func (p *Meter) Start(label string, total float64) {
	p.Spin(label)
	p.SetTotal(total)
}

// Set sets progress to the "current" step
func (p *Meter) Set(value float64) {
	p.Values = append(p.Values, value)
}

// SetTotal sets "total" steps needed
func (p *Meter) SetTotal(total float64) {
	p.Totals = append(p.Totals, total)
}

// Finished finishes the progress display
func (p *Meter) Finished() {
	p.Finishes++
}

// Spin indicates indefinite activity by showing a spinner
func (p *Meter) Spin(label string) {
	p.Labels = append(p.Labels, label)
}

func (p *Meter) Write(bs []byte) (n int, err error) {
	p.Written = append(p.Written, bs)
	n = len(bs)

	return
}

// Notify notifies the user of miscellaneous events
func (p *Meter) Notify(msg string) {
	p.Notices = append(p.Notices, msg)
}
