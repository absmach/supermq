package api

import (
	"github.com/mainflux/mainflux/internal/api"
	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/things/groups"
)

type createGroupReq struct {
	groups.Group
	token string
}

func (req createGroupReq) validate() error {
	if len(req.Name) > api.MaxNameSize || req.Name == "" {
		return apiutil.ErrNameSize
	}
	if len(req.Name) > api.MaxNameSize {
		return apiutil.ErrNameSize
	}

	// Do the validation only if request contains ID
	if req.ID != "" {
		return api.ValidateUUID(req.ID)
	}
	return nil
}

type createGroupsReq struct {
	token  string
	Groups []groups.Group
}

func (req createGroupsReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}

	if len(req.Groups) <= 0 {
		return apiutil.ErrEmptyList
	}

	for _, channel := range req.Groups {
		if channel.ID != "" {
			if err := api.ValidateUUID(channel.ID); err != nil {
				return err
			}
		}
		if len(channel.Name) > api.MaxNameSize {
			return apiutil.ErrNameSize
		}
	}

	return nil
}

type updateGroupReq struct {
	token       string
	id          string
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

func (req updateGroupReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}

	if req.id == "" {
		return apiutil.ErrMissingID
	}

	if len(req.Name) > api.MaxNameSize {
		return apiutil.ErrNameSize
	}
	return nil
}

type listGroupsReq struct {
	groups.GroupsPage
	token string
	// - `true`  - result is JSON tree representing groups hierarchy,
	// - `false` - result is JSON array of groups.
	tree bool
}

func (req listGroupsReq) validate() error {
	if req.Level < groups.MinLevel || req.Level > groups.MaxLevel {
		return apiutil.ErrInvalidLevel
	}

	return nil
}

type listMembershipReq struct {
	groups.GroupsPage
	token    string
	clientID string
}

func (req listMembershipReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}

	if req.clientID == "" {
		return apiutil.ErrMissingID
	}

	if req.Limit > api.MaxLimitSize || req.Limit < 1 {
		return apiutil.ErrLimitSize
	}

	if len(req.Name) > api.MaxNameSize {
		return apiutil.ErrNameSize
	}

	return nil
}

type groupReq struct {
	token string
	id    string
}

func (req groupReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.id == "" {
		return apiutil.ErrMissingID
	}

	return nil
}

type changeGroupStatusReq struct {
	token string
	id    string
}

func (req changeGroupStatusReq) validate() error {
	if req.id == "" {
		return apiutil.ErrMissingID
	}
	return nil
}
