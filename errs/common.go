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

package errs

import "errors"

// ErrAlreadyExists describes an error when a resource already exists
var ErrAlreadyExists = errors.New("Already exists")

// ErrInvalidInstanceType describes the error when an invalid instance type was used
var ErrInvalidInstanceType = errors.New("Invalid instance type")

// ErrNoSuchObject describes the error when a resource does not exists
var ErrNoSuchObject = errors.New("Not found")

type content struct {
	What string
}
