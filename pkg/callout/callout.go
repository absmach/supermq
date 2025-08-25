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
	"maps"
	"net/http"
	"net/url"
	"time"

	"github.com/absmach/supermq/pkg/errors"
	"github.com/absmach/supermq/pkg/server"
)

var errFailedToRead = errors.New("failed to read callout response body")

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

type CallOutReq struct {
	Operation   string         `json:"operation"`
	SubjectID   string         `json:"subject_id"`
	SubjectType string         `json:"subject_type"`
	Payload     map[string]any `json:"payload"`
}

// Callout send request to an external service.
type Callout interface {
	Callout(ctx context.Context, perm string, pl map[string]any) error
}

// New creates a new instance of Callout.
func New(cfg Config) (Callout, error) {
	httpClient, err := newCalloutClient(cfg.TLSVerification, cfg.Cert, cfg.Key, cfg.CACert, cfg.Timeout)
	if err != nil {
		return nil, fmt.Errorf("failied to initialize http client: %w", err)
	}

	if cfg.Method != http.MethodPost && cfg.Method != http.MethodGet {
		return nil, fmt.Errorf("unsupported auth callout method: %s", cfg.Method)
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

func newCalloutClient(ctls bool, certPath, keyPath, caPath string, timeout time.Duration) (*http.Client, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: !ctls,
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

func (c *callout) makeRequest(ctx context.Context, urlStr string, params map[string]any) error {
	var req *http.Request
	var err error

	switch c.method {
	case http.MethodGet:
		query := url.Values{}
		for key, value := range params {
			query.Set(key, fmt.Sprintf("%v", value))
		}
		req, err = http.NewRequestWithContext(ctx, c.method, urlStr+"?"+query.Encode(), nil)
	case http.MethodPost:
		payload := make(map[string]any)
		maps.Copy(payload, params)
		operation, _ := params["operation"].(string)
		subjectID, _ := params["subject_id"].(string)
		subjectType, _ := params["subject_type"].(string)

		delete(payload, "subject_id")
		delete(payload, "subject_type")
		delete(payload, "operation")

		calloutReq := CallOutReq{
			Operation:   operation,
			SubjectID:   subjectID,
			SubjectType: subjectType,
			Payload:     payload,
		}

		data, jsonErr := json.Marshal(calloutReq)
		if jsonErr != nil {
			return jsonErr
		}
		req, err = http.NewRequestWithContext(ctx, c.method, urlStr, bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
	}

	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
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

func (c *callout) Callout(ctx context.Context, op string, pl map[string]any) error {
	if len(c.urls) == 0 {
		return nil
	}

	// Check if the operation is in the allowed list
	// Otherwise, only call webhook if the operation is in the map
	if _, exists := c.allowedOperation[op]; !exists {
		return nil
	}

	pl["operation"] = op

	// We iterate through all URLs in sequence
	// if any request fails, we return the error immediately
	for _, url := range c.urls {
		if err := c.makeRequest(ctx, url, pl); err != nil {
			return err
		}
	}

	return nil
}
