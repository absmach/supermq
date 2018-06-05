package cmd

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/mainflux/mainflux"
)

const (
	envCertFile = "MF_CERT_FILE"
	envKeyFile  = "MF_KEY_FILE"
	envCaFile   = "MF_CA_FILE"
)

var (
	httpClient = &http.Client{}
	serverAddr = "https://0.0.0.0"

	defCertFile = os.Getenv("GOPATH") +
		"src/github.com/mainflux/mainflux/docker/ssl/certs/mainflux-server.crt"
	defKeyFile = os.Getenv("GOPATH") +
		"src/github.com/mainflux/mainflux/docker/ssl/certs/mainflux-server.key"
	defCaFile = os.Getenv("GOPATH") +
		"src/github.com/mainflux/mainflux/docker/ssl/certs/ca.crt"
)

// SetServerAddr - set addr using host and port
func SetServerAddr(host string, port int) {
	serverAddr = "https://" + host

	if port != 0 {
		serverAddr += ":" + strconv.Itoa(port)
	}
}

func SetCerts() {
	// Set certificates paths
	certFile := mainflux.Env(envCertFile, defCertFile)
	keyFile := mainflux.Env(envKeyFile, defKeyFile)
	caFile := mainflux.Env(envCaFile, defCaFile)

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
