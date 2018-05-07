package grpc

import (
	"github.com/asaskevich/govalidator"
	"github.com/mainflux/mainflux/users"
)

type accessReq struct {
	clientKey string
	chanID    string
}

func (req accessReq) validate() error {
	if !govalidator.IsUUID(req.chanID) || req.clientKey == "" {
		return users.ErrMalformedEntity
	}
	return nil
}
