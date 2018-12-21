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
	"fmt"
	"net/url"

	"github.com/greenbrew/rest/cert"
	"github.com/greenbrew/rest/client"
	"github.com/greenbrew/rest/tlsconfig"
)

// NewClient returns a REST client pointing to the default remote
func NewClient() (client.Client, error) {
	cfg, err := loadRemotesConfig()
	if err != nil {
		return nil, err
	}

	if _, ok := cfg.Availables[cfg.Active]; !ok {
		return nil, fmt.Errorf("The configured active remote '%s' does not exist", cfg.Active)
	}

	active := cfg.Availables[cfg.Active]
	u, err := url.Parse(active)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "https":
		tlsCfg := tlsconfig.New()
		certPath := serverCertPath(assets, active)
		serverCert, err := cert.Read(certPath)
		if err != nil {
			return nil, err
		}
		tlsconfig.Finalize(tlsCfg, serverCert)
		return client.New(u, tlsCfg)
	case "unix":
		return client.New(u.Path, nil)
	default:
		return client.New(u, nil)
	}
}
