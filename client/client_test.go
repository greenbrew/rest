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

package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	check "gopkg.in/check.v1"

	"github.com/greenbrew/rest/api"
)

func Test(t *testing.T) { check.TestingT(t) }

type clientSuite struct {
	defaultURL        *url.URL
	defaultTestApkURL string
	cli               *client
	req               *http.Request
	reqs              []*http.Request
	rsp               string
	rsps              []string
	err               error
	doCalls           int
	header            http.Header
	status            int
}

var _ = check.Suite(&clientSuite{})

func (cs *clientSuite) SetUpTest(c *check.C) {
	u, err := url.Parse("http://0.0.0.0:0")
	c.Assert(err, check.IsNil)
	cs.defaultURL = u
	cs.defaultTestApkURL = "http://0.0.0.0:0/apks/1"

	cli, err := New(cs.defaultURL, nil)
	c.Assert(err, check.IsNil)
	cs.cli = cli.(*client)
	cs.cli.Doer = cs
	cs.err = nil
	cs.req = nil
	cs.reqs = nil
	cs.rsp = ""
	cs.rsps = nil
	cs.req = nil
	cs.header = nil
	cs.status = 200
	cs.doCalls = 0
}

func (cs *clientSuite) Do(req *http.Request) (*http.Response, error) {
	cs.req = req
	cs.reqs = append(cs.reqs, req)
	body := cs.rsp
	if cs.doCalls < len(cs.rsps) {
		body = cs.rsps[cs.doCalls]
	}
	rsp := &http.Response{
		Body:       ioutil.NopCloser(strings.NewReader(body)),
		Header:     cs.header,
		StatusCode: cs.status,
	}
	cs.doCalls++
	return rsp, cs.err
}

func (cs *clientSuite) TestDoesNotClientForInvalidURL(c *check.C) {
	cl, err := New(nil, nil)
	c.Assert(err, check.DeepEquals, fmt.Errorf("Empty address given"))
	c.Assert(cl, check.IsNil)
}

func (cs *clientSuite) TestMethodAndPath(c *check.C) {
	cs.rsp = "{}"
	response, _, err := cs.cli.CallAPI("THEMETHOD", "/the/path", nil, nil, nil, "")
	c.Assert(err, check.IsNil)
	c.Check(cs.req.Method, check.Equals, "THEMETHOD")
	c.Check(cs.req.URL.Path, check.Equals, "/the/path")
	c.Check(response.Metadata, check.IsNil)
}
func (cs *clientSuite) TestMethod_invalid(c *check.C) {
	_, _, err := cs.cli.CallAPI("", "", nil, nil, nil, "")
	c.Assert(err, check.NotNil)
}

func (cs *clientSuite) TestPath_invalid(c *check.C) {
	_, _, err := cs.cli.CallAPI("THEMETHOD", "", nil, nil, nil, "")
	c.Assert(err, check.NotNil)
}

func (cs *clientSuite) TestErr(c *check.C) {
	cs.err = errors.New("An error")
	_, _, err := cs.cli.CallAPI("THEMETHOD", "/the/path", nil, nil, nil, "")
	c.Assert(err, check.NotNil)
	c.Assert(err, check.Equals, cs.err)
}

func (cs *clientSuite) TestPassedBackServiceError(c *check.C) {
	cs.rsp = `{"error_code": 1, "error": "Service error"}`
	cs.status = 401
	resp, _, err := cs.cli.CallAPI("THEMETHOD", "/the/path", nil, nil, nil, "")
	c.Assert(err, check.IsNil)
	c.Assert(resp.Code, check.Equals, 1)
	c.Assert(resp.Error, check.Equals, "Service error")
}

func (cs *clientSuite) TestResponseMetadataAsStruct(c *check.C) {
	type Metadata struct {
		Name1      string `json:"name1"`
		Name2      int    `json:"name2"`
		Subcontent struct {
			Name3 string `json:"name3"`
		} `json:"subcontent"`
	}

	var expected struct {
		Metadata `json:"metadata"`
	}

	expected.Metadata.Name1 = "TheName"
	expected.Metadata.Name2 = 42
	expected.Metadata.Subcontent.Name3 = "AnotherName"

	b, err := json.Marshal(&expected)
	c.Assert(err, check.IsNil)
	cs.rsp = string(b)

	response, _, err := cs.cli.CallAPI("THEMETHOD", "/the/path", nil, nil, nil, "")
	c.Assert(err, check.IsNil)

	got := Metadata{}
	response.MetadataAsStruct(&got)
	c.Assert(got, check.DeepEquals, expected.Metadata)
}

