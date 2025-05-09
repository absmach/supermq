// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package callout_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/absmach/supermq/pkg/callout"
	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	permission     = "test_permission"
	entityType     = "client"
	userID         = "user_id"
	domainID       = "domain_id"
	filePermission = 0o644
)

var pl = map[string]interface{}{
	"entity_type": entityType,
	"sender":      userID,
	"domain":      domainID,
	"time":        time.Now().String(),
	"permission":  permission,
}

func TestNewCalloutClient(t *testing.T) {
	cases := []struct {
		desc     string
		ctls     bool
		certPath string
		keyPath  string
		caPath   string
		timeout  time.Duration
		err      error
	}{
		{
			desc:    "successful client creation without TLS",
			ctls:    false,
			timeout: time.Second,
		},
		{
			desc:     "successful client creation with TLS",
			ctls:     true,
			certPath: "client.crt",
			keyPath:  "client.key",
			caPath:   "ca.crt",
			timeout:  time.Second,
		},
		{
			desc:     "failed client creation with invalid cert",
			ctls:     true,
			certPath: "invalid.crt",
			keyPath:  "invalid.key",
			caPath:   "invalid.ca",
			timeout:  time.Second,
			err:      errors.New("tls: failed to find any PEM data in certificate input"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.desc == "successful client creation with TLS" {
				generateAndWriteCertificates(t, tc.caPath, tc.certPath, tc.keyPath)

				defer func() {
					os.Remove(tc.certPath)
					os.Remove(tc.keyPath)
					os.Remove(tc.caPath)
				}()
			} else if tc.desc == "failed client creation with invalid cert" {
				writeFile(t, tc.certPath, []byte("invalid cert content"))
				writeFile(t, tc.keyPath, []byte("invalid key content"))
				writeFile(t, tc.caPath, []byte("invalid ca content"))

				defer func() {
					os.Remove(tc.certPath)
					os.Remove(tc.keyPath)
					os.Remove(tc.caPath)
				}()
			}

			client, err := callout.NewCalloutClient(tc.ctls, tc.certPath, tc.keyPath, tc.caPath, tc.timeout)
			assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
			if err == nil {
				assert.NotNil(t, client)
			}
		})
	}
}

func generateAndWriteCertificates(t *testing.T, caPath, certPath, keyPath string) {
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Failed to generate CA private key")

	caTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test CA"},
			CommonName:   "Test CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caKey.PublicKey, caKey)
	require.NoError(t, err, "Failed to create CA certificate")

	clientKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err, "Failed to generate client private key")

	clientTemplate := x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Organization: []string{"Test Client"},
			CommonName:   "Test Client",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	clientBytes, err := x509.CreateCertificate(rand.Reader, &clientTemplate, &caTemplate, &clientKey.PublicKey, caKey)
	require.NoError(t, err, "Failed to create client certificate")

	caPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	writeFile(t, caPath, caPEM)

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: clientBytes,
	})
	writeFile(t, certPath, certPEM)

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(clientKey),
	})
	writeFile(t, keyPath, keyPEM)
}

func writeFile(t *testing.T, path string, content []byte) {
	err := os.WriteFile(path, content, filePermission)
	require.NoError(t, err, "Failed to write file: %s", path)
}

