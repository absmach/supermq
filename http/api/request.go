// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	apiutil "github.com/absmach/supermq/api/http/util"
	"github.com/absmach/supermq/pkg/messaging"
)

type publishReq struct {
	msg   *messaging.Message
	token string
}

func (req publishReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerKey
	}
	if len(req.msg.Payload) == 0 {
		return apiutil.ErrEmptyMessage
	}

	return nil
}

type connReq struct {
	username  string
	password  string
	channelID string
	domainID  string
	subtopic  string
}

type healthCheckReq struct {
	domain string
	token  string
}

func (req healthCheckReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerKey
	}
	if req.domain == "" {
		return apiutil.ErrMissingDomainID
	}

	return nil
}
