package api

import (
	provSDK "github.com/mainflux/mainflux/provision/sdk"
)

type addThingReq struct {
	token       string
	ExternalID  string `json:"externalid"`
	ExternalKey string `json:"externalkey"`
}

func (req addThingReq) validate() error {
	if req.ExternalID == "" || req.ExternalKey == "" {
		return provSDK.ErrMalformedEntity
	}
	return nil
}
