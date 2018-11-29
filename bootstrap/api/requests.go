package http

import "nov/bootstrap"

const maxLimitSize = 100

type apiReq interface {
	validate() error
}

type addReq struct {
	key        string
	ExternalID string `json:"external_id"`
}

func (req addReq) validate() error {
	if req.ExternalID == "" {
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

type boostrapReq struct {
	externalID string
}

func (req boostrapReq) validate() error {
	if req.externalID == "" {
		return bootstrap.ErrUnauthorizedAccess
	}

	return nil
}

type changeStatusReq struct {
	key    string
	ID     string           `json:"id"`
	Status bootstrap.Status `json:"status"`
}

func (req changeStatusReq) validate() error {
	if req.ID == "" || req.key == "" {
		return bootstrap.ErrUnauthorizedAccess
	}

	if req.Status != bootstrap.Created &&
		req.Status != bootstrap.Inactive &&
		req.Status != bootstrap.Active {
		return bootstrap.ErrMalformedEntity
	}

	return nil
}
