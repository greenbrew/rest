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

package api

// StatusCode represents a valid REST operation status code
type StatusCode int

// status codes
const (
	Created    StatusCode = 100
	Started    StatusCode = 101
	Stopped    StatusCode = 102
	Running    StatusCode = 103
	Cancelling StatusCode = 104
	Pending    StatusCode = 105
	Starting   StatusCode = 106
	Stopping   StatusCode = 107
	Aborting   StatusCode = 108
	Freezing   StatusCode = 109
	Frozen     StatusCode = 110
	Thawed     StatusCode = 111
	Error      StatusCode = 112

	Success StatusCode = 200

	Failure   StatusCode = 400
	Cancelled StatusCode = 401
)

// String returns a suitable string representation for the status code
func (o StatusCode) String() string {
	return map[StatusCode]string{
		Created:    "Created",
		Started:    "Started",
		Stopped:    "Stopped",
		Running:    "Running",
		Cancelling: "Cancelling",
		Pending:    "Pending",
		Success:    "Success",
		Failure:    "Failure",
		Cancelled:  "Cancelled",
		Starting:   "Starting",
		Stopping:   "Stopping",
		Aborting:   "Aborting",
		Freezing:   "Freezing",
		Frozen:     "Frozen",
		Thawed:     "Thawed",
		Error:      "Error",
	}[o]
}

// IsFinal will return true if the status code indicates an end state
func (o StatusCode) IsFinal() bool {
	return int(o) >= 200
}
