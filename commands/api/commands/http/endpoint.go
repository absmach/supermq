// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/mainflux/mainflux/commands"
)

func createCommandEndpoint(svc commands.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createCommandReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		command, err := svc.CreateCommand(req.command)
		if err != nil {
			return nil, err
		}

		res := createCommandRes{
			command: command,
		}
		return res, nil
	}
}

func viewCommandEndpoint(svc commands.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(viewCommandReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		greeting, err := svc.ViewCommand(req.Secret)
		if err != nil {
			return nil, err
		}

		res := viewCommandRes{
			Greeting: greeting,
		}
		return res, nil
	}
}

func listCommandEndpoint(svc commands.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listCommandReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		greeting, err := svc.ListCommand(req.Secret)
		if err != nil {
			return nil, err
		}

		res := listCommandRes{
			Greeting: greeting,
		}
		return res, nil
	}
}

func updateCommandEndpoint(svc commands.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateCommandReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		greeting, err := svc.UpdateCommand(req.Secret)
		if err != nil {
			return nil, err
		}

		res := updateCommandRes{
			Greeting: greeting,
		}
		return res, nil
	}
}

func removeCommandEndpoint(svc commands.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(removeCommandReq)

		if err := req.validate(); err != nil {
			return nil, err
		}

		greeting, err := svc.RemoveCommand(req.Secret)
		if err != nil {
			return nil, err
		}

		res := removeCommandRes{
			Greeting: greeting,
		}
		return res, nil
	}
}
