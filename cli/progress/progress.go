// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2014-2015 Canonical Ltd
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

package progress

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

// Meter is an interface to show progress to the user
type Meter interface {
	// Start progress with max "total" steps
	Start(label string, total float64)

	// set progress to the "current" step
	Set(current float64)

	// set "total" steps needed
	SetTotal(total float64)

	// Finish the progress display
	Finished()

	// Indicate indefinite activity by showing a spinner
	Spin(msg string)

	// interface for writer
	Write(p []byte) (n int, err error)

	// notify the user of miscellaneous events
	Notify(string)
}

// NullMeter is a Meter that does nothing
type NullMeter struct{}

// Null is a default NullMeter instance
var Null = NullMeter{}

// Start progress with max "total" steps
func (NullMeter) Start(string, float64) {}

// Set sets progress to the "current" step
func (NullMeter) Set(float64) {}

// SetTotal sets "total" steps needed
func (NullMeter) SetTotal(float64) {}

// Finished finishes the progress display
func (NullMeter) Finished() {}

// Finished finishes the progress display
func (NullMeter) Write(p []byte) (int, error) { return len(p), nil }

// Notify notifies the user of miscellaneous events
func (NullMeter) Notify(string) {}

// Spin indicates indefinite activity by showing a spinner
func (NullMeter) Spin(msg string) {}

// QuietMeter is a Meter that _just_ shows Notify()s.
type QuietMeter struct{ NullMeter }

// Notify notifies the user of miscellaneous events
func (QuietMeter) Notify(msg string) {
	fmt.Fprintln(stdout, msg)
}

// testMeter, if set, is returned by MakeProgressBar; set it from tests.
var testMeter Meter

// MockMeter mocks the meter instance for testing purposes
func MockMeter(meter Meter) func() {
	testMeter = meter
	return func() {
		testMeter = nil
	}
}

var inTesting = len(os.Args) > 0 && strings.HasSuffix(os.Args[0], ".test") || os.Getenv("SPREAD_SYSTEM") != ""

// MakeProgressBar creates an appropriate progress.Meter for the environ in
// which it is called:
//
// * if MockMeter has been called, return that.
// * if no terminal is attached, or we think we're running a test, a
//   minimalistic QuietMeter is returned.
// * otherwise, an ANSIMeter is returned.
//
// TODO: instead of making the pivot at creation time, do it at every call.
func MakeProgressBar() Meter {
	if testMeter != nil {
		return testMeter
	}
	if !inTesting && terminal.IsTerminal(int(os.Stdin.Fd())) {
		return &ANSIMeter{}
	}

	return QuietMeter{}
}
