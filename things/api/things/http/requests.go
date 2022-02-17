// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"github.com/gofrs/uuid"
	"github.com/mainflux/mainflux/internal/httputil"
	"github.com/mainflux/mainflux/things"
)

const (
	maxLimitSize = 100
	maxNameSize  = 1024
	nameOrder    = "name"
	idOrder      = "id"
	ascDir       = "asc"
	descDir      = "desc"
	readPolicy   = "read"
	writePolicy  = "write"
	deletePolicy = "delete"
)

type createThingReq struct {
	token    string
	Name     string                 `json:"name,omitempty"`
	Key      string                 `json:"key,omitempty"`
	ID       string                 `json:"id,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func validateUUID(extID string) (err error) {
	id, err := uuid.FromString(extID)
	if id.String() != extID || err != nil {
		return httputil.ErrInvalidIDFormat
	}

	return nil
}

func (req createThingReq) validate() error {
	if req.token == "" {
		return httputil.ErrMissingToken
	}

	if err := validateUUID(req.ID); req.ID != "" && err != nil {
		return err
	}

	if len(req.Name) > maxNameSize {
		return httputil.ErrNameSize
	}

	return nil
}

type createThingsReq struct {
	token  string
	Things []createThingReq
}

func (req createThingsReq) validate() error {
	if req.token == "" {
		return httputil.ErrMissingToken
	}

	if len(req.Things) <= 0 {
		return httputil.ErrEmptyList
	}

	for _, thing := range req.Things {
		if err := validateUUID(thing.ID); thing.ID != "" && err != nil {
			return err
		}

		if len(thing.Name) > maxNameSize {
			return httputil.ErrNameSize
		}
	}

	return nil
}

type shareThingReq struct {
	token    string
	thingID  string
	UserIDs  []string `json:"user_ids"`
	Policies []string `json:"policies"`
}

func (req shareThingReq) validate() error {
	if req.token == "" {
		return httputil.ErrMissingToken
	}

	if req.thingID == "" || len(req.UserIDs) == 0 {
		return httputil.ErrMissingID
	}

	if len(req.Policies) == 0 {
		return httputil.ErrEmptyList
	}

	for _, p := range req.Policies {
		if p != readPolicy && p != writePolicy && p != deletePolicy {
			return httputil.ErrMalformedPolicy
		}
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
	if req.token == "" {
		return httputil.ErrMissingToken
	}

	if req.id == "" {
		return httputil.ErrMissingID
	}

	if len(req.Name) > maxNameSize {
		return httputil.ErrNameSize
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
		return httputil.ErrMissingToken
	}

	if req.id == "" {
		return httputil.ErrMissingID
	}

	if req.Key == "" {
		return httputil.ErrMissingKey
	}

	return nil
}

type createChannelReq struct {
	token    string
	Name     string                 `json:"name,omitempty"`
	ID       string                 `json:"id,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func (req createChannelReq) validate() error {
	if req.token == "" {
		return httputil.ErrMissingToken
	}

	if err := validateUUID(req.ID); req.ID != "" && err != nil {
		return err
	}

	if len(req.Name) > maxNameSize {
		return httputil.ErrNameSize
	}

	return nil
}

type createChannelsReq struct {
	token    string
	Channels []createChannelReq
}

func (req createChannelsReq) validate() error {
	if req.token == "" {
		return httputil.ErrMissingToken
	}

	if len(req.Channels) <= 0 {
		return httputil.ErrEmptyList
	}

	for _, channel := range req.Channels {
		if err := validateUUID(channel.ID); channel.ID != "" && err != nil {
			return err
		}

		if len(channel.Name) > maxNameSize {
			return httputil.ErrNameSize
		}
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
	if req.token == "" {
		return httputil.ErrMissingToken
	}

	if req.id == "" {
		return httputil.ErrMissingID
	}

	if len(req.Name) > maxNameSize {
		return httputil.ErrNameSize
	}

	return nil
}

type viewResourceReq struct {
	token string
	id    string
}

func (req viewResourceReq) validate() error {
	if req.token == "" {
		return httputil.ErrMissingToken
	}

	if req.id == "" {
		return httputil.ErrMissingID
	}

	return nil
}

type listResourcesReq struct {
	token        string
	pageMetadata things.PageMetadata
}

func (req *listResourcesReq) validate() error {
	if req.token == "" {
		return httputil.ErrMissingToken
	}

	if req.pageMetadata.Limit > maxLimitSize || req.pageMetadata.Limit < 1 {
		return httputil.ErrLimitSize
	}

	if len(req.pageMetadata.Name) > maxNameSize {
		return httputil.ErrNameSize
	}

	if req.pageMetadata.Order != "" &&
		req.pageMetadata.Order != nameOrder && req.pageMetadata.Order != idOrder {
		return httputil.ErrInvalidOrder
	}

	if req.pageMetadata.Dir != "" &&
		req.pageMetadata.Dir != ascDir && req.pageMetadata.Dir != descDir {
		return httputil.ErrInvalidDirection
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
		return httputil.ErrMissingToken
	}

	if req.id == "" {
		return httputil.ErrMissingID
	}

	if req.pageMetadata.Limit > maxLimitSize || req.pageMetadata.Limit < 1 {
		return httputil.ErrLimitSize
	}

	if req.pageMetadata.Order != "" &&
		req.pageMetadata.Order != nameOrder && req.pageMetadata.Order != idOrder {
		return httputil.ErrInvalidOrder
	}

	if req.pageMetadata.Dir != "" &&
		req.pageMetadata.Dir != ascDir && req.pageMetadata.Dir != descDir {
		return httputil.ErrInvalidDirection
	}

	return nil
}

type connectThingReq struct {
	token   string
	chanID  string
	thingID string
}

func (req connectThingReq) validate() error {
	if req.token == "" {
		return httputil.ErrMissingToken
	}

	if req.chanID == "" || req.thingID == "" {
		return httputil.ErrMissingID
	}

	return nil
}

type connectReq struct {
	token      string
	ChannelIDs []string `json:"channel_ids,omitempty"`
	ThingIDs   []string `json:"thing_ids,omitempty"`
}

func (req connectReq) validate() error {
	if req.token == "" {
		return httputil.ErrMissingToken
	}

	if len(req.ChannelIDs) == 0 || len(req.ThingIDs) == 0 {
		return httputil.ErrEmptyList
	}

	for _, chID := range req.ChannelIDs {
		if chID == "" {
			return httputil.ErrMissingID
		}
	}
	for _, thingID := range req.ThingIDs {
		if thingID == "" {
			return httputil.ErrMissingID
		}
	}

	return nil
}

type listThingsGroupReq struct {
	token        string
	groupID      string
	pageMetadata things.PageMetadata
}

func (req listThingsGroupReq) validate() error {
	if req.token == "" {
		return httputil.ErrMissingToken
	}

	if req.groupID == "" {
		return httputil.ErrMissingID
	}

	if req.pageMetadata.Limit > maxLimitSize || req.pageMetadata.Limit < 1 {
		return httputil.ErrLimitSize
	}

	if len(req.pageMetadata.Name) > maxNameSize {
		return httputil.ErrNameSize
	}

	if req.pageMetadata.Order != "" &&
		req.pageMetadata.Order != nameOrder && req.pageMetadata.Order != idOrder {
		return httputil.ErrInvalidOrder
	}

	if req.pageMetadata.Dir != "" &&
		req.pageMetadata.Dir != ascDir && req.pageMetadata.Dir != descDir {
		return httputil.ErrInvalidDirection
	}

	return nil

}
