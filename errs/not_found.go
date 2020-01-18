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

import "fmt"

// ErrNotFound error struct for a not existing resource
type ErrNotFound struct {
	content
}

// Error returns the error string
func (e ErrNotFound) Error() string {
	return fmt.Sprintf("%v not found", e.What)
}

// NewNotFound returns a new ErrNotFound struct
func NewNotFound(what string) ErrNotFound {
	return ErrNotFound{content{what}}
}
