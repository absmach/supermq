//
// Copyright (c) 2019
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package http

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/mainflux/mainflux/twins"
)

func pingEndpoint(svc twins.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(pingReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		greeting, err := svc.Ping(req.Secret)
		if err != nil {
			return nil, err
		}

		res := pingRes{
			Greeting: greeting,
		}
		return res, nil
	}
}

func addTwinEndpoint(svc twins.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(addTwinReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		twin := twins.Twin{
			Key:      req.Key,
			Name:     req.Name,
			Metadata: req.Metadata,
		}
		saved, err := svc.AddTwin(ctx, req.token, twin)
		if err != nil {
			return nil, err
		}

		res := twinRes{
			id:      saved.ID,
			created: true,
		}
		return res, nil
	}
}

func updateTwinEndpoint(svc twins.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateTwinReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		twin := twins.Twin{
			ID:       req.id,
			Name:     req.Name,
			Metadata: req.Metadata,
		}

		if err := svc.UpdateTwin(ctx, req.token, twin); err != nil {
			return nil, err
		}

		res := twinRes{id: req.id, created: false}
		return res, nil
	}
}

func updateKeyEndpoint(svc twins.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateKeyReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		if err := svc.UpdateKey(ctx, req.token, req.id, req.Key); err != nil {
			return nil, err
		}

		res := twinRes{id: req.id, created: false}
		return res, nil
	}
}
