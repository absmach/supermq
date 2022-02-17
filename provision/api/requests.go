package api

import (
	"github.com/mainflux/mainflux/internal/httputil"
)

type provisionReq struct {
	token       string
	Name        string `json:"name"`
	ExternalID  string `json:"external_id"`
	ExternalKey string `json:"external_key"`
}

func (req provisionReq) validate() error {
	if req.ExternalID == "" {
		return httputil.ErrMissingID
	}

	if req.ExternalKey == "" {
		return httputil.ErrMissingKey
	}

	return nil
}

type mappingReq struct {
	token string
}

func (req mappingReq) validate() error {
	if req.token == "" {
		return httputil.ErrMissingToken
	}
	return nil
}
