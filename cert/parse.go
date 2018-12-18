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

package cert

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
)

// Read reads an x509 certificate from a filepath
func Read(path string) (*x509.Certificate, error) {
	cf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	certBlock, _ := pem.Decode(cf)
	if certBlock == nil {
		return nil, fmt.Errorf("Invalid certificate file")
	}

	return x509.ParseCertificate(certBlock.Bytes)
}

// FingerprintFromStr returns a fingerprint of a raw certificate parameter
func FingerprintFromStr(c string) (string, error) {
	pemCertificate, _ := pem.Decode([]byte(c))
	if pemCertificate == nil {
		return "", errors.New("invalid certificate")
	}

	cert, err := x509.ParseCertificate(pemCertificate.Bytes)
	if err != nil {
		return "", err
	}

	return Fingerprint(cert), nil
}

// Fingerprint returns the fingerprint of a certificate
func Fingerprint(cert *x509.Certificate) string {
	return fmt.Sprintf("%x", sha256.Sum256(cert.Raw))
}
