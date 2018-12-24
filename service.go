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
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"github.com/greenbrew/rest/endpoints"
	"github.com/greenbrew/rest/logger"
	"github.com/greenbrew/rest/pool"
	"github.com/pkg/errors"
)

// A Service can respond to http requests to the REST API
type Service struct {
	// Endpoint services to start at the same time.
	endpoints []endpoints.EndpointEngine
	events    *eventsManager

	UnixSocketPath string
	// Group to own the unix socket created to expose REST locally
	UnixSocketOwner string
	ServerCertPath  string
	ServerKeyPath   string
	CAPath          string

	Router *mux.Router
	// Host where server listens. If empty, server listens in all host IPs
	Host string
	// Service port
	Port int

	// http or https for the main endpoint. Only one of them is started at a time
	schema string

	// Dispatcher of jobs/workers to attend asynchronous requests
	MaxQueuedOperations     int
	MaxConcurrentOperations int
	dispatcher              *pool.Dispatcher

	cache *cache
}

// Init initializes REST service daemon by creating mux router if not created, populate
// router with defined array of APIs, open and setup database
func (d *Service) Init(apis []*API) {
	if d.Router == nil {
		d.Router = mux.NewRouter()
		d.Router.StrictSlash(true)
	}

	if d.Router.NotFoundHandler == nil {
		d.Router.NotFoundHandler = http.HandlerFunc(notFoundHandler)
	}

	d.checkTLSConfig()

	if d.poolEnabled() {
		d.dispatcher = pool.NewDispatcher(d.MaxQueuedOperations, d.MaxConcurrentOperations)
	}

	d.cache = &cache{operations: make(map[string]*Operation)}
	d.events = &eventsManager{listeners: make(map[string]*eventsListener)}

	apis = append(apis, builtinAPI)
	for _, api := range apis {
		for _, c := range api.Commands {
			d.createCmd(api, c)
		}
	}
}

// Start starts the daemon on the configure endpoint
func (d *Service) Start() error {
	if d.dispatcher != nil {
		d.dispatcher.Start()
	}

	if err := d.startEndpoints(); err != nil {
		return errors.Errorf("Failed to start service: %v", err)
	}

	logger.Info("Service ready")
	return nil
}

// Shutdown additional tasks when service shutdown
func (d *Service) Shutdown() error {
	var errs []string
	for _, ep := range d.endpoints {
		err := ep.Stop()
		if err != nil {
			// Join all errors in one single message after stop
			// has been called for all endpoints
			errs = append(errs, err.Error())
		}
	}

	if d.dispatcher != nil {
		d.dispatcher.Stop(true)
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.New(strings.Join(errs, " - "))
}

func (d *Service) checkTLSConfig() {
	// Try TLS enabled by default
	d.schema = "https"
	if len(d.ServerCertPath) == 0 || len(d.ServerKeyPath) == 0 {
		d.schema = "http"
	}
}

func (d *Service) poolEnabled() bool {
	return d.MaxConcurrentOperations > 0 || d.MaxQueuedOperations > 0
}

func (d *Service) createCmd(api *API, c *Command) {
	uri := filepath.Join("/", api.Version, c.Name)

	// Compose middlewares to be applied from ext to in:
	// ApiMiddleware(CommandMiddleware(http.HandlerFunc(...)))
	mws := []MiddlewareFunc{}
	if c.Middleware != nil {
		mws = append(mws, c.Middleware)
	}
	if api.Middleware != nil {
		mws = append(mws, api.Middleware)
	}

	d.Router.Handle(uri, doMws(mws, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var resp Response
		var handler handlerFunc

		switch r.Method {
		case "HEAD":
			// Intercept any response body by hijacking the response writter
			w = &hijack{w}
			fallthrough
		case "GET":
			handler = c.GET
		case "PUT":
			handler = c.PUT
		case "POST":
			handler = c.POST
		case "DELETE":
			handler = c.DELETE
		case "PATCH":
			handler = c.PATCH
		default:
			resp = NotImplemented
		}

		if resp == nil {
			if handler != nil {
				req := &Request{
					HTTPRequest: r,
					daemon:      d,
					version:     api.Version,
				}
				resp = handler(req)
			} else {
				resp = NotFound
			}
		}

		if err := resp.Render(w); err != nil {
			err := SmartError(err).Render(w)
			if err != nil {
				logger.Errorf("Failed writing error for error, giving up")
			}
		}
	})))
}

func (d *Service) startEndpoints() error {
	// Add unix socket endpoint
	if len(d.UnixSocketPath) > 0 {
		localEp := endpoints.NewLocalEndpoint(d.Router, d.UnixSocketPath, d.UnixSocketOwner)
		d.endpoints = append(d.endpoints, endpoints.NewEndpointEngine(localEp))
	}

	if len(d.Host) > 0 || d.Port > 0 {
		// Start either HTTP or HTTPS but not both at the same time
		addr := d.serverAddress()
		if d.tlsEnabled() {
			httpsEp, err := endpoints.NewHTTPSEndpoint(
				d.Router, addr, d.ServerCertPath, d.ServerKeyPath, d.CAPath)
			if err != nil {
				return err
			}
			d.endpoints = append(d.endpoints, endpoints.NewEndpointEngine(httpsEp))
		} else {
			httpEp := endpoints.NewHTTPEndpoint(d.Router, addr)
			d.endpoints = append(d.endpoints, endpoints.NewEndpointEngine(httpEp))
		}
	}

	if len(d.endpoints) == 0 {
		return errors.New("No endpoint configured")
	}

	var wg sync.WaitGroup
	wg.Add(len(d.endpoints))
	for _, ep := range d.endpoints {
		d.serve(ep, &wg)
	}

	// Synchronize all endpoints start process and don't return until
	// all them have started
	wg.Wait()
	return nil
}

func (d *Service) tlsEnabled() bool {
	return d.schema == "https"
}

func (d *Service) serve(ep endpoints.EndpointEngine, wg *sync.WaitGroup) {
	go func(ep endpoints.EndpointEngine, wg *sync.WaitGroup) {
		if err := ep.Start(wg); err != nil {
			logger.Errorf("Could not start endpoint: %v", err)
		}
	}(ep, wg)
}

func (d *Service) serverAddress() string {
	ip := ""
	if len(d.Host) > 0 {
		ip = d.Host
	}

	port := 8080
	if d.Port > 0 {
		port = d.Port
	}

	return ip + ":" + strconv.Itoa(port)
}

// doMws executes middlewares in a daisy chain
func doMws(fs []MiddlewareFunc, inner http.Handler) http.Handler {
	for _, f := range fs {
		inner = f(inner)
	}
	return inner
}

func requestUsesUnixSocket(r *http.Request) bool {
	return r.RemoteAddr == "@"
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "HEAD" {
		w.WriteHeader(http.StatusNotFound)
		w.Write(nil)
	} else {
		NotFound.Render(w)
	}
}