func (cs *clientSuite) TestResponseMetadataAsStruct_queryStruct(c *check.C) {
	type Metadata struct {
		Name1      string `json:"name1"`
		Name2      int    `json:"name2"`
		Subcontent struct {
			Name3 string `json:"name3"`
		} `json:"subcontent"`
	}

	var expected struct {
		Metadata `json:"metadata"`
	}

	expected.Metadata.Name1 = "TheName"
	expected.Metadata.Name2 = 42
	expected.Metadata.Subcontent.Name3 = "AnotherName"

	b, err := json.Marshal(&expected)
	c.Assert(err, check.IsNil)
	cs.rsp = string(b)

	got := Metadata{}
	_, err = cs.cli.QueryStruct("THEMETHOD", "/the/path", nil, nil, nil, "", &got)
	c.Assert(err, check.IsNil)
	c.Assert(got, check.DeepEquals, expected.Metadata)
}

func (cs *clientSuite) TestResponseMetadataAsMap(c *check.C) {
	type Metadata struct {
		Name1      string `json:"name1"`
		Name2      int    `json:"name2"`
		Subcontent struct {
			Name3 string `json:"name3"`
		} `json:"subcontent"`
	}

	var expected struct {
		Metadata `json:"metadata"`
	}

	expected.Metadata.Name1 = "TheName"
	expected.Metadata.Name2 = 42
	expected.Metadata.Subcontent.Name3 = "AnotherName"

	b, err := json.Marshal(&expected)
	c.Assert(err, check.IsNil)
	cs.rsp = string(b)

	response, _, err := cs.cli.CallAPI("THEMETHOD", "/the/path", nil, nil, nil, "")
	c.Assert(err, check.IsNil)

	gotMap, err := response.MetadataAsMap()
	c.Assert(err, check.IsNil)

	val, ok := gotMap["name1"]
	c.Assert(ok, check.Equals, true)
	c.Assert(val, check.Equals, "TheName")

	val, ok = gotMap["name2"]
	c.Assert(ok, check.Equals, true)
	c.Assert(val, check.Equals, float64(42))

	val, ok = gotMap["subcontent"]
	c.Assert(ok, check.Equals, true)
	c.Assert(val, check.DeepEquals, map[string]interface{}{"name3": "AnotherName"})

	// Cannot unmarshall into a string slice
	_, err = response.MetadataAsStringSlice()
	c.Assert(err, check.NotNil)
}

func (cs *clientSuite) TestResponseMetadataAsStringSlice_notPossible(c *check.C) {
	type Metadata struct {
		Name1      string `json:"name1"`
		Name2      int    `json:"name2"`
		Subcontent struct {
			Name3 string `json:"name3"`
		} `json:"subcontent"`
	}

	var expected struct {
		Metadata `json:"metadata"`
	}

	expected.Metadata.Name1 = "TheName"
	expected.Metadata.Name2 = 42
	expected.Metadata.Subcontent.Name3 = "AnotherName"

	b, err := json.Marshal(&expected)
	c.Assert(err, check.IsNil)
	cs.rsp = string(b)

	response, _, err := cs.cli.CallAPI("THEMETHOD", "/the/path", nil, nil, nil, "")
	c.Assert(err, check.IsNil)

	// Cannot unmarshall a complex struct into a string slice
	_, err = response.MetadataAsStringSlice()
	c.Assert(err, check.NotNil)
}

func (cs *clientSuite) TestResponseMetadataAsStringSlice(c *check.C) {
	var expected struct {
		Metadata []string `json:"metadata"`
	}

	expected.Metadata = []string{"Foo1", "Foo2", "foo3"}

	b, err := json.Marshal(&expected)
	c.Assert(err, check.IsNil)
	cs.rsp = string(b)

	response, _, err := cs.cli.CallAPI("THEMETHOD", "/the/path", nil, nil, nil, "")
	c.Assert(err, check.IsNil)

	arr, err := response.MetadataAsStringSlice()
	c.Assert(err, check.IsNil)
	c.Assert(arr, check.HasLen, 3)
	c.Assert(arr[0], check.Equals, "Foo1")
	c.Assert(arr[1], check.Equals, "Foo2")
	c.Assert(arr[2], check.Equals, "foo3")
}

func (cs *clientSuite) TestAPIPath(c *check.C) {
	str := APIPath("a", "path")
	c.Assert(str, check.Equals, fmt.Sprintf("/%s/%s/%s", api.Version, "a", "path"))
}
