package cmd

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

var (
	httpClient = &http.Client{}
	serverAddr = "https://0.0.0.0"
)

const (
	certFile = "../../docker/ssl/certs/mainflux-server.crt"
	keyFile  = "../../docker/ssl/certs/mainflux-server.key"
	caFile   = "../../docker/ssl/certs/ca.crt"
)

// SetServerAddr - set addr using host and port
func SetServerAddr(host string, port int) {
	serverAddr = "https://" + host

	if port != 0 {
		serverAddr += ":" + strconv.Itoa(port)
	}
}

func SetCerts() {
	// Load client cert
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatal(err)
	}

	// Load CA cert
	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Setup HTTPS client
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}
	tlsConfig.BuildNameToCertificate()
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	httpClient = &http.Client{Transport: transport}
}
