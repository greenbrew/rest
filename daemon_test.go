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
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	telnet "github.com/reiver/go-telnet"

	check "gopkg.in/check.v1"

	"github.com/greenbrew/rest/cert"
	"github.com/greenbrew/rest/freeport"
	"github.com/greenbrew/rest/tlsconfig"
)

func Test(t *testing.T) { check.TestingT(t) }

type daemonSuite struct{}

var _ = check.Suite(&daemonSuite{})

func (s *daemonSuite) SetUpTest(c *check.C) {
}

func (s *daemonSuite) TestCannotStartWithoutEndpoints(c *check.C) {
	d := Daemon{}
	d.Init([]*API{})
	err := d.Start()
	c.Assert(err, check.ErrorMatches, "Failed to start service: No endpoint configured")
}

func (s *daemonSuite) TestCanStartUnixSocket(c *check.C) {
	tmpDir, err := ioutil.TempDir("", "")
	c.Assert(err, check.IsNil)

	unixSocketPath := filepath.Join(tmpDir, "unix.socket")
	d := Daemon{
		UnixSocketPath: unixSocketPath,
	}

	d.Init([]*API{})
	err = d.Start()
	c.Assert(err, check.IsNil)

	assertUnixSocket(unixSocketPath, c)

	err = d.Shutdown()
	c.Assert(err, check.IsNil)
}

func (s *daemonSuite) TestCanStartHTTP(c *check.C) {
	port, err := freeport.GetFreePort()
	c.Assert(err, check.IsNil)

	host, err := os.Hostname()
	c.Assert(err, check.IsNil)

	d := Daemon{
		Host: host,
		Port: port,
	}

	d.Init([]*API{})
	err = d.Start()
	c.Assert(err, check.IsNil)

	assertHTTP(port, c)
	assertNotHTTPS(port, c)

	err = d.Shutdown()
	c.Assert(err, check.IsNil)
}

func (s *daemonSuite) TestCanStartHTTPWithOnlyPort(c *check.C) {
	port, err := freeport.GetFreePort()
	c.Assert(err, check.IsNil)

	d := Daemon{
		Port: port,
	}

	d.Init([]*API{})
	err = d.Start()
	c.Assert(err, check.IsNil)

	assertHTTP(port, c)
	assertNotHTTPS(port, c)

	err = d.Shutdown()
	c.Assert(err, check.IsNil)
}

func (s *daemonSuite) TestCanStartHTTPS(c *check.C) {
	// Find a free port
	port, err := freeport.GetFreePort()
	c.Assert(err, check.IsNil)

	// Generate temporary server certificate for the test
	tmpDir, err := ioutil.TempDir("", "")
	c.Assert(err, check.IsNil)
	defer os.RemoveAll(tmpDir)
	certPath := filepath.Join(tmpDir, "server.crt")
	keyPath := filepath.Join(tmpDir, "server.key")
	err = cert.FindOrGenerate(certPath, keyPath, cert.ServerCertificateType)
	c.Assert(err, check.IsNil)

	d := Daemon{
		Port:           port,
		ServerCertPath: certPath,
		ServerKeyPath:  keyPath,
	}

	d.Init([]*API{})
	err = d.Start()
	c.Assert(err, check.IsNil)

	assertHTTPS(port, c)
	assertNotHTTP(port, c)

	err = d.Shutdown()
	c.Assert(err, check.IsNil)
}

func (s *daemonSuite) TestAPI(c *check.C) {
	// define the API
	whateverCmd := &Command{
		Name: "whatever",
		GET:  whateverGet,
		POST: whateverPost,
	}
	notFoundResourceCmd := &Command{
		Name: "notfound",
		GET:  notFoundResourceGet,
	}
	var api = &API{
		Version: "0.9",
		Commands: []*Command{
			whateverCmd,
			notFoundResourceCmd,
		},
	}

	// Start the service
	port, err := freeport.GetFreePort()
	c.Assert(err, check.IsNil)

	d := Daemon{
		Port: port,
	}

	d.Init([]*API{api})
	err = d.Start()
	c.Assert(err, check.IsNil)

	assertHTTP(port, c)
	assertNotHTTPS(port, c)

	// Call api methods
	host, err := os.Hostname()
	c.Assert(err, check.IsNil)

	// HEAD
	response, err := http.Head(fmt.Sprintf("http://%s:%d/0.9/whatever", host, port))
	c.Assert(err, check.IsNil)
	c.Assert(response.StatusCode, check.Equals, 200)
	c.Assert(response.ContentLength <= 0, check.Equals, true)

	// GET
	response, err = http.Get(fmt.Sprintf("http://%s:%d/0.9/whatever", host, port))
	c.Assert(err, check.IsNil)
	c.Assert(response.StatusCode, check.Equals, 200)

	// GET 'notfound' resource
	response, err = http.Get(fmt.Sprintf("http://%s:%d/0.9/notfound", host, port))
	c.Assert(err, check.IsNil)
	c.Assert(response.StatusCode, check.Equals, 404)

	// POST
	var jsonStr = []byte(`{"title":"Buy cheese and bread for breakfast."}`)
	response, err = http.Post(fmt.Sprintf("http://%s:%d/0.9/whatever", host, port), "application/json", bytes.NewBuffer(jsonStr))
	c.Assert(err, check.IsNil)
	c.Assert(response.StatusCode, check.Equals, 202)

	// GET Not existing resource
	response, err = http.Get(fmt.Sprintf("http://%s:%d/0.9/invalid", host, port))
	c.Assert(err, check.IsNil)
	c.Assert(response.StatusCode, check.Equals, 404)

	// HEAD Not existing resource
	response, err = http.Head(fmt.Sprintf("http://%s:%d/0.9/invalid", host, port))
	c.Assert(err, check.IsNil)
	c.Assert(response.StatusCode, check.Equals, 404)
	c.Assert(response.ContentLength <= 0, check.Equals, true)

	// HEAD 'notfound' resource
	response, err = http.Head(fmt.Sprintf("http://%s:%d/0.9/notfound", host, port))
	c.Assert(err, check.IsNil)
	c.Assert(response.StatusCode, check.Equals, 404)
	c.Assert(response.ContentLength <= 0, check.Equals, true)

	err = d.Shutdown()
	c.Assert(err, check.IsNil)

}

