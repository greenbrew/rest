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

package cli

import (
	"path/filepath"

	"github.com/greenbrew/rest/system"
	"github.com/pkg/errors"
)

// ErrAssetsPathNotSet reported error when tried to use the cli tool before setting a local assets path
var ErrAssetsPathNotSet = errors.New("Assets path is not set")

var assets string
var remotes *RemotesCfg

// SetAssetsPath sets the local path to store cli tool configuration files, certificates an
// any needed stuff. Triggers the initial configuration setup
func SetAssetsPath(path string) error {
	assets = path

	cfgFilePath := filepath.Join(assets, remotesCfgFile)
	if system.PathExists(cfgFilePath) {
		return nil
	}

	c := newRemotesConfig(cfgFilePath)
	return c.generateDefault()
}

func loadRemotesConfig() (*RemotesCfg, error) {
	if len(assets) == 0 {
		return nil, ErrAssetsPathNotSet
	}

	c := newRemotesConfig(filepath.Join(assets, remotesCfgFile))

	if err := c.load(); err != nil {
		return nil, err
	}

	return c, nil
}
