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

package api

import (
	"encoding/json"
)

// "there are three standard return types: Standard return value,
// Background operation, Error", each returning a JSON object with the
// following "type" field:
const (
	ResponseTypeSync  ResponseType = "sync"
	ResponseTypeAsync ResponseType = "async"
	ResponseTypeError ResponseType = "error"
)

// ResponseType represents a valid LXD response type
type ResponseType string

// ResponseRaw represents a REST operation in its original form
type ResponseRaw struct {
	Response `yaml:",inline"`
	Metadata interface{} `json:"metadata" yaml:"metadata"`
}

// Response represents a LXD operation
type Response struct {
	Type ResponseType `json:"type" yaml:"type"`

	// Valid only for Sync responses
	Status     string `json:"status" yaml:"status"`
	StatusCode int    `json:"status_code" yaml:"status_code"`

	// Valid only for Async responses
	Operation string `json:"operation" yaml:"operation"`

	// Valid only for Error responses
	Code  int    `json:"error_code" yaml:"error_code"`
	Error string `json:"error" yaml:"error"`

	// Valid for Sync and Error responses
	Metadata json.RawMessage `json:"metadata" yaml:"metadata"`
}

// MetadataAsMap parses the Response metadata into a map
func (r *Response) MetadataAsMap() (map[string]interface{}, error) {
	ret := map[string]interface{}{}
	err := r.MetadataAsStruct(&ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

// MetadataAsOperation turns the Response metadata into an Operation
func (r *Response) MetadataAsOperation() (*Operation, error) {
	op := Operation{}
	err := r.MetadataAsStruct(&op)
	if err != nil {
		return nil, err
	}

	return &op, nil
}

// MetadataAsStringSlice parses the Response metadata into a slice of string
func (r *Response) MetadataAsStringSlice() ([]string, error) {
	sl := []string{}
	err := r.MetadataAsStruct(&sl)
	if err != nil {
		return nil, err
	}

	return sl, nil
}

// MetadataAsStruct parses the Response metadata into a provided struct
func (r *Response) MetadataAsStruct(target interface{}) error {
	return json.Unmarshal(r.Metadata, &target)
}
