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
	"strings"

	"github.com/greenbrew/rest/api"
	check "gopkg.in/check.v1"
)

type responseSyncSuite struct {
	body struct {
		Astruct struct {
			Content string `json:"content"`
		} `json:"other_struct"`
		Number int `json:"number"`
	}
}

var _ = check.Suite(&responseSyncSuite{})

func (s *responseSyncSuite) SetUpSuite(c *check.C) {
	s.body = struct {
		Astruct struct {
			Content string `json:"content"`
		} `json:"other_struct"`
		Number int `json:"number"`
	}{
		Astruct: struct {
			Content string `json:"content"`
		}{
			Content: "My Content",
		},
		Number: 42,
	}
}

func (s *responseSyncSuite) TestSyncResponse(c *check.C) {
	sr := SyncResponse(true, s.body)

	w := newBufferedResponseWriter()
	err := sr.Render(w)
	c.Assert(err, check.IsNil)

	// Compare the content of the buffer to the expected result
	metadata, err := json.Marshal(s.body)
	c.Assert(err, check.IsNil)
	expected := &api.Response{
		Type:       api.ResponseTypeSync,
		Status:     api.Success.String(),
		StatusCode: int(api.Success),
		Metadata:   metadata,
	}

	result := &api.Response{}
	err = json.Unmarshal(w.buffer.Bytes(), result)
	c.Assert(err, check.IsNil)
	c.Assert(expected, check.DeepEquals, result)
}

func (s *responseSyncSuite) TestSyncResponseETag(c *check.C) {
	etag := "myetag"
	sr := SyncResponseETag(true, s.body, etag)

	w := newBufferedResponseWriter()
	err := sr.Render(w)
	c.Assert(err, check.IsNil)

	// Compare the content of the buffer to the expected result
	metadata, err := json.Marshal(s.body)
	c.Assert(err, check.IsNil)
	expected := &api.Response{
		Type:       api.ResponseTypeSync,
		Status:     api.Success.String(),
		StatusCode: int(api.Success),
		Metadata:   metadata,
	}

	result := &api.Response{}
	err = json.Unmarshal(w.buffer.Bytes(), result)
	c.Assert(err, check.IsNil)
	c.Assert(expected, check.DeepEquals, result)

	// Verify etag header
	expectedHash, err := etagHash(etag)
	c.Assert(err, check.IsNil)

	// search for 'Etag' instead of 'ETag', because header keys are
	// all formatted for just have first letter in upper case
	resultHash, ok := w.headers["Etag"]
	c.Assert(ok, check.Equals, true)
	c.Assert(resultHash, check.HasLen, 1)
	c.Assert(resultHash[0], check.Equals, expectedHash)
}

func (s *responseSyncSuite) TestSyncResponseLocation(c *check.C) {
	location := "http://host:8080/the/location/path"
	sr := SyncResponseLocation(true, s.body, location)

	w := newBufferedResponseWriter()
	err := sr.Render(w)
	c.Assert(err, check.IsNil)

	// Compare the content of the buffer to the expected result
	metadata, err := json.Marshal(s.body)
	c.Assert(err, check.IsNil)
	expected := &api.Response{
		Type:       api.ResponseTypeSync,
		Status:     api.Success.String(),
		StatusCode: int(api.Success),
		Metadata:   metadata,
	}

	result := &api.Response{}
	err = json.Unmarshal(w.buffer.Bytes(), result)
	c.Assert(err, check.IsNil)
	c.Assert(expected, check.DeepEquals, result)

	resultLocation, ok := w.headers["Location"]
	c.Assert(ok, check.Equals, true)
	c.Assert(resultLocation, check.HasLen, 1)
	c.Assert(resultLocation[0], check.Equals, location)
}

func (s *responseSyncSuite) TestSyncResponseRedirect(c *check.C) {
	addr := "http://host:8080/the/location/path"
	sr := SyncResponseRedirect(addr)

	w := newBufferedResponseWriter()
	err := sr.Render(w)
	c.Assert(err, check.IsNil)

	// Compare the content of the buffer to the expected result
	metadata, err := json.Marshal(nil)
	c.Assert(err, check.IsNil)
	expected := &api.Response{
		Type:       api.ResponseTypeSync,
		Status:     api.Success.String(),
		StatusCode: int(api.Success),
		Metadata:   metadata,
	}

	result := &api.Response{}
	err = json.Unmarshal(w.buffer.Bytes(), result)
	c.Assert(err, check.IsNil)
	c.Assert(expected, check.DeepEquals, result)

	c.Assert(w.statusCode, check.Equals, http.StatusPermanentRedirect)

	resultLocation, ok := w.headers["Location"]
	c.Assert(ok, check.Equals, true)
	c.Assert(resultLocation, check.HasLen, 1)
	c.Assert(resultLocation[0], check.Equals, addr)
}

func (s *responseSyncSuite) TestSyncResponseHeaders(c *check.C) {
	headers := map[string]string{
		"firstHeader":  "FirstValue",
		"SecondHeader": "SecondValue",
	}
	sr := SyncResponseHeaders(true, s.body, headers)

	w := newBufferedResponseWriter()
	err := sr.Render(w)
	c.Assert(err, check.IsNil)

	// Compare the content of the buffer to the expected result
	metadata, err := json.Marshal(s.body)
	c.Assert(err, check.IsNil)
	expected := &api.Response{
		Type:       api.ResponseTypeSync,
		Status:     api.Success.String(),
		StatusCode: int(api.Success),
		Metadata:   metadata,
	}

	result := &api.Response{}
	err = json.Unmarshal(w.buffer.Bytes(), result)
	c.Assert(err, check.IsNil)
	c.Assert(expected, check.DeepEquals, result)

	// Verify headers. We cannot compare directly because they are not
	// the same type.
	for k, v := range headers {
		// Keys in http.Header have first letter in upper case and the rest in lower
		key := strings.ToUpper(string(k[0])) + strings.ToLower(k[1:])

		value, ok := w.headers[key]
		c.Assert(ok, check.Equals, true)
		c.Assert(value, check.HasLen, 1)
		c.Assert(value[0], check.Equals, v)
		delete(w.headers, key)
	}
	c.Assert(w.headers, check.HasLen, 0)
}
