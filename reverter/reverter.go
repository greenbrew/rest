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

package reverter

import "log"

// RevertFunc describes a function used to revert an operation
type RevertFunc func() error

// Reverter provides functionality to automatically revert a set of executed operations. It
// can be used this way:
//
// r := reverter.New()
// defer r.Finish()
//
// doOperation()
// r.Add(func() error {
//   revertOperation()
//   return nil
// })
//
// if err := doOtherOperation(); err != nil {
//   return err
// }
//
// r.Defuse()
type Reverter struct {
	needRevert bool
	reverters  []RevertFunc
}

// Add adds a new revert function to the reverter which will be called when Finish() is
// called unless the reverter gets defused.
func (r *Reverter) Add(f ...RevertFunc) {
	r.reverters = append(r.reverters, f...)
}

// Defuse defuses the reverter. If defused none of the added revert functions will be
// called when Finish() is invoked.
func (r *Reverter) Defuse() {
	r.needRevert = false
}

// Finish invokes all added revert functions if the reverter was not defused.
func (r *Reverter) Finish() {
	if !r.needRevert {
		return
	}
	// Walk in reverse order through our revertes and call them all
	for n := range r.reverters {
		revert := r.reverters[len(r.reverters)-n-1]
		if err := revert(); err != nil {
			log.Printf("Failed to revert: %v", err)
		}
	}
}

// New constructs a new reverter.
func New() *Reverter {
	return &Reverter{needRevert: true}
}
