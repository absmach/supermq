// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/mainflux/mainflux/commands"
)

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
