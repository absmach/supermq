// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"time"

	"github.com/mainflux/mainflux/commands"
)

type apiReq interface {
	validate() error
}

type createCommandReq struct {
	command   string    `josn:"secret"`
	channel   string    `json: “<chan_id>”`
	CreatedAt time.Time `json:"created_at"`
}

func (req createCommandReq) validate() error {
	if req.command == "" {
		return commands.ErrMalformedEntity
	}

	return nil
}

type viewCommandReq struct {
	Secret string `json:"secret"`
}

func (req viewCommandReq) validate() error {
	if req.Secret == "" {
		return commands.ErrMalformedEntity
	}

	return nil
}

type listCommandReq struct {
	Secret string `json:"secret"`
}

func (req listCommandReq) validate() error {
	if req.Secret == "" {
		return commands.ErrMalformedEntity
	}

	return nil
}

type updateCommandReq struct {
	Secret string `json:"secret"`
}

func (req updateCommandReq) validate() error {
	if req.Secret == "" {
		return commands.ErrMalformedEntity
	}

	return nil
}

type removeCommandReq struct {
	Secret string `json:"secret"`
}

func (req removeCommandReq) validate() error {
	if req.Secret == "" {
		return commands.ErrMalformedEntity
	}

	return nil
}
