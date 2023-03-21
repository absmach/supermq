package api

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/mainflux/mainflux/homing"
)

func saveEndpoint(svc homing.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(telemetryReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		if err := svc.Save(ctx, req.Telemetry, req.ServiceName); err != nil {
			return nil, err
		}
		res := saveTelemetryRes{
			created: true,
		}
		return res, nil
	}
}

func getAllEndpoint(svc homing.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(listTelemetryReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		pm := homing.PageMetadata{
			Offset: req.offset,
			Limit:  req.limit,
		}
		tm, err := svc.GetAll(ctx, req.token, pm)
		if err != nil {
			return nil, err
		}
		res := telemetryPageRes{
			pageRes: pageRes{
				Total:  tm.Total,
				Offset: tm.Offset,
				Limit:  tm.Limit,
			},
			Telemetry: tm.Telemetry,
		}
		return res, nil
	}
}
