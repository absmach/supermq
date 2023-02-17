package api

import (
	"github.com/mainflux/mainflux/internal/apiutil"
)

type createPolicyReq struct {
	token   string
	Owner   string `json:"owner,omitempty"`
	ThingID string `json:"thing,omitempty"`
	ChanID  string `json:"channel,omitempty"`
}

func (req createPolicyReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}

	if req.ChanID == "" || req.ThingID == "" {
		return apiutil.ErrMissingID
	}
	return nil
}

type createPoliciesReq struct {
	token      string
	Owner      string   `json:"owner,omitempty"`
	ThingIDs   []string `json:"thing_ids,omitempty"`
	ChannelIDs []string `json:"channel_ids,omitempty"`
}

func (req createPoliciesReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}

	if len(req.ChannelIDs) == 0 || len(req.ThingIDs) == 0 {
		return apiutil.ErrEmptyList
	}

	for _, chID := range req.ChannelIDs {
		if chID == "" {
			return apiutil.ErrMissingID
		}
	}
	for _, thingID := range req.ThingIDs {
		if thingID == "" {
			return apiutil.ErrMissingID
		}
	}
	return nil
}

type identifyReq struct {
	Token string `json:"token"`
}

func (req identifyReq) validate() error {
	if req.Token == "" {
		return apiutil.ErrBearerKey
	}

	return nil
}

type canAccessByKeyReq struct {
	chanID string
	Token  string `json:"token"`
}

func (req canAccessByKeyReq) validate() error {
	if req.Token == "" {
		return apiutil.ErrBearerKey
	}

	if req.chanID == "" {
		return apiutil.ErrMissingID
	}

	return nil
}

type canAccessByIDReq struct {
	chanID  string
	ThingID string `json:"thing_id"`
}

func (req canAccessByIDReq) validate() error {
	if req.ThingID == "" || req.chanID == "" {
		return apiutil.ErrMissingID
	}

	return nil
}
