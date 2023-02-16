package api

import (
	"github.com/mainflux/mainflux/clients/policies"
	"github.com/mainflux/mainflux/internal/apiutil"
)

type authorizeReq struct {
	Subject    string   `json:"subject,omitempty"`
	Object     string   `json:"object,omitempty"`
	Actions    []string `json:"actions,omitempty"`
	EntityType string   `json:"entity_type,omitempty"`
}

func (req authorizeReq) validate() error {
	for _, a := range req.Actions {
		if ok := policies.ValidateAction(a); !ok {
			return apiutil.ErrMissingPolicyAct
		}
	}
	if req.Subject == "" {
		return apiutil.ErrMissingPolicySub
	}
	if req.Object == "" {
		return apiutil.ErrMissingPolicyObj
	}
	return nil
}

type createPolicyReq struct {
	token   string
	Owner   string `json:"owner,omitempty"`
	thingID string `json:"thing,omitempty"`
	chanID  string `json:"channel,omitempty"`
}

func (req createPolicyReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}

	if req.chanID == "" || req.thingID == "" {
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

type updatePolicyReq struct {
	token   string
	Subject string   `json:"subject,omitempty"`
	Object  string   `json:"object,omitempty"`
	Actions []string `json:"actions,omitempty"`
}

func (req updatePolicyReq) validate() error {
	for _, a := range req.Actions {
		if ok := policies.ValidateAction(a); !ok {
			return apiutil.ErrMissingPolicyAct
		}
	}
	if req.Subject == "" {
		return apiutil.ErrMissingPolicySub
	}
	if req.Object == "" {
		return apiutil.ErrMissingPolicyObj
	}
	return nil
}

type listPolicyReq struct {
	token   string
	Total   uint64
	Offset  uint64
	Limit   uint64
	OwnerID string
	Subject string
	Object  string
	Actions string
}

func (req listPolicyReq) validate() error {
	if req.Actions != "" {
		if ok := policies.ValidateAction(req.Actions); !ok {
			return apiutil.ErrMissingPolicyAct
		}
	}
	return nil
}

type deletePolicyReq struct {
	token   string
	Subject string `json:"subject,omitempty"`
	Object  string `json:"object,omitempty"`
}

func (req deletePolicyReq) validate() error {
	if req.Subject == "" {
		return apiutil.ErrMissingPolicySub
	}
	if req.Object == "" {
		return apiutil.ErrMissingPolicyObj
	}

	return nil
}
