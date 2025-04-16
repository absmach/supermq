// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package authsvc

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/absmach/supermq/pkg/authz"
	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
)

type callback struct {
	httpClient        *http.Client
	urls              []string
	method            string
	allowedPermission map[string]struct{}
}

// NewCallback creates a new instance of CallBack.
func NewCallback(httpClient *http.Client, method string, urls []string, permissions []string) (authz.CallBack, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if method != http.MethodPost && method != http.MethodGet {
		return nil, fmt.Errorf("unsupported auth callback method: %s", method)
	}

	allowedPermission := make(map[string]struct{})
	for _, permission := range permissions {
		allowedPermission[permission] = struct{}{}
	}

	return &callback{
		httpClient:        httpClient,
		urls:              urls,
		method:            method,
		allowedPermission: allowedPermission,
	}, nil
}

func (c *callback) Authorize(ctx context.Context, pr authz.PolicyReq) error {
	if len(c.urls) == 0 {
		return nil
	}

	// Check if the permission is in the allowed list
	// Otherwise, only call webhook if the permission is in the map
	if len(c.allowedPermission) > 0 {
		_, exists := c.allowedPermission[pr.Permission]
		if !exists {
			return nil
		}
	}

	payload := map[string]string{
		"domain":           pr.Domain,
		"subject":          pr.Subject,
		"subject_type":     pr.SubjectType,
		"subject_kind":     pr.SubjectKind,
		"subject_relation": pr.SubjectRelation,
		"object":           pr.Object,
		"object_type":      pr.ObjectType,
		"object_kind":      pr.ObjectKind,
		"relation":         pr.Relation,
		"permission":       pr.Permission,
	}

	var err error
	// We iterate through all URLs in sequence
	// if any request fails, we return the error immediately
	for _, url := range c.urls {
		if err = c.makeRequest(ctx, url, payload); err != nil {
			return err
		}
	}

	return nil
}

func (c *callback) makeRequest(ctx context.Context, urlStr string, params map[string]string) error {
	var req *http.Request
	var err error

	switch c.method {
	case http.MethodGet:
		query := url.Values{}
		for key, value := range params {
			query.Set(key, value)
		}
		req, err = http.NewRequestWithContext(ctx, c.method, urlStr+"?"+query.Encode(), nil)
	case http.MethodPost:
		data, jsonErr := json.Marshal(params)
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
		return errors.NewSDKErrorWithStatus(svcerr.ErrAuthorization, resp.StatusCode)
	}

	return nil
}

func NewCalloutClient(authCalloutTLSVerification bool, authCalloutCert, authCalloutKey, authCalloutCACert string, authCalloutTimeout time.Duration) (*http.Client, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: !authCalloutTLSVerification,
	}
	if authCalloutCert != "" || authCalloutKey != "" {
		clientTLSCert, err := tls.LoadX509KeyPair(authCalloutCert, authCalloutKey)
		if err != nil {
			return nil, err
		}
		certPool, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
		caCert, err := os.ReadFile(authCalloutCACert)
		if err != nil {
			return nil, err
		}
		if !certPool.AppendCertsFromPEM(caCert) {
			return nil, errors.New("failed to append CA certificate")
		}
		tlsConfig.RootCAs = certPool
		tlsConfig.Certificates = []tls.Certificate{clientTLSCert}
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: authCalloutTimeout,
	}

	return httpClient, nil
}
