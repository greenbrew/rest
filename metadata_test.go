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

package rest

import (
	"strings"

	check "gopkg.in/check.v1"
)

type metadataSuite struct{}

var _ = check.Suite(&metadataSuite{})

func (s *metadataSuite) TestValidMap(c *check.C) {
	m := map[string]interface{}{
		"one":   1,
		"two":   "dos",
		"three": 3.0,
	}

	m2, err := parseMetadata(m)
	c.Assert(err, check.IsNil)
	c.Assert(m2, check.HasLen, 3)
	c.Assert(m2["one"], check.Equals, 1)
	c.Assert(m2["two"], check.Equals, "dos")
	c.Assert(m2["three"], check.Equals, 3.0)
}

func (s *metadataSuite) TestInvalidValidMap(c *check.C) {
	m := map[interface{}]interface{}{
		78:      1,
		"two":   "dos",
		"three": 3.0,
	}

	m2, err := parseMetadata(m)
	c.Assert(m2, check.IsNil)
	c.Assert(err, check.NotNil)
	c.Assert(strings.Contains(err.Error(), "Invalid"), check.Equals, true)
}

func (s *metadataSuite) TestNilMetadata(c *check.C) {
	m, err := parseMetadata(nil)
	c.Assert(err, check.IsNil)
	c.Assert(m, check.IsNil)
}

func (s *metadataSuite) TestInvalidMetadata(c *check.C) {
	m, err := parseMetadata(42)
	c.Assert(m, check.IsNil)
	c.Assert(err, check.NotNil)
	c.Assert(strings.Contains(err.Error(), "Invalid"), check.Equals, true)
}
