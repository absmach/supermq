package api

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/mainflux/mainflux/things/clients"
	"github.com/mainflux/mainflux/things/policies"
)

func connectThingEndpoint(svc policies.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		cr := request.(createPolicyReq)

		if err := cr.validate(); err != nil {
			return nil, err
		}
		policy := policies.Policy{
			Subject: cr.ThingID,
			Object:  cr.ChanID,
		}
		if err := svc.AddPolicy(ctx, cr.token, policy); err != nil {
			return nil, err
		}

		return addPolicyRes{created: true}, nil
	}
}

func connectEndpoint(svc policies.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		cr := request.(createPoliciesReq)

		if err := cr.validate(); err != nil {
			return nil, err
		}
		for _, tid := range cr.ThingIDs {
			for _, cid := range cr.ChannelIDs {
				policy := policies.Policy{
					Subject: tid,
					Object:  cid,
				}
				if err := svc.AddPolicy(ctx, cr.token, policy); err != nil {
					return nil, err
				}
			}
		}

		return addPolicyRes{created: true}, nil
	}
}

func disconnectEndpoint(svc policies.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		cr := request.(createPolicyReq)
		if err := cr.validate(); err != nil {
			return nil, err
		}

		policy := policies.Policy{
			Subject: cr.ThingID,
			Object:  cr.ChanID,
		}
		if err := svc.DeletePolicy(ctx, cr.token, policy); err != nil {
			return nil, err
		}

		return deletePolicyRes{}, nil
	}
}

func disconnectThingEndpoint(svc policies.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createPoliciesReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		for _, tid := range req.ThingIDs {
			for _, cid := range req.ChannelIDs {
				policy := policies.Policy{
					Subject: tid,
					Object:  cid,
				}
				if err := svc.DeletePolicy(ctx, req.token, policy); err != nil {
					return nil, err
				}
			}
		}

		return deletePolicyRes{}, nil
	}
}

func identifyEndpoint(svc clients.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(identifyReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		id, err := svc.Identify(ctx, req.Token)
		if err != nil {
			return nil, err
		}

		res := identityRes{
			ID: id,
		}

		return res, nil
	}
}

func canAccessByKeyEndpoint(svc policies.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(canAccessByKeyReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		id, err := svc.CanAccessByKey(ctx, req.chanID, req.Token)
		if err != nil {
			return nil, err
		}

		res := identityRes{
			ID: id,
		}

		return res, nil
	}
}

func canAccessByIDEndpoint(svc policies.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(canAccessByIDReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		if err := svc.CanAccessByID(ctx, req.chanID, req.ThingID); err != nil {
			return nil, err
		}

		res := canAccessByIDRes{}
		return res, nil
	}
}
