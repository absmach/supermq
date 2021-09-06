// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package http

import "github.com/mainflux/mainflux/commands"

type apiReq interface {
	validate() error
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
