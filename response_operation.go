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
	"net/http"

	"github.com/greenbrew/rest/api"
)

// Operation response
type operationResponse struct {
	op *Operation
}

func (r *operationResponse) Render(w http.ResponseWriter) error {
	err := r.op.Run()
	if err != nil {
		return err
	}

	url, md, err := r.op.Render()
	if err != nil {
		return err
	}

	body := api.ResponseRaw{
		Response: api.Response{
			Type:       api.ResponseTypeAsync,
			Status:     api.Created.String(),
			StatusCode: int(api.Created),
			Operation:  url,
		},
		Metadata: md,
	}

	w.Header().Set("Location", url)
	w.WriteHeader(202)

	return writeJSON(w, body)
}

func (r *operationResponse) String() string {
	_, md, err := r.op.Render()
	if err != nil {
		return fmt.Sprintf("error: %s", err)
	}

	return md.ID
}

// OperationResponse returns an http response renderer for an operation request
func OperationResponse(op *Operation) Response {
	return &operationResponse{op}
}
