package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

var httpServer *http.Server

func initTLS(c *cli.Context) {
	cert, err := tls.LoadX509KeyPair(config.TLSCertificate, config.TLSPrivateKey)
	if err != nil {
		logrus.Fatalf("could not load X.509 key pair: %s", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
	}

	if config.TLSCACertificate != "" {
		pem, err := ioutil.ReadFile(config.TLSCACertificate)
		if err != nil {
			logrus.Fatalf("could not read X.509 certificate: %s", err)
		}

		caPool := x509.NewCertPool()
		caPool.AppendCertsFromPEM(pem)
		tlsConfig.ClientCAs = caPool
		tlsConfig.RootCAs = caPool
	}

	tlsConfig.BuildNameToCertificate()

	http.DefaultClient.Transport = &http.Transport{TLSClientConfig: tlsConfig}
	httpServer = &http.Server{Addr: config.BindAddress, TLSConfig: tlsConfig}
}
