package api

import (
	"github.com/mainflux/mainflux/internal/api"
	"github.com/mainflux/mainflux/internal/apiutil"
)

type createPolicyReq struct {
	token    string
	Owner    string   `json:"owner,omitempty"`
	ClientID string   `json:"client,omitempty"`
	GroupID  string   `json:"group,omitempty"`
	Actions  []string `json:"actions,omitempty"`
}

func (req createPolicyReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.GroupID == "" || req.ClientID == "" {
		return apiutil.ErrMissingID
	}
	return nil
}

type createPoliciesReq struct {
	token     string
	Owner     string   `json:"owner,omitempty"`
	ClientIDs []string `json:"client_ids,omitempty"`
	GroupIDs  []string `json:"group_ids,omitempty"`
	Actions   []string `json:"actions,omitempty"`
}

func (req createPoliciesReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}

	if len(req.GroupIDs) == 0 || len(req.ClientIDs) == 0 {
		return apiutil.ErrEmptyList
	}

	for _, chID := range req.GroupIDs {
		if chID == "" {
			return apiutil.ErrMissingID
		}
	}
	for _, thingID := range req.ClientIDs {
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

type authorizeReq struct {
	ClientSecret string `json:"secret"`
	GroupID      string `json:"group_id"`
	Action       string `json:"action"`
	EntityType   string `json:"entity_type"`
}

func (req authorizeReq) validate() error {
	if req.GroupID == "" {
		return apiutil.ErrMissingID
	}
	if req.ClientSecret == "" {
		return apiutil.ErrMissingSecret
	}

	return nil
}

type policyReq struct {
	token    string
	Owner    string `json:"owner,omitempty"`
	ClientID string `json:"client,omitempty"`
	GroupID  string `json:"group,omitempty"`
	Action   string `json:"action,omitempty"`
}

func (req policyReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	return nil
}

type listPoliciesReq struct {
	token  string
	offset uint64
	limit  uint64
	client string
	group  string
	action string
	owner  string
}

func (req listPoliciesReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.limit > api.MaxLimitSize || req.limit < 1 {
		return apiutil.ErrLimitSize
	}

	return nil
}
