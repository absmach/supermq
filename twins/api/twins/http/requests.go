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

type apiReq interface {
	validate() error
}

type addTwinReq struct {
	token    string
	Name     string                 `json:"name,omitempty"`
	Key      string                 `json:"key,omitempty"`
	ThingID  string                 `json:"thingID"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func (req addTwinReq) validate() error {
	if req.token == "" {
		return twins.ErrUnauthorizedAccess
	}

	if req.ThingID == "" {
		return twins.ErrMalformedEntity
	}

	if len(req.Name) > maxNameSize {
		return twins.ErrMalformedEntity
	}

	return nil
}

type updateTwinReq struct {
	token    string
	id       string
	Name     string                 `json:"name,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func (req updateTwinReq) validate() error {
	if req.token == "" {
		return twins.ErrUnauthorizedAccess
	}

	if req.id == "" {
		return twins.ErrMalformedEntity
	}

	if len(req.Name) > maxNameSize {
		return twins.ErrMalformedEntity
	}

	return nil
}

type updateKeyReq struct {
	token string
	id    string
	Key   string `json:"key"`
}

func (req updateKeyReq) validate() error {
	if req.token == "" {
		return twins.ErrUnauthorizedAccess
	}

	if req.id == "" || req.Key == "" {
		return twins.ErrMalformedEntity
	}

	return nil
}

type viewTwinReq struct {
	token string
	id    string
}

func (req viewTwinReq) validate() error {
	if req.token == "" {
		return twins.ErrUnauthorizedAccess
	}

	if req.id == "" {
		return twins.ErrMalformedEntity
	}

	return nil
}
