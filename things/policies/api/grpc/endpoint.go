package grpc

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/mainflux/mainflux/things/clients"
	"github.com/mainflux/mainflux/things/policies"
)

func authorizeEndpoint(svc policies.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(authorizeReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		ar := policies.AccessRequest{
			Subject: req.clientID,
			Object:  req.groupID,
			Action:  req.action,
		}
		thindID, err := svc.Authorize(ctx, ar, req.entityType)
		if err != nil {
			return authorizeRes{authorized: false}, err
		}

		return authorizeRes{authorized: true, thingID: thindID}, nil
	}
}

func identifyEndpoint(svc clients.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(identifyReq)
		id, err := svc.Identify(ctx, req.key)
		if err := req.validate(); err != nil {
			return nil, err
		}
		if err != nil {
			return identityRes{}, err
		}
		return identityRes{id: id}, nil
	}
}
