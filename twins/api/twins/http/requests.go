//
// Copyright (c) 2019
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package http

import (
	"github.com/mainflux/mainflux/things"
	"github.com/mainflux/mainflux/twins"
)

const maxNameSize = 1024
const maxLimitSize = 100

type apiReq interface {
	validate() error
}

type addTwinReq struct {
	token      string
	Name       string                 `json:"name,omitempty"`
	Key        string                 `json:"key,omitempty"`
	ThingID    string                 `json:"thingID"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	State      map[string]interface{} `json:"state,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
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
	token      string
	id         string
	Name       string                 `json:"name,omitempty"`
	Key        string                 `json:"key,omitempty"`
	ThingID    string                 `json:"thingID,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	State      map[string]interface{} `json:"state,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
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

type listReq struct {
	token    string
	limit    uint64
	name     string
	metadata map[string]interface{}
}

func (req *listReq) validate() error {
	if req.token == "" {
		return things.ErrUnauthorizedAccess
	}

	if req.limit == 0 || req.limit > maxLimitSize {
		return things.ErrMalformedEntity
	}

	if len(req.name) > maxNameSize {
		return things.ErrMalformedEntity
	}

	return nil
}

type listByThingReq struct {
	token    string
	limit    uint64
	thing    string
	metadata map[string]interface{}
}

func (req *listByThingReq) validate() error {
	if req.token == "" {
		return things.ErrUnauthorizedAccess
	}

	if req.limit == 0 || req.limit > maxLimitSize {
		return things.ErrMalformedEntity
	}

	if len(req.thing) < 1 {
		return things.ErrMalformedEntity
	}

	return nil
}
