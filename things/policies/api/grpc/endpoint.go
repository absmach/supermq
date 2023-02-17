package grpc

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/mainflux/mainflux/things/clients"
	"github.com/mainflux/mainflux/things/groups"
	"github.com/mainflux/mainflux/things/policies"
)

func canAccessEndpoint(svc policies.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(accessByKeyReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		id, err := svc.CanAccessByKey(ctx, req.chanID, req.thingKey)
		if err != nil {
			return identityRes{}, err
		}
		return identityRes{id: id}, nil
	}
}

func canAccessByIDEndpoint(svc policies.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(accessByIDReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		err := svc.CanAccessByID(ctx, req.chanID, req.thingID)
		return emptyRes{err: err}, err
	}
}

func isChannelOwnerEndpoint(svc groups.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(channelOwnerReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		err := svc.IsChannelOwner(ctx, req.owner, req.chanID)
		return emptyRes{err: err}, err
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
