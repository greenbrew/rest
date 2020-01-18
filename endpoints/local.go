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
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/lxc/lxd/shared"
	"github.com/pkg/errors"

	"github.com/greenbrew/rest/logger"
	"github.com/greenbrew/rest/system"
)

const (
	unixSocketPermissions = 0660
)

type local struct {
	endpoint
	unixSocketPath string
	owner          string
}

// NewLocalEndpoint returns the endpoint associated to local unix socket
func NewLocalEndpoint(r *mux.Router, unixSocketPath, owner string) Endpoint {
	return &local{
		endpoint{
			server: &http.Server{
				Handler: r,
			},
		},
		unixSocketPath,
		owner,
	}
}

func (l *local) Name() string {
	return "local"
}

func (l *local) Init() (interface{}, error) {
	if len(l.unixSocketPath) == 0 {
		return nil, errors.New("Empty unix socket path given")
	}

	if err := removeStaleUnixSocket(l.unixSocketPath); err != nil {
		return nil, err
	}

	addr, err := net.ResolveUnixAddr("unix", l.unixSocketPath)
	if err != nil {
		return nil, errors.WithMessage(err, "Cannot resolve socket path")
	}

	listener, err := net.ListenUnix("unix", addr)
	if err != nil {
		return nil, errors.WithMessage(err, "Cannot bind socket")
	}

	var owner string
	if len(l.owner) > 0 {
		owner = l.owner
	}

	err = setUnixSocketOwnership(l.unixSocketPath, owner)
	if err != nil {
		return nil, errors.WithMessage(err, "Could not set unix socket ownership")
	}

	err = setUnixSocketPermissions(l.unixSocketPath, unixSocketPermissions)
	if err != nil {
		return nil, errors.WithMessage(err, "Could not set unix socket permissions")
	}

	return listener, nil
}

func (l *local) Start(data interface{}) error {
	listener := data.(*net.UnixListener)
	defer listener.Close()

	logger.Infof("Start listening on unix socket at %v", l.unixSocketPath)
	return l.server.Serve(listener)
}

func (l *local) Stop() error {
	return l.endpoint.Stop()
}

// Remove any stale socket file at the given path.
func removeStaleUnixSocket(path string) error {
	// If there's no socket file at all, there's nothing to do.
	if !system.PathExists(path) {
		return nil
	}

	logger.Debugf("Detected stale unix socket, deleting")
	err := os.Remove(path)
	if err != nil {
		return fmt.Errorf("could not delete stale local socket: %v", err)
	}

	return nil
}

// Change the file mode of the given unix socket file,
func setUnixSocketPermissions(path string, mode os.FileMode) error {
	err := os.Chmod(path, mode)
	if err != nil {
		return fmt.Errorf("cannot set permissions on local socket: %v", err)
	}
	return nil
}

// Change the ownership of the given unix socket file,
func setUnixSocketOwnership(path string, group string) error {
	var gid int
	var err error

	if len(group) > 0 {
		gid, err = shared.GroupId(group)
		if err != nil {
			return fmt.Errorf("cannot get group ID of '%s': %v", group, err)
		}
	} else {
		gid = os.Getgid()
	}

	err = os.Chown(path, os.Getuid(), gid)
	if err != nil {
		return fmt.Errorf("cannot change ownership on local socket: %v", err)

	}

	return nil
}
