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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"os/user"
	"path"
	"time"

	"github.com/greenbrew/rest/system"
)

// CertificateType represents a client or server certificate
type CertificateType int

// Types of certificate that can be generated
const (
	ServerCertificateType CertificateType = 0
	ClientCertificateType CertificateType = 1
)

// FindOrGenerate generates new key and cert if they don't exist
func FindOrGenerate(certFile, keyFile string, certType CertificateType) error {
	if system.PathExists(certFile) && system.PathExists(keyFile) {
		return nil
	}
	return generate(certFile, keyFile, certType)
}

// generate will create and populate a certificate file and a key file
func generate(certFile, keyFile string, certType CertificateType) error {
	dir := path.Dir(certFile)
	err := os.MkdirAll(dir, 0750)
	if err != nil {
		return err
	}
	dir = path.Dir(keyFile)
	err = os.MkdirAll(dir, 0750)
	if err != nil {
		return err
	}

	certBytes, keyBytes, err := generateMemCert(certType)
	if err != nil {
		return err
	}

	certOut, err := os.Create(certFile)
	if err != nil {
		return fmt.Errorf("Failed to open %s for writing: %v", certFile, err)
	}
	certOut.Write(certBytes)
	certOut.Close()

	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("Failed to open %s for writing: %v", keyFile, err)
	}
	keyOut.Write(keyBytes)
	keyOut.Close()
	return nil
}

// generateMemCert creates client or server certificate and key pair,
// returning them as byte arrays in memory.
func generateMemCert(certType CertificateType) ([]byte, []byte, error) {
	privk, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to generate key: %v", err)
	}

	hosts, err := listPossibleHostnames()
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to get my hostname: %v", err)
	}

	validFrom := time.Now()
	validTo := validFrom.Add(10 * 365 * 24 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to generate serial number: %v", err)
	}

	userEntry, err := user.Current()
	var username string
	if err == nil {
		username = userEntry.Username
		if username == "" {
			username = "UNKNOWN"
		}
	} else {
		username = "UNKNOWN"
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "UNKNOWN"
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"greenbrew.io"},
			CommonName:   fmt.Sprintf("%s@%s", username, hostname),
		},
		NotBefore: validFrom,
		NotAfter:  validTo,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}

	switch certType {
	case ServerCertificateType:
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	case ClientCertificateType:
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
	}

	for _, h := range hosts {
		if ip, _, err := net.ParseCIDR(h); err == nil {
			if !ip.IsLinkLocalUnicast() && !ip.IsLinkLocalMulticast() {
				template.IPAddresses = append(template.IPAddresses, ip)
			}
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privk.PublicKey, privk)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to create certificate: %v", err)
	}

	cert := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	key := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privk)})
	return cert, key, nil
}

/*
 * Generate a list of names for which the certificate will be valid.
 * This will include the hostname and ip address
 */
func listPossibleHostnames() ([]string, error) {
	h, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	ret := []string{h}

	ifs, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifs {
		if isLoopback(&iface) {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}

		for _, addr := range addrs {
			ret = append(ret, addr.String())
		}
	}

	return ret, nil
}

// isLoopback returns true if the interface is a loopback network device
func isLoopback(iface *net.Interface) bool {
	return int(iface.Flags&net.FlagLoopback) > 0
}
