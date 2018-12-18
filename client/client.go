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
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"sync"
	"time"

	"github.com/greenbrew/rest/api"
	"github.com/greenbrew/rest/logger"
	"github.com/greenbrew/rest/reverter"
)

// Default value for client requests to wait for a reply
const (
	DefaultTransportTimeout = 30 * time.Second
)

// Doer is the implementation of the Client engine
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// http client to use REST API
type client struct {
	Doer

	serviceURL *url.URL

	listener *EventListener
	mux      sync.Mutex
}

// QueryParams request query parameter
type QueryParams map[string]string

// New returns a REST client. Depending on provided addr parameter, it connects to
// a remote network server or through a unix socket
func New(addr interface{}, tlsConfig *tls.Config) (Client, error) {
	if addr == nil {
		return nil, errors.New("Empty address given")
	}

	switch addr.(type) {
	case *url.URL:
		return newNetworkClient(addr.(*url.URL), tlsConfig)
	case string:
		return newUnixSocketClient(addr.(string))
	default:
		return nil, errors.New("Invalid address type given")
	}
}

// newNetworkClient returns a new REST client pointing to remote address received as parameter
// The connection is TLS enabled or not depending on the remote address schema.
// If TLS is enabled, a proper TLS config must be supplied as second parameter
func newNetworkClient(url *url.URL, tlsConfig *tls.Config) (Client, error) {
	if url == nil {
		return nil, errors.New("Invalid URL given")
	}

	if tlsConfig == nil {
		tlsConfig = &tls.Config{InsecureSkipVerify: true}
	}

	c := &client{
		Doer: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
			Timeout: DefaultTransportTimeout,
		},
		serviceURL: url,
	}

	return c, nil
}

// newUnixSocketClient returns a REST client pointing to local unix socket
func newUnixSocketClient(path string) (Client, error) {
	// Setup a Unix socket dialer
	unixDial := func(network, addr string) (net.Conn, error) {
		raddr, err := net.ResolveUnixAddr("unix", path)
		if err != nil {
			return nil, err
		}

		return net.DialUnix("unix", nil, raddr)
	}

	unixSocketServiceURL, err := url.Parse("http://unix")
	if err != nil {
		return nil, err
	}

	c := &client{
		Doer: &http.Client{
			Transport: &http.Transport{
				Dial:              unixDial,
				DisableKeepAlives: true,
			},
			Timeout: DefaultTransportTimeout,
		},
		serviceURL: unixSocketServiceURL,
	}

	return c, nil
}

func extractErrorFromResponse(resp *http.Response) error {
	var errorResponse struct {
		Code    int    `json:"error_code"`
		Message string `json:"error"`
	}

	err := json.NewDecoder(resp.Body).Decode(&errorResponse)
	if err != nil {
		return err
	}

	return fmt.Errorf(errorResponse.Message)
}

// SetTimeout overwrites default timeout of the client with a new one
func (c *client) SetTransportTimeout(timeout time.Duration) {
	if c.Doer == nil {
		return
	}

	c.Doer.(*http.Client).Timeout = timeout
}

// QueryStruct sends a request to the server and stores response in a struct
func (c *client) QueryStruct(method, path string, params QueryParams, header http.Header, body io.Reader, etag string, target interface{}) (string, error) {
	resp, etag, err := c.CallAPI(method, path, params, header, body, etag)
	if err != nil {
		return "", err
	}

	err = resp.MetadataAsStruct(&target)
	return etag, err
}

// QueryOperation sends a request to the server that will return an async response in an Operation object
// that allows additional logic like wait for completion or cancel it
func (c *client) QueryOperation(method, path string, params QueryParams, header http.Header, body io.Reader, etag string) (Operation, string, error) {
	// Attempt to setup an early event listener
	listener, err := c.GetEvents()

	r := reverter.New()
	defer r.Finish()
	r.Add(func() error {
		if listener != nil {
			listener.Disconnect()
		}
		listener = nil
		return nil
	})

	resp, etag, err := c.CallAPI(method, path, params, header, body, etag)
	if err != nil {
		return nil, "", err
	}

	apiOp, err := resp.MetadataAsOperation()
	if err != nil {
		return nil, "", err
	}

	op := operation{
		Operation: *apiOp,
		c:         &operations{c},
		listener:  listener,
		doneCh:    make(chan bool),
	}

	// Log the data
	logger.Debugf("Got operation from server")
	logger.Debugf(logger.Pretty(op.Operation))

	r.Defuse()
	return &op, etag, nil
}

// CallAPI requests a REST api method with provided query params and body and returns related http response
func (c *client) CallAPI(method, path string, params QueryParams, header http.Header, body io.Reader, etag string) (*api.Response, string, error) {
	u := c.serviceURL.ResolveReference(
		&url.URL{
			Path: path,
		},
	)

	r, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, "", err
	}

	v := r.URL.Query()
	for key, value := range params {
		v.Add(key, value)
	}
	r.URL.RawQuery = v.Encode()

	r.Header.Set("Accept", "application/json")

	if len(etag) > 0 {
		r.Header.Set("If-Match", etag)
	}

	if header != nil {
		for k, v := range header {
			for _, s := range v {
				r.Header.Add(k, s)
			}
		}
	}

	resp, err := c.Doer.Do(r)
	if err != nil {
		return nil, "", err
	}

	defer resp.Body.Close()

	return c.parseResponse(resp)
}

// Internal functions
func (c *client) parseResponse(resp *http.Response) (*api.Response, string, error) {
	// Get the ETag
	etag := resp.Header.Get("ETag")

	// Decode the response
	decoder := json.NewDecoder(resp.Body)
	response := api.Response{}

	err := decoder.Decode(&response)
	if err != nil {
		// Check the return value for a cleaner error
		if resp.StatusCode != http.StatusOK {
			return nil, "", fmt.Errorf("Failed to fetch %s: %s", resp.Request.URL.String(), resp.Status)
		}

		return nil, "", err
	}

	// Handle errors
	if response.Type == api.ResponseTypeError {
		return nil, "", fmt.Errorf(response.Error)
	}

	return &response, etag, nil
}

// APIPath prefixes API version to a path
func APIPath(path ...string) string {
	p := append([]string{"/", api.Version}, path...)
	return filepath.Join(p...)
}