func whateverGet(d *Daemon, r *http.Request) Response {
	return SyncResponse(true, []string{filepath.Join(r.URL.Path, "1")})
}

func whateverPost(d *Daemon, r *http.Request) Response {
	run := func(op *Operation) error {
		return nil
	}

	resources := map[string][]string{}
	resources["whatever"] = []string{"1"}

	op, err := d.CreateOperation("Creating whatever", resources, nil, run, nil)
	if err != nil {
		return InternalError(err)
	}
	return OperationResponse(op)
}

func notFoundResourceGet(d *Daemon, r *http.Request) Response {
	return NotFound
}

func assertUnixSocket(socketAddr string, c *check.C) {
	// Test if unix socket is reachable
	unixDial := func(network, addr string) (net.Conn, error) {
		raddr, err := net.ResolveUnixAddr("unix", socketAddr)
		if err != nil {
			return nil, err
		}

		return net.DialUnix("unix", nil, raddr)
	}

	httpc := &http.Client{
		Transport: &http.Transport{
			Dial: unixDial,
		},
	}
	// Do a get through unix socket to a builtIn URL
	response, err := httpc.Get("http://unix/1.0/operations")
	c.Assert(err, check.IsNil)
	c.Assert(response.StatusCode, check.Equals, 200)
}

func assertHTTP(port int, c *check.C) {
	host, err := os.Hostname()
	c.Assert(err, check.IsNil)

	// Do a get to a well known URL and verify it returns 200 Ok
	response, err := http.Get(fmt.Sprintf("http://%s:%d/1.0/operations", host, port))
	c.Assert(err, check.IsNil)
	c.Assert(response.StatusCode, check.Equals, 200)
}

func assertHTTPS(port int, c *check.C) {
	host, err := os.Hostname()
	c.Assert(err, check.IsNil)

	// Fetch server certificate
	remoteCert, err := cert.FetchRemoteCertificate(fmt.Sprintf("https://%s:%d", host, port))
	c.Assert(err, check.IsNil)
	c.Assert(remoteCert, check.NotNil)

	// Telnet localhost:port trusting remote certificate into the TLS config to use
	tlsConfig := tlsconfig.New()
	tlsConfig.RootCAs = x509.NewCertPool()
	tlsConfig.RootCAs.AddCert(remoteCert)
	err = telnet.DialToAndCallTLS(fmt.Sprintf("%s:%d", host, port), telnet.StandardCaller, tlsConfig)
	c.Assert(err, check.IsNil)

	// Do a get to a well known URL and verify it returns 200 Ok
	httpc := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}
	response, err := httpc.Get(fmt.Sprintf("https://%s:%d/1.0/operations", host, port))
	c.Assert(err, check.IsNil)
	c.Assert(response.StatusCode, check.Equals, 200)
}

func assertNotHTTP(port int, c *check.C) {
	host, err := os.Hostname()
	c.Assert(err, check.IsNil)

	// Do a get to a well known URL and verify it returns error
	_, err = http.Get(fmt.Sprintf("http://%s:%d/1.0/operations", host, port))
	c.Assert(err, check.NotNil)
}

func assertNotHTTPS(port int, c *check.C) {
	host, err := os.Hostname()
	c.Assert(err, check.IsNil)

	// Impossible to fetch server certificate
	_, err = cert.FetchRemoteCertificate(fmt.Sprintf("https://%s:%d", host, port))
	c.Assert(err, check.NotNil)
}
