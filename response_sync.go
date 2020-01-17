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
	"net/http"

	"github.com/greenbrew/rest/api"
)

// Sync response
type syncResponse struct {
	success  bool
	etag     interface{}
	metadata interface{}
	location string
	code     int
	headers  map[string]string
}

func (r *syncResponse) String() string {
	if r.success {
		return "success"
	}
	return "failure"
}

func (r *syncResponse) Render(w http.ResponseWriter) error {
	// Set an appropriate ETag header
	if r.etag != nil {
		etag, err := etagHash(r.etag)
		if err == nil {
			w.Header().Set("ETag", etag)
		}
	}

	// Prepare the JSON response
	status := api.Success
	if !r.success {
		status = api.Failure
	}

	if r.headers != nil {
		for h, v := range r.headers {
			w.Header().Set(h, v)
		}
	}

	if r.location != "" {
		w.Header().Set("Location", r.location)
		code := r.code
		if code == 0 {
			code = 201
		}
		w.WriteHeader(code)
	}

	resp := api.ResponseRaw{
		Response: api.Response{
			Type:       api.ResponseTypeSync,
			Status:     status.String(),
			StatusCode: int(status)},
		Metadata: r.metadata,
	}

	return writeJSON(w, resp)
}

// SyncResponse returns a synchronous http response renderer
func SyncResponse(success bool, metadata interface{}) Response {
	return &syncResponse{success: success, metadata: metadata}
}

// SyncResponseETag returns a synchronous http response renderer with a etag header value
func SyncResponseETag(success bool, metadata interface{}, etag interface{}) Response {
	return &syncResponse{success: success, metadata: metadata, etag: etag}
}

// SyncResponseLocation returns a synchronous http response renderer with a location header
func SyncResponseLocation(success bool, metadata interface{}, location string) Response {
	return &syncResponse{success: success, metadata: metadata, location: location}
}

// SyncResponseRedirect returns a 3xx permanent redirect response
func SyncResponseRedirect(address string) Response {
	return &syncResponse{success: true, location: address, code: http.StatusPermanentRedirect}
}

// SyncResponseHeaders returns a synchronous response with custom headers
func SyncResponseHeaders(success bool, metadata interface{}, headers map[string]string) Response {
	return &syncResponse{success: success, metadata: metadata, headers: headers}
}

// EmptySyncResponse returns an empty synchronous response
var EmptySyncResponse = &syncResponse{success: true, metadata: make(map[string]interface{})}
