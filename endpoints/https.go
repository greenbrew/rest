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
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/greenbrew/rest/cert"
	"github.com/greenbrew/rest/logger"
	"github.com/greenbrew/rest/system"
	"github.com/greenbrew/rest/tlsconfig"
)

type httpsEp struct {
	endpoint
	certPath string
	keyPath  string
	caPath   string
}

// NewHTTPSEndpoint returns the endpoint for HTTPS transport
func NewHTTPSEndpoint(r *mux.Router, addr, certPath, keyPath, caPath string) (Endpoint, error) {
	cert, err := loadCert(certPath, keyPath)
	if err != nil {
		return nil, err
	}

	ca, err := loadCA(caPath)
	if err != nil {
		return nil, err
	}

	return &httpsEp{
		endpoint: endpoint{
			server: &http.Server{
				Addr:      addr,
				Handler:   r,
				TLSConfig: serverTLSConfig(cert, ca),
			},
		},
		certPath: certPath,
		keyPath:  keyPath,
	}, nil
}

func (h *httpsEp) Name() string {
	return "https"
}

func (h *httpsEp) Init() (interface{}, error) {
	addr := h.server.Addr
	// In case there is no address, the endpoint for https is set to ':https'
	// address, what means, 'listen for https requests in default port and in
	// any available network interface'
	if addr == "" {
		addr = ":https"
	}

	return net.Listen("tcp", addr)
}

func (h *httpsEp) Start(data interface{}) error {
	ln := data.(*net.TCPListener)
	defer ln.Close()

	logger.Infof("Start listening on https://%v", h.server.Addr)
	return h.server.ServeTLS(tcpKeepAliveListener{ln}, h.certPath, h.keyPath)
}

func (h *httpsEp) Stop() error {
	return h.endpoint.Stop()
}

// serverTLSConfig returns a new server-side tls.Config generated from the give
// certificate info.
func serverTLSConfig(cert *cert.Cert, ca *x509.Certificate) *tls.Config {
	config := tlsconfig.New()
	config.Certificates = []tls.Certificate{cert.KeyPair}

	if ca != nil {
		pool := x509.NewCertPool()
		pool.AddCert(ca)
		config.RootCAs = pool
		config.ClientCAs = pool

		logger.Infof("Server is in CA mode, only CA-signed certificates will be allowed")
	}

	config.BuildNameToCertificate()
	return config
}

func loadCert(certPath, keyPath string) (*cert.Cert, error) {
	err := cert.FindOrGenerate(certPath, keyPath, cert.ServerCertificateType)
	if err != nil {
		return nil, err
	}

	keypair, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}

	return &cert.Cert{KeyPair: keypair}, nil
}

func loadCA(path string) (*x509.Certificate, error) {
	var ca *x509.Certificate
	var err error
	if system.PathExists(path) {
		ca, err = cert.Read(path)
	}
	return ca, err
}
