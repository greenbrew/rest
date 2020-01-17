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

package system

import (
	"io/ioutil"
	"os"
	"testing"

	check "gopkg.in/check.v1"
)

func Test(t *testing.T) { check.TestingT(t) }

type existsSuite struct{}

var _ = check.Suite(&existsSuite{})

func (s *existsSuite) TestExists(c *check.C) {
	c.Assert(PathExists("/invented"), check.Equals, false)

	path, err := ioutil.TempDir("", "")
	defer os.RemoveAll(path)
	c.Assert(err, check.IsNil)
	c.Assert(PathExists(path), check.Equals, true)
}
