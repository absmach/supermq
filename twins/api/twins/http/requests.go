//
// Copyright (c) 2019
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package http

import (
	"github.com/mainflux/mainflux/twins"
)

const maxNameSize = 1024

const maxNameSize = 1024

type apiReq interface {
	validate() error
}

type pingReq struct {
	Secret string
}

func (req pingReq) validate() error {
	if req.Secret == "" {
		return twins.ErrUnauthorizedAccess
	}

	return nil
}

type addTwinReq struct {
	token    string
	Name     string                 `json:"name,omitempty"`
	Key      string                 `key:"key,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func (req addTwinReq) validate() error {
	if req.token == "" {
		return twins.ErrUnauthorizedAccess
	}

	if len(req.Name) > maxNameSize {
		return twins.ErrMalformedEntity
	}

	return nil
}
