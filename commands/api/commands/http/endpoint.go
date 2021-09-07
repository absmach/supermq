// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/mainflux/mainflux/commands"
)

func createCommandsEndpoint(svc commands.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createCommandsReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		greeting, err := svc.CreateCommands(req.Secret)
		if err != nil {
			return nil, err
		}

		res := createCommandsRes{
			Greeting: greeting,
		}
		return res, nil
	}
}

func viewCommandsEndpoint(svc commands.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(viewCommandsReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		greeting, err := svc.ViewCommands(req.Secret)
		if err != nil {
			return nil, err
		}

		res := viewCommandsRes{
			Greeting: greeting,
		}
		return res, nil
	}
}

func listCommandsEndpoint(svc commands.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listCommandsReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		greeting, err := svc.ListCommands(req.Secret)
		if err != nil {
			return nil, err
		}

		res := listCommandsRes{
			Greeting: greeting,
		}
		return res, nil
	}
}

func updateCommandsEndpoint(svc commands.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateCommandsReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		greeting, err := svc.UpdateCommands(req.Secret)
		if err != nil {
			return nil, err
		}

		res := updateCommandsRes{
			Greeting: greeting,
		}
		return res, nil
	}
}

func removeCommandsEndpoint(svc commands.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(removeCommandsReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		greeting, err := svc.RemoveCommands(req.Secret)
		if err != nil {
			return nil, err
		}

		res := removeCommandsRes{
			Greeting: greeting,
		}
		return res, nil
	}
}
