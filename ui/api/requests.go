// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"github.com/mainflux/mainflux/things"
	"github.com/mainflux/mainflux/ui"
)

const (
	maxLimitSize = 100
	maxNameSize  = 1024
	nameOrder    = "name"
	idOrder      = "id"
	ascDir       = "asc"
	descDir      = "desc"
)

type indexReq struct {
	token string
}

type createThingsReq struct {
	token    string
	Name     string                 `json:"name,omitempty"`
	Key      string                 `json:"key,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func (req createThingsReq) validate() error {
	if req.token == "" {
		return ui.ErrUnauthorizedAccess
	}

	if len(req.Name) > maxNameSize {
		return ui.ErrMalformedEntity
	}

	return nil
}

type listThingsReq struct {
	token string
}

type viewResourceReq struct {
	token string
	id    string
}

func (req viewResourceReq) validate() error {
	if req.token == "" {
		return things.ErrUnauthorizedAccess
	}

	if req.id == "" {
		return things.ErrMalformedEntity
	}

	return nil
}

type updateThingReq struct {
	token    string
	id       string
	Name     string                 `json:"name,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func (req updateThingReq) validate() error {

	if req.id == "" {
		return things.ErrMalformedEntity
	}

	if len(req.Name) > maxNameSize {
		return things.ErrMalformedEntity
	}

	return nil
}

type createChannelsReq struct {
	token    string
	Name     string                 `json:"name,omitempty"`
	Key      string                 `json:"key,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func (req createChannelsReq) validate() error {
	if req.token == "" {
		return ui.ErrUnauthorizedAccess
	}

	if len(req.Name) > maxNameSize {
		return ui.ErrMalformedEntity
	}

	return nil
}

type updateChannelReq struct {
	token    string
	id       string
	Name     string                 `json:"name,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func (req updateChannelReq) validate() error {

	if req.id == "" {
		return things.ErrMalformedEntity
	}

	if len(req.Name) > maxNameSize {
		return things.ErrMalformedEntity
	}

	return nil
}

type listChannelsReq struct {
	token string
}

type createGroupsReq struct {
	token       string
	ID          string                 `json:"id,omitempty"`
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	ParentID    string                 `json:"parent_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

func (req createGroupsReq) validate() error {
	if req.ID == "" {
		return ui.ErrUnauthorizedAccess
	}

	if len(req.Name) > maxNameSize {
		return ui.ErrMalformedEntity
	}

	return nil
}

type listGroupsReq struct {
	token string
}

type updateGroupReq struct {
	token    string
	id       string
	Name     string                 `json:"name,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func (req updateGroupReq) validate() error {

	if req.token == "" {
		return things.ErrMalformedEntity
	}

	if len(req.Name) > maxNameSize {
		return things.ErrMalformedEntity
	}

	return nil
}

type connectThingReq struct {
	token   string
	ChanID  string `json:"chan_id,omitempty"`
	ThingID string `json:"thing_id,omitempty"`
}

func (req connectThingReq) validate() error {
	// if req.token == "" {
	// 	return things.ErrUnauthorizedAccess
	// }

	if req.ChanID == "" || req.ThingID == "" {
		return things.ErrMalformedEntity
	}

	return nil
}

type connectReq struct {
	token      string
	ChannelIDs []string `json:"channel_ids,omitempty"`
	ThingIDs   []string `json:"thing_ids,omitempty"`
}

func (req connectReq) validate() error {
	// if req.token == "" {
	// 	return things.ErrUnauthorizedAccess
	// }

	if len(req.ChannelIDs) == 0 || len(req.ThingIDs) == 0 {
		return things.ErrMalformedEntity
	}

	for _, chID := range req.ChannelIDs {
		if chID == "" {
			return things.ErrMalformedEntity
		}
	}
	for _, thingID := range req.ThingIDs {
		if thingID == "" {
			return things.ErrMalformedEntity
		}
	}

	return nil
}

type listByConnectionReq struct {
	token        string
	id           string
	pageMetadata things.PageMetadata
}

func (req listByConnectionReq) validate() error {
	if req.token == "" {
		return things.ErrUnauthorizedAccess
	}

	if req.id == "" {
		return things.ErrMalformedEntity
	}

	if req.pageMetadata.Limit == 0 || req.pageMetadata.Limit > maxLimitSize {
		return things.ErrMalformedEntity
	}

	if req.pageMetadata.Order != "" &&
		req.pageMetadata.Order != nameOrder && req.pageMetadata.Order != idOrder {
		return things.ErrMalformedEntity
	}

	if req.pageMetadata.Dir != "" &&
		req.pageMetadata.Dir != ascDir && req.pageMetadata.Dir != descDir {
		return things.ErrMalformedEntity
	}

	return nil
}
