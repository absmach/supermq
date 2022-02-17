// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"github.com/mainflux/mainflux/internal/httputil"
)

type createSubReq struct {
	token   string
	Topic   string `json:"topic,omitempty"`
	Contact string `json:"contact,omitempty"`
}

func (req createSubReq) validate() error {
	if req.token == "" {
		return httputil.ErrMissingToken
	}
	if req.Topic == "" {
		return httputil.ErrInvalidTopic
	}
	if req.Contact == "" {
		return httputil.ErrInvalidContact
	}
	return nil
}

type subReq struct {
	token string
	id    string
}

func (req subReq) validate() error {
	if req.token == "" {
		return httputil.ErrMissingToken
	}
	if req.id == "" {
		return httputil.ErrMissingID
	}
	return nil
}

type listSubsReq struct {
	token   string
	topic   string
	contact string
	offset  uint
	limit   uint
}

func (req listSubsReq) validate() error {
	if req.token == "" {
		return httputil.ErrMissingToken
	}
	return nil
}
