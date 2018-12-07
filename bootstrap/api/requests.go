package http

import "nov/bootstrap"

type apiReq interface {
	validate() error
}

type addReq struct {
	key         string
	ExternalID  string   `json:"external_id"`
	ExternalKey string   `json:"external_key"`
	Channels    []string `json:"channels"`
	Config      string   `json:"config"`
}

func (req addReq) validate() error {
	if req.ExternalID == "" || req.ExternalKey == "" {
		return bootstrap.ErrMalformedEntity
	}

	return nil
}

type entityReq struct {
	key string
	id  string
}

func (req entityReq) validate() error {
	if req.key == "" {
		return bootstrap.ErrUnauthorizedAccess
	}

	return nil
}

type updateReq struct {
	key        string
	id         string
	MFChannels []string         `json:"channels"`
	Config     string           `json:"config"`
	Status     bootstrap.Status `json:"status"`
}

func (req updateReq) validate() error {
	if req.key == "" {
		return bootstrap.ErrUnauthorizedAccess
	}

	if req.Status != bootstrap.Inactive &&
		req.Status != bootstrap.Active {
		return bootstrap.ErrMalformedEntity
	}

	return nil
}

type listReq struct {
	key    string
	offset uint64
	limit  uint64
}

func (req listReq) validate() error {
	if req.key == "" {
		return bootstrap.ErrUnauthorizedAccess
	}

	return nil
}

type boostrapReq struct {
	key string
	id  string
}

func (req boostrapReq) validate() error {
	if req.key == "" {
		return bootstrap.ErrUnauthorizedAccess
	}

	if req.id == "" {
		return bootstrap.ErrMalformedEntity
	}

	return nil
}

type changeStatusReq struct {
	key    string
	id     string
	Status bootstrap.Status `json:"status"`
}

func (req changeStatusReq) validate() error {
	if req.id == "" || req.key == "" {
		return bootstrap.ErrUnauthorizedAccess
	}

	if req.Status != bootstrap.Created &&
		req.Status != bootstrap.Inactive &&
		req.Status != bootstrap.Active {
		return bootstrap.ErrMalformedEntity
	}

	return nil
}
