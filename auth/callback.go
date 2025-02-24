// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	"github.com/absmach/supermq/pkg/policies"
)

type callback struct {
	httpClient *http.Client
	urls       []string
	method     string
}

// CallBack send auth request to an external service.
//
//go:generate mockery --name CallBack --output=./mocks --filename callback.go --quiet --note "Copyright (c) Abstract Machines"
type CallBack interface {
	Authorize(ctx context.Context, pr policies.Policy) error
}

// NewCallback creates a new instance of CallBack.
func NewCallback(httpClient *http.Client, method string, urls []string) (CallBack, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if method != http.MethodPost && method != http.MethodGet {
		return nil, fmt.Errorf("unsupported auth callback method: %s", method)
	}

	return &callback{
		httpClient: httpClient,
		urls:       urls,
		method:     method,
	}, nil
}

func (c *callback) Authorize(ctx context.Context, pr policies.Policy) error {
	if len(c.urls) == 0 {
		return nil
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
	for i := range c.urls {
		if err = c.makeRequest(ctx, c.method, c.urls[i], payload); err == nil {
			return nil
		}
	}

	return err
}

func (c *callback) makeRequest(ctx context.Context, method, urlStr string, params map[string]string) error {
	var req *http.Request
	var err error

	switch method {
	case http.MethodGet:
		query := url.Values{}
		for key, value := range params {
			query.Set(key, value)
		}
		req, err = http.NewRequestWithContext(ctx, method, urlStr+"?"+query.Encode(), nil)
	case http.MethodPost:
		data, jsonErr := json.Marshal(params)
		if jsonErr != nil {
			return jsonErr
		}
		req, err = http.NewRequestWithContext(ctx, method, urlStr, bytes.NewReader(data))
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
		return errors.Wrap(svcerr.ErrAuthorization, fmt.Errorf("status code %d", resp.StatusCode))
	}

	return nil
}
