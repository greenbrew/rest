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
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"github.com/greenbrew/rest/api"
	"github.com/greenbrew/rest/errs"

	check "gopkg.in/check.v1"
)

type responseErrorSuite struct {
	err error
}

var _ = check.Suite(&responseErrorSuite{})

func (s *responseErrorSuite) SetUpSuite(c *check.C) {
	s.err = errors.New("A testing error")
}

func (s *responseErrorSuite) TestBadRequest(c *check.C) {
	s.testErrorResponse(BadRequest, http.StatusBadRequest, c)
}

func (s *responseErrorSuite) TestInternalError(c *check.C) {
	s.testErrorResponse(InternalError, http.StatusInternalServerError, c)
}

func (s *responseErrorSuite) TestAuthorizationError(c *check.C) {
	s.testErrorResponse(AuthorizationError, http.StatusUnauthorized, c)
}

func (s *responseErrorSuite) TestPreconditionFailed(c *check.C) {
	s.testErrorResponse(PreconditionFailed, http.StatusPreconditionFailed, c)
}

func (s *responseErrorSuite) TestNotFoundError(c *check.C) {
	response := NotFoundError(s.err.Error())

	w := newBufferedResponseWriter()
	err := response.Render(w)
	c.Assert(err, check.IsNil)

	// Compare the content of the buffer to the expected result
	expected := &api.Response{
		Type:     api.ResponseTypeError,
		Code:     http.StatusNotFound,
		Error:    errs.NewNotFound(s.err.Error()).Error(),
		Metadata: nil,
	}

	desired := &api.Response{}
	err = json.Unmarshal(w.buffer.Bytes(), desired)
	c.Assert(err, check.IsNil)
	c.Assert(expected, check.DeepEquals, desired)
}

func (s *responseErrorSuite) testErrorResponse(fn func(error) Response, code int, c *check.C) {
	response := fn(s.err)

	w := newBufferedResponseWriter()
	err := response.Render(w)
	c.Assert(err, check.IsNil)

	// Compare the content of the buffer to the expected result
	expected := &api.Response{
		Type:     api.ResponseTypeError,
		Code:     code,
		Error:    s.err.Error(),
		Metadata: nil,
	}

	desired := &api.Response{}
	err = json.Unmarshal(w.buffer.Bytes(), desired)
	c.Assert(err, check.IsNil)
	c.Assert(expected, check.DeepEquals, desired)
}
