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
	"fmt"
	"io"
	"net/http"

	"github.com/greenbrew/rest/client"
)

type forwardedResponse struct {
	doer    client.Doer
	request *http.Request
	path    string
}

func (r *forwardedResponse) Render(w http.ResponseWriter) error {
	if len(r.path) == 0 {
		r.path = r.request.RequestURI
	}

	forwarded, err := http.NewRequest(r.request.Method, r.path, r.request.Body)
	if err != nil {
		return err
	}

	for key := range r.request.Header {
		forwarded.Header.Set(key, r.request.Header.Get(key))
	}

	response, err := r.doer.Do(forwarded)
	if err != nil {
		return err
	}

	for key := range response.Header {
		w.Header().Set(key, response.Header.Get(key))
	}

	w.WriteHeader(response.StatusCode)
	_, err = io.Copy(w, response.Body)
	return err
}

func (r *forwardedResponse) String() string {
	return fmt.Sprintf("request to %s", r.request.URL)
}

// ForwardedResponse forwards a request to another endpoint and propagates back the response
func ForwardedResponse(doer client.Doer, request *http.Request, path string) Response {
	return &forwardedResponse{doer: doer, request: request, path: path}
}
