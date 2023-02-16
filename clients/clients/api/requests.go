package api

import (
	"github.com/mainflux/mainflux/clients/clients"
	"github.com/mainflux/mainflux/internal/api"
	"github.com/mainflux/mainflux/internal/apiutil"
)

const maxLimitSize = 100

type createClientReq struct {
	client clients.Client
	token  string
}

func (req createClientReq) validate() error {
	if len(req.client.Name) > api.MaxNameSize {
		return apiutil.ErrNameSize
	}
	return nil
}

type viewClientReq struct {
	token string
	id    string
}

func (req viewClientReq) validate() error {
	return nil
}

type listClientsReq struct {
	token      string
	status     clients.Status
	offset     uint64
	limit      uint64
	name       string
	tag        string
	owner      string
	sharedBy   string
	visibility string
	metadata   clients.Metadata
}

func (req listClientsReq) validate() error {
	if req.limit > maxLimitSize || req.limit < 1 {
		return apiutil.ErrLimitSize
	}
	if req.visibility != "" &&
		req.visibility != api.AllVisibility &&
		req.visibility != api.MyVisibility &&
		req.visibility != api.SharedVisibility {
		return apiutil.ErrInvalidVisibilityType
	}
	return nil
}

type listMembersReq struct {
	clients.Page
	token   string
	groupID string
}

func (req listMembersReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}

	if req.groupID == "" {
		return apiutil.ErrMissingID
	}

	return nil
}

type updateClientReq struct {
	token    string
	id       string
	Name     string                 `json:"name,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Tags     []string               `json:"tags,omitempty"`
}

func (req updateClientReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}

	if req.id == "" {
		return apiutil.ErrMissingID
	}

	return nil
}

type updateClientTagsReq struct {
	id    string
	token string
	Tags  []string `json:"tags,omitempty"`
}

func (req updateClientTagsReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}

	if req.id == "" {
		return apiutil.ErrMissingID
	}
	return nil
}

type updateClientOwnerReq struct {
	id    string
	token string
	Owner string `json:"owner,omitempty"`
}

func (req updateClientOwnerReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	if req.id == "" {
		return apiutil.ErrMissingID
	}
	if req.Owner == "" {
		return apiutil.ErrMissingOwner
	}
	return nil
}

type updateClientCredentialsReq struct {
	token     string
	id        string
	Identity  string `json:"identity,omitempty"`
	OldSecret string `json:"old_secret,omitempty"`
	NewSecret string `json:"new_secret,omitempty"`
}

func (req updateClientCredentialsReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}
	return nil
}

type changeClientStatusReq struct {
	token string
	id    string
}

func (req changeClientStatusReq) validate() error {
	if req.id == "" {
		return apiutil.ErrMissingID
	}
	return nil
}

type loginClientReq struct {
	Identity string `json:"identity,omitempty"`
	Secret   string `json:"secret,omitempty"`
}

func (req loginClientReq) validate() error {
	if req.Identity == "" {
		return apiutil.ErrMissingIdentity
	}
	if req.Secret == "" {
		return apiutil.ErrMissingSecret
	}
	return nil
}

type tokenReq struct {
	RefreshToken string `json:"refresh_token,omitempty"`
}

func (req tokenReq) validate() error {
	if req.RefreshToken == "" {
		return apiutil.ErrBearerToken
	}
	return nil
}
