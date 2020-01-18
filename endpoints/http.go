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

package endpoints

import (
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/greenbrew/rest/logger"
)

type httpEp struct {
	endpoint
}

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

// NewHTTPEndpoint returns the endpoint for HTTP transport
func NewHTTPEndpoint(r *mux.Router, addr string) Endpoint {
	return &httpEp{
		endpoint{
			server: &http.Server{
				Addr:    addr,
				Handler: r,
			},
		},
	}
}

func (h *httpEp) Name() string {
	return "http"
}

func (h *httpEp) Init() (interface{}, error) {
	addr := h.server.Addr
	// In case there is no address, the endpoint for http is set to ':https'
	// address, what means, 'listen for https requests in default port and in
	// any available network interface'
	if addr == "" {
		addr = ":http"
	}

	return net.Listen("tcp", addr)
}

func (h *httpEp) Start(data interface{}) error {
	ln := data.(*net.TCPListener)
	defer ln.Close()

	logger.Infof("Start listening on http://%v", h.server.Addr)
	return h.server.Serve(tcpKeepAliveListener{ln})
}

func (h *httpEp) Stop() error {
	return h.endpoint.Stop()
}
