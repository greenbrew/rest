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
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"sort"

	"github.com/greenbrew/rest/system"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"

	yaml "gopkg.in/yaml.v2"
)

// RemoteCommand is the command for AMS remote management
type RemoteCommand struct {
	List     RemoteListCmd     `command:"list" alias:"ls" description:"List available remotes"`
	Add      RemoteAddCmd      `command:"add" description:"Add a new remote"`
	Remove   RemoteRemoveCmd   `command:"remove" description:"Remove a remote"`
	Activate RemoteActivateCmd `command:"activate" description:"Set active remote"`
	SetURL   RemoteSetURLCmd   `command:"set-url" description:"Set url of an existing remote"`
}

// RemoteListCmd lists available remotes
type RemoteListCmd struct {
	Format string `long:"format" description:"Output format (table|json)" default:"table"`
}

// RemoteAddCmd adds a new remote
type RemoteAddCmd struct {
	Args struct {
		Name string `positional-arg-name:"name" required:"yes"`
		URL  string `positional-arg-name:"url" required:"yes"`
	} `positional-args:"yes"`
	AcceptCertificate bool `long:"accept-certificate" description:"Implicit accepts remote server certificate if https"`
}

// RemoteRemoveCmd removes an existing remote
type RemoteRemoveCmd struct {
	Args struct {
		Name string `positional-arg-name:"name"`
	} `positional-args:"yes" required:"yes"`
}

// RemoteActivateCmd sets the remote endpoint to use from this client
type RemoteActivateCmd struct {
	Args struct {
		Name string `positional-arg-name:"name"`
	} `positional-args:"yes" required:"yes"`
}

// RemoteSetURLCmd sets the URL of a existing remote
type RemoteSetURLCmd struct {
	Args struct {
		Name string `positional-arg-name:"name"`
		URL  string `positional-arg-name:"url"`
	} `positional-args:"yes" required:"yes"`
}

// Execute lists all currently configured remotes
func (cmd *RemoteListCmd) Execute(args []string) error {
	c, err := loadRemotesConfig()
	if err != nil {
		return err
	}

	if cmd.Format == "table" {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetAutoWrapText(false)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetRowLine(true)
		table.SetHeader([]string{"NAME", "URL", "ACTIVE"})

		data := [][]string{}
		for name, url := range c.Availables {
			data = append(data, []string{
				name,
				url,
				fmt.Sprintf("%t", name == c.Active)})
		}

		sort.Sort(byName(data))
		table.AppendBulk(data)
		table.Render()
	} else if cmd.Format == "yaml" {
		b, err := yaml.Marshal(&c.Availables)
		if err != nil {
			return err
		}
		fmt.Printf(string(b))
	} else if cmd.Format == "json" {
		b, err := json.Marshal(&c.Availables)
		if err != nil {
			return err
		}
		fmt.Printf(string(b))
	} else {
		return errors.Errorf("Unsupported format '%s'", cmd.Format)
	}

	return nil
}

// Execute adds a new remote
func (cmd *RemoteAddCmd) Execute(args []string) error {
	c, err := loadRemotesConfig()
	if err != nil {
		return err
	}

	if _, ok := c.Availables[cmd.Args.Name]; ok {
		return errors.Errorf("Remote with name '%s' already exists", cmd.Args.Name)
	}

	if c.Availables == nil {
		c.Availables = make(map[string]string)
		c.Active = cmd.Args.Name
	}

	u, err := url.Parse(cmd.Args.URL)
	if err != nil {
		return errors.WithMessage(err, "Could not parse provided URL")
	}

	if u.Scheme == "https" {
		// We overwrite existing local server certificate always
		if err := saveServerCertificate(assets, cmd.Args.Name, cmd.Args.URL, cmd.AcceptCertificate, true); err != nil {
			return err
		}
	}

	c.Availables[cmd.Args.Name] = cmd.Args.URL
	return c.save()
}

// Execute removes an existing remote
func (cmd *RemoteRemoveCmd) Execute(args []string) error {
	c, err := loadRemotesConfig()
	if err != nil {
		return err
	}

	if _, ok := c.Availables[cmd.Args.Name]; !ok {
		return errors.Errorf("Remote with name '%s' does not exists", cmd.Args.Name)
	}

	delete(c.Availables, cmd.Args.Name)

	if c.Active == cmd.Args.Name {
		// Take the next remote from the list as the default one (this is random
		// as the map of remotes isn't sorted).
		c.Active = ""
		for name := range c.Availables {
			c.Active = name
			break
		}
	}

	if err := c.save(); err != nil {
		return err
	}

	// As last thing get rid of the certificate we store locally for the remote
	certPath := serverCertPath(assets, cmd.Args.Name)
	if system.PathExists(certPath) {
		if err := os.Remove(certPath); err != nil {
			return err
		}
	}

	return nil
}

// Execute sets the active remote
func (cmd *RemoteActivateCmd) Execute(args []string) error {
	c, err := loadRemotesConfig()
	if err != nil {
		return err
	}

	if _, ok := c.Availables[cmd.Args.Name]; !ok {
		return errors.Errorf("Remote with name '%s' does not exist and can not be set as new default", cmd.Args.Name)
	}

	c.Active = cmd.Args.Name
	return c.save()
}

// Execute sets the URL of a existing remote
func (cmd *RemoteSetURLCmd) Execute(args []string) error {
	c, err := loadRemotesConfig()
	if err != nil {
		return err
	}

	if _, ok := c.Availables[cmd.Args.Name]; !ok {
		return errors.Errorf("Remote with name '%s' does not exist and its URL cannot be set", cmd.Args.Name)
	}

	c.Availables[cmd.Args.Name] = cmd.Args.URL
	return c.save()
}
