// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"time"

	"github.com/mainflux/mainflux/auth"
)

type issueKeyReq struct {
	issuer   string
	Secret   string        `json:"secret,omitempty"`
	Type     uint32        `json:"type,omitempty"`
	Duration time.Duration `json:"duration,omitempty"`
}

// It is not possible to issue Reset key using HTTP API.
func (req issueKeyReq) validate() error {
	if req.issuer == "" || (req.Type != auth.SessionKey && req.Type != auth.UserKey) {
		return auth.ErrMalformedEntity
	}
	return nil
}

type keyReq struct {
	issuer string
	id     string
}

func (req keyReq) validate() error {
	if req.issuer == "" || req.id == "" {
		return auth.ErrMalformedEntity
	}
	return nil
}
