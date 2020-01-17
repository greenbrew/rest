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
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/greenbrew/rest/api"
	"github.com/greenbrew/rest/errs"
)

// Different error responses
var (
	NotImplemented     = &errorResponse{http.StatusNotImplemented, "not implemented"}
	NotFound           = &errorResponse{http.StatusNotFound, "not found"}
	Forbidden          = &errorResponse{http.StatusForbidden, "not authorized"}
	Conflict           = &errorResponse{http.StatusConflict, "already exists"}
	ServiceUnavailable = &errorResponse{http.StatusServiceUnavailable, "service unavailable"}
)

// Error response
type errorResponse struct {
	code int
	msg  string
}

func (r *errorResponse) String() string {
	return r.msg
}

func (r *errorResponse) Render(w http.ResponseWriter) error {
	var output io.Writer

	buf := &bytes.Buffer{}
	output = buf

	err := json.NewEncoder(output).Encode(jmap{
		"type":       api.ResponseTypeError,
		"error":      r.msg,
		"error_code": r.code,
	})
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(r.code)
	_, err = w.Write(buf.Bytes())
	return err
}

// BadRequest returns a 400 http response renderer
func BadRequest(err error) Response {
	return &errorResponse{http.StatusBadRequest, err.Error()}
}

// InternalError returns a 500 http response renderer
func InternalError(err error) Response {
	return &errorResponse{http.StatusInternalServerError, err.Error()}
}

// AuthorizationError returns a 401 http response renderer
func AuthorizationError(err error) Response {
	return &errorResponse{http.StatusUnauthorized, err.Error()}
}

// PreconditionFailed returns a 412 http response renderer
func PreconditionFailed(err error) Response {
	return &errorResponse{http.StatusPreconditionFailed, err.Error()}
}

// NotFoundError returns a 404 http response renderer
func NotFoundError(what string) Response {
	return &errorResponse{http.StatusNotFound, errs.NewNotFound(what).Error()}
}

// SmartError returns the right error message based on err.
func SmartError(err error) Response {
	switch err {
	case nil:
		return EmptySyncResponse
	case os.ErrNotExist:
		return NotFound
	case sql.ErrNoRows:
		return NotFound
	case errs.ErrNoSuchObject:
		return NotFound
	case os.ErrPermission:
		return Forbidden
	case errs.ErrAlreadyExists:
		return Conflict
	default:
		return InternalError(err)
	}
}
