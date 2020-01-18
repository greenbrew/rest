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
	"context"
	"io/ioutil"

	"github.com/pkg/errors"

	yaml "gopkg.in/yaml.v2"
)

const (
	remotesCfgFile    = "remotes.yaml"
	defaultActiveName = "local"
)

var (
	// ErrEmptyRemotesCfgPath returned when tried to access remotes config permament storage
	// before providing a path for it
	ErrEmptyRemotesCfgPath = errors.New("Remotes configuration file path is not set")
)

// RemotesCfg contains the remotes configuration
type RemotesCfg struct {
	path       string
	Active     string            `yaml:"default-remote,omitempty"`
	Availables map[string]string `yaml:"remotes,omitempty"`
}

// newRemotesConfig creates a new configuration object with default values
func newRemotesConfig(path string) *RemotesCfg {
	return &RemotesCfg{path: path}
}

func (c *RemotesCfg) generateDefault() error {
	return ShowProgressSpin(context.Background(), "Creating initial configuration...", func(ctx context.Context) error {
		// Generate unix socket as default remote
		c.Availables = make(map[string]string)
		c.Availables[defaultActiveName] = "unix://"
		c.Active = defaultActiveName

		return c.save()
	})
}

// Load reads the remotes configuration from the underlaying file overwritting existing config values
func (c *RemotesCfg) load() error {
	if len(c.path) == 0 {
		return ErrEmptyRemotesCfgPath
	}

	bytes, err := ioutil.ReadFile(c.path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(bytes, c)
}

// Save writes the remotes configuration to the underlaying file
func (c *RemotesCfg) save() error {
	if len(c.path) == 0 {
		return ErrEmptyRemotesCfgPath
	}

	bytes, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(c.path, bytes, 0644)
}
