package api

import (
	"github.com/mainflux/mainflux/homing"
	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/pkg/errors"
)

const maxLimitSize = 100

type telemetryReq struct {
	homing.Telemetry
	ServiceName string `json:"service"`
}

func (req telemetryReq) validate() error {
	if req.ServiceName == "" {
		return errors.ErrMalformedEntity
	}

	if req.IpAddress == "" {
		return errors.ErrMalformedEntity
	}
	if req.Version == "" {
		return errors.ErrMalformedEntity
	}

	return nil
}

type listTelemetryReq struct {
	token     string
	offset    uint64
	limit     uint64
	IpAddress string `json:"ip_address"`
}

func (req listTelemetryReq) validate() error {
	if req.token == "" {
		return apiutil.ErrBearerToken
	}

	if req.limit > maxLimitSize || req.limit < 1 {
		return apiutil.ErrLimitSize
	}

	return nil
}
