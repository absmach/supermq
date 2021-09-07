// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package http

import "github.com/mainflux/mainflux/commands"

type apiReq interface {
	validate() error
}

type createCommandsReq struct {
	Secret string `json:"secret"`
}

func (req createCommandsReq) validate() error {
	if req.Secret == "" {
		return commands.ErrMalformedEntity
	}

	return nil
}

type viewCommandsReq struct {
	Secret string `json:"secret"`
}

func (req viewCommandsReq) validate() error {
	if req.Secret == "" {
		return commands.ErrMalformedEntity
	}

	return nil
}

type listCommandsReq struct {
	Secret string `json:"secret"`
}

func (req listCommandsReq) validate() error {
	if req.Secret == "" {
		return commands.ErrMalformedEntity
	}

	return nil
}

type updateCommandsReq struct {
	Secret string `json:"secret"`
}

func (req updateCommandsReq) validate() error {
	if req.Secret == "" {
		return commands.ErrMalformedEntity
	}

	return nil
}

type removeCommandsReq struct {
	Secret string `json:"secret"`
}

func (req removeCommandsReq) validate() error {
	if req.Secret == "" {
		return commands.ErrMalformedEntity
	}

	return nil
}
