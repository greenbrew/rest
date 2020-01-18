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
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
)

// Cert captures TLS certificate information about public/private keypair
// and an optional CA certificate
type Cert struct {
	KeyPair tls.Certificate
}

// New returns a new Cert instance
func New(keyPair tls.Certificate) *Cert {
	return &Cert{KeyPair: keyPair}
}

// PublicKey is a convenience to encode the underlying public key to ASCII.
func (c *Cert) PublicKey() []byte {
	data := c.KeyPair.Certificate[0]
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: data})
}

// PrivateKey is a convenience to encode the underlying private key.
func (c *Cert) PrivateKey() []byte {
	key, ok := c.KeyPair.PrivateKey.(*rsa.PrivateKey)
	if !ok {
		return nil
	}
	data := x509.MarshalPKCS1PrivateKey(key)
	return pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: data})
}

// Fingerprint returns the fingerprint of the public key.
func (c *Cert) Fingerprint() (string, error) {
	return FingerprintFromStr(string(c.PublicKey()))
}
