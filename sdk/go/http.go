package sdk

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/mainflux/mainflux"
)

const (
	contentTypeJSON      = "application/json"
	contentTypeSenMLJSON = "application/senml+json"
	contentTypeBinary    = "application/octet-stream"

	defCertsPath = "/src/github.com/mainflux/mainflux/docker/ssl/certs/"

	envCertFile = "MF_CERT_FILE"
	envKeyFile  = "MF_KEY_FILE"
	envCaFile   = "MF_CA_FILE"
)

var (
	httpClient = &http.Client{}
	serverAddr = fmt.Sprintf("https://%s", "localhost")

	defCertFile = fmt.Sprintf("%s%s%s", os.Getenv("GOPATH"), defCertsPath, "mainflux-server.crt")
	defKeyFile  = fmt.Sprintf("%s%s%s", os.Getenv("GOPATH"), defCertsPath, "mainflux-server.key")
	defCaFile   = fmt.Sprintf("%s%s%s", os.Getenv("GOPATH"), defCertsPath, "ca.crt")

	limit  = 10
	offset = 0
)

// SetServerAddr - set addr using host and port
func SetServerAddr(proto, host string, port int) {
	serverAddr = fmt.Sprintf("%s://%s", proto, host)

	if port != 0 {
		serverAddr = fmt.Sprintf("%s:%s", serverAddr, strconv.Itoa(port))
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

func sendRequest(req *http.Request, token, contentType string) (*http.Response, error) {
	req.Header.Set("Authorization", token)
	req.Header.Add("Content-Type", contentType)

	return httpClient.Do(req)
}
