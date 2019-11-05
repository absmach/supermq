// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package grpc

import "github.com/mainflux/mainflux/auth"

type identityReq struct {
	token string
}

func (req identityReq) validate() error {
	if req.token == "" {
		return auth.ErrMalformedEntity
	}
	return nil
}

type issueReq struct {
	issuer  string
	keyType uint32
}

func (req issueReq) validate() error {
	if req.issuer == "" {
		return auth.ErrUnauthorizedAccess
	}
	if req.keyType != auth.SessionKey &&
		req.keyType != auth.UserKey &&
		req.keyType != auth.ResetKey {
		return auth.ErrMalformedEntity
	}
	return nil
}
