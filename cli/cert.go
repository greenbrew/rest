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
	"os"
	"path/filepath"

	"github.com/greenbrew/rest/cert"
	"github.com/greenbrew/rest/system"
	"github.com/pkg/errors"
)

const (
	serverCertsPath = "servercerts"
)

// Saves a server cert to a local path. This is used to connect to an HTTPS endpoint
func saveServerCertificate(cfgPath, name, address string, acceptCertificate, overrideExisting bool) error {
	c, err := cert.FetchRemoteCertificate(address)
	if err != nil {
		return errors.WithMessage(err, "Could not get server certificate")
	}

	fp := cert.Fingerprint(c)
	if !acceptCertificate {
		fmt.Printf("Remote certificate fingerprint: %s\n", fp)
		if !AskBool("ok? (yes/no) [default=no]: ", "no") {
			return errors.New("User aborted configuration")
		}
	}

	// Check if already exists the server certificate locally and get its fingerprint
	serverCertsAbsPath := filepath.Join(cfgPath, serverCertsPath)
	certFilename := filepath.Join(serverCertsAbsPath, name+".crt")
	if system.PathExists(certFilename) && !overrideExisting {
		existingCert, err := cert.Read(certFilename)
		if err != nil {
			return err
		}

		existingFp := cert.Fingerprint(existingCert)
		// If remote and local are the same we don't have anything to do
		if existingFp == fp {
			return nil
		}
	}

	if err := os.MkdirAll(serverCertsAbsPath, 0750); err != nil {
		return errors.New("Could not create server cert dir")
	}

	return cert.Save(c, certFilename)
}

// Returns the path to the certificate of the given server
func serverCertPath(configPath, name string) string {
	return filepath.Join(configPath, serverCertsPath, name+".crt")
}
