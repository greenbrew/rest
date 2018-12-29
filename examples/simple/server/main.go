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

package main

import (
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/greenbrew/rest"

	"github.com/greenbrew/rest/examples/simple"
	"github.com/greenbrew/rest/examples/simple/logger"
)

const (
	host = "0.0.0.0"
	port = 8443
)

func main() {
	d, err := ioutil.TempDir("", "example_")
	if err != nil {
		logger.Critf("Could not create a temp dir: %v", err)
	}

	unixSocketPath := filepath.Join(d, "unix.socket")
	unixSocketOwner := os.Getenv("USER")

	serverCertPath := filepath.Join(d, "server.crt")
	serverKeyPath := filepath.Join(d, "server.key")

	s := &rest.Service{
		UnixSocketPath:  unixSocketPath,
		UnixSocketOwner: unixSocketOwner,
		ServerCertPath:  serverCertPath,
		ServerKeyPath:   serverKeyPath,
		Host:            host,
		Port:            port,
	}

	s.Init([]*rest.API{&simple.API})
	if err := s.Start(); err != nil {
		logger.Critf("Could not start the service: %v", err)
	}

	logger.Infof("Server unix socket started at %v\n", unixSocketPath)
	logger.Infof("Server HTTPS started in port %v\n", port)

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	select {
	case sig := <-ch:
		logger.Errorf("Exiting on %s", sig)
	}
}