func TestCallback_MakeRequest(t *testing.T) {
	cases := []struct {
		desc          string
		serverHandler http.HandlerFunc
		method        string
		contextSetup  func() context.Context
		urls          []string
		permissions   []string
		client        *http.Client
		expectError   bool
		err           error
	}{
		{
			desc: "successful POST request",
			serverHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				w.WriteHeader(http.StatusOK)
			}),
			client:       http.DefaultClient,
			method:       http.MethodPost,
			contextSetup: func() context.Context { return context.Background() },
			permissions:  []string{permission},
			expectError:  false,
		},
		{
			desc: "successful GET request with query params",
			serverHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, domainID, r.URL.Query().Get("domain"))
				assert.Equal(t, userID, r.URL.Query().Get("sender"))
				w.WriteHeader(http.StatusOK)
			}),
			client:       http.DefaultClient,
			method:       http.MethodGet,
			contextSetup: func() context.Context { return context.Background() },
			permissions:  []string{permission},
			expectError:  false,
		},
		{
			desc: "server returns forbidden status",
			serverHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusForbidden)
			}),
			method:       http.MethodPost,
			client:       http.DefaultClient,
			contextSetup: func() context.Context { return context.Background() },
			permissions:  []string{permission},
			expectError:  true,
			err:          svcerr.ErrAuthorization,
		},
		{
			desc:         "invalid URL",
			client:       http.DefaultClient,
			method:       http.MethodGet,
			contextSetup: func() context.Context { return context.Background() },
			urls:         []string{"http://invalid-url"},
			permissions:  []string{permission},
			expectError:  true,
		},
		{
			desc: "cancelled context",
			serverHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
			client: http.DefaultClient,
			method: http.MethodGet,
			contextSetup: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			permissions: []string{permission},
			expectError: true,
		},
		{
			desc: "multiple URLs all succeed",
			serverHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
			client:       http.DefaultClient,
			method:       http.MethodPost,
			contextSetup: func() context.Context { return context.Background() },
			permissions:  []string{permission},
			expectError:  false,
		},
		{
			desc: "use default client when nil is provided",
			serverHandler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
			client:       nil,
			method:       http.MethodPost,
			contextSetup: func() context.Context { return context.Background() },
			permissions:  []string{permission},
			expectError:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			var servers []*httptest.Server
			var urls []string

			if tc.desc == "invalid URL" {
				urls = tc.urls
			} else {
				if tc.desc == "multiple URLs all succeed" {
					for i := 0; i < 2; i++ {
						servers, urls = newServer(tc.serverHandler)
					}
				} else {
					servers, urls = newServer(tc.serverHandler)
				}
			}

			defer func() {
				for _, server := range servers {
					server.Close()
				}
			}()

			cb, err := callout.NewCallback(tc.client, tc.method, urls, tc.permissions)
			assert.NoError(t, err)

			ctx := tc.contextSetup()
			err = cb.Callout(ctx, permission, pl)

			if tc.expectError {
				assert.Error(t, err)
				if tc.err != nil {
					assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func newServer(serverHandler http.HandlerFunc) ([]*httptest.Server, []string) {
	var servers []*httptest.Server
	var urls []string

	server := httptest.NewServer(serverHandler)
	servers = append(servers, server)
	urls = append(urls, server.URL)

	return servers, urls
}

func TestCallback_InvalidMethod(t *testing.T) {
	cases := []struct {
		desc        string
		method      string
		urls        []string
		permissions []string
		err         error
	}{
		{
			desc:        "valid POST method",
			method:      http.MethodPost,
			urls:        []string{"http://example.com"},
			permissions: []string{},
		},
		{
			desc:        "valid GET method",
			method:      http.MethodGet,
			urls:        []string{"http://example.com"},
			permissions: []string{},
		},
		{
			desc:        "invalid method",
			method:      "INVALID-METHOD",
			urls:        []string{"http://example.com"},
			permissions: []string{},
			err:         errors.New("unsupported auth callout method: INVALID-METHOD"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			cb, err := callout.NewCallback(http.DefaultClient, tc.method, tc.urls, tc.permissions)
			assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
			if tc.err == nil {
				assert.NotNil(t, cb)
			}
		})
	}
}

func TestCallback_Permissions(t *testing.T) {
	cases := []struct {
		desc         string
		permissions  []string
		payload      map[string]interface{}
		serverCalled bool
	}{
		{
			desc:        "matching permission is called",
			permissions: []string{permission},
			payload: map[string]interface{}{
				"entity_type": entityType,
				"sender":      userID,
				"domain":      domainID,
				"time":        time.Now().String(),
				"permission":  permission,
			},
			serverCalled: true,
		},
		{
			desc:        "non-matching permission is not called",
			permissions: []string{"other_permission"},
			payload: map[string]interface{}{
				"entity_type": entityType,
				"sender":      userID,
				"domain":      domainID,
				"time":        time.Now().String(),
				"permission":  permission,
			},
			serverCalled: false,
		},
		{
			desc:        "empty permissions list calls always",
			permissions: []string{},
			payload: map[string]interface{}{
				"entity_type": entityType,
				"sender":      userID,
				"domain":      domainID,
				"time":        time.Now().String(),
				"permission":  permission,
			},
			serverCalled: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			serverCalled := false
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				serverCalled = true
				w.WriteHeader(http.StatusOK)
			}))
			defer ts.Close()

			client, err := callout.NewCalloutClient(false, "", "", "", time.Second)
			assert.NoError(t, err)

			cb, err := callout.NewCallback(client, http.MethodPost, []string{ts.URL}, tc.permissions)
			assert.NoError(t, err)

			err = cb.Callout(context.Background(), permission, tc.payload)
			assert.NoError(t, err)
			assert.Equal(t, tc.serverCalled, serverCalled, "Server call status does not match expected")
		})
	}
}

func TestCallback_NoURLs(t *testing.T) {
	client, err := callout.NewCalloutClient(false, "", "", "", time.Second)
	assert.NoError(t, err)

	cb, err := callout.NewCallback(client, http.MethodPost, []string{}, []string{permission})
	assert.NoError(t, err)

	err = cb.Callout(context.Background(), permission, pl)
	assert.NoError(t, err, "No error should be returned when URL list is empty")
}
