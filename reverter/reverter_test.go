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

import (
	"testing"

	check "gopkg.in/check.v1"
)

func Test(t *testing.T) { check.TestingT(t) }

type reverterSuite struct{}

var _ = check.Suite(&reverterSuite{})

func (s *reverterSuite) TestRevert(c *check.C) {
	n := 5
	count := n
	results := []int{}
	revertOperation := func() {
		results = append(results, count)
		count--
	}

	func() {
		r := New()
		defer r.Finish()

		for i := 0; i < n; i++ {
			r.Add(func() error {
				revertOperation()
				return nil
			})
		}
	}()

	c.Assert(count, check.Equals, 0)
	for i := range results {
		c.Assert(results[i], check.Equals, n-i)
	}
}

func (s *reverterSuite) TestDefuse(c *check.C) {
	n := 5
	count := n
	results := []int{}
	revertOperation := func() {
		results = append(results, count)
		count--
	}

	func() {
		r := New()
		defer r.Finish()

		for i := 0; i < n; i++ {
			r.Add(func() error {
				revertOperation()
				return nil
			})
		}

		r.Defuse()
	}()

	c.Assert(count, check.Equals, n)
	c.Assert(results, check.HasLen, 0)
}
