// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package sdk

import (
	"time"

	"github.com/mainflux/mainflux/auth"
)

type assignRequest struct {
	Type    string   `json:"type,omitempty"`
	Members []string `json:"members"`
}

// UserPasswordReq contains old and new passwords
type UserPasswordReq struct {
	OldPassword string `json:"old_password,omitempty"`
	Password    string `json:"password,omitempty"`
}

// ConnectionIDs contains ID lists of things and channels to be connected
type ConnectionIDs struct {
	ChannelIDs []string `json:"channel_ids"`
	ThingIDs   []string `json:"thing_ids"`
}

type issueKeyReq struct {
	token    string
	Type     uint32        `json:"type,omitempty"`
	Duration time.Duration `json:"duration,omitempty"`
}

// It is not possible to issue Reset key using HTTP API.
func (req issueKeyReq) validate() error {
	if req.Type == auth.LoginKey {
		return nil
	}
	if req.token == "" || (req.Type != auth.APIKey) {
		return auth.ErrMalformedEntity
	}
	return nil
}

type keyReq struct {
	token string
	id    string
}

func (req keyReq) validate() error {
	if req.token == "" || req.id == "" {
		return auth.ErrMalformedEntity
	}
	return nil
}
