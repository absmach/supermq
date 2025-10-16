// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package callout

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/absmach/supermq/pkg/errors"
	"github.com/absmach/supermq/pkg/server"
)

var errFailedToRead = errors.New("failed to read callout response body")

type Request struct {
	Operation   string         `json:"operation,omitempty"`
	SubjectID   string         `json:"subject_id,omitempty"`
	SubjectType string         `json:"subject_type,omitempty"`
	ObjectType  string         `json:"object_type,omitempty"`
	Payload     map[string]any `json:"payload,omitempty"`
}

// Callout send a request to an external service.
type Callout interface {
	Callout(ctx context.Context, req Request) error
}

type Config struct {
	URLs            []string      `env:"URLS"             envDefault:"" envSeparator:","`
	Method          string        `env:"METHOD"           envDefault:"POST"`
	TLSVerification bool          `env:"TLS_VERIFICATION" envDefault:"true"`
	Timeout         time.Duration `env:"TIMEOUT"          envDefault:"10s"`
	CACert          string        `env:"CA_CERT"          envDefault:""`
	Cert            string        `env:"CERT"             envDefault:""`
	Key             string        `env:"KEY"              envDefault:""`
	Operations      []string      `env:"OPERATIONS"       envDefault:"" envSeparator:","`
}

type callout struct {
	httpClient       *http.Client
	urls             []string
	method           string
	allowedOperation map[string]struct{}
}

// New creates a new instance of Callout.
func New(cfg Config) (Callout, error) {
	if cfg.Method != http.MethodPost && cfg.Method != http.MethodGet {
		return nil, fmt.Errorf("unsupported auth callout method: %s", cfg.Method)
	}

	httpClient, err := newCalloutClient(cfg.TLSVerification, cfg.Cert, cfg.Key, cfg.CACert, cfg.Timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize http client: %w", err)
	}

	allowedOperation := make(map[string]struct{})
	for _, operation := range cfg.Operations {
		allowedOperation[operation] = struct{}{}
	}

	return &callout{
		httpClient:       httpClient,
		urls:             cfg.URLs,
		method:           cfg.Method,
		allowedOperation: allowedOperation,
	}, nil
}

func (c *callout) Callout(ctx context.Context, req Request) error {
	if len(c.urls) == 0 {
		return nil
	}

	if _, exists := c.allowedOperation[req.Operation]; !exists {
		return nil
	}

	// Make requests sequentially as they appear in the URL
	// slice and fail fast as soon as any request fails.
	for _, url := range c.urls {
		if err := c.makeRequest(ctx, url, req); err != nil {
			return err
		}
	}

	return nil
}

func newCalloutClient(skipInsecure bool, certPath, keyPath, caPath string, timeout time.Duration) (*http.Client, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: !skipInsecure,
	}
	if certPath != "" || keyPath != "" {
		clientTLSCert, err := server.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return nil, err
		}

		caCert, err := server.LoadRootCACerts(caPath)
		if err != nil {
			return nil, err
		}
		tlsConfig.RootCAs = caCert
		tlsConfig.Certificates = []tls.Certificate{clientTLSCert}
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: timeout,
	}

	return httpClient, nil
}

func (c *callout) makeRequest(ctx context.Context, urlStr string, req Request) error {
	var r *http.Request
	var err error

	switch c.method {
	case http.MethodGet:
		query := url.Values{}
		query.Set("operation", req.Operation)
		query.Set("subject_id", req.SubjectID)
		query.Set("subject_type", req.SubjectType)
		query.Set("object_type", req.ObjectType)
		for key, value := range req.Payload {
			query.Set(key, fmt.Sprintf("%v", value))
		}
		r, err = http.NewRequestWithContext(ctx, c.method, urlStr+"?"+query.Encode(), nil)
	case http.MethodPost:
		data, jsonErr := json.Marshal(req)
		if jsonErr != nil {
			return jsonErr
		}
		r, err = http.NewRequestWithContext(ctx, c.method, urlStr, bytes.NewReader(data))
		if err == nil {
			r.Header.Set("Content-Type", "application/json")
		}
	}

	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.NewSDKErrorWithStatus(errors.Wrap(errFailedToRead, err), http.StatusInternalServerError)
		}
		return errors.NewSDKErrorWithStatus(errors.New(string(msg)), resp.StatusCode)
	}

	return nil
}
