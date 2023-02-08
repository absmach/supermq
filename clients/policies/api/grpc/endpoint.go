package grpc

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/mainflux/mainflux/clients/clients"
	"github.com/mainflux/mainflux/clients/policies"
)

func authorizeEndpoint(svc policies.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(authReq)

		if err := req.validate(); err != nil {
			return authorizeRes{}, err
		}

		err := svc.Authorize(ctx, req.EntityType, policies.Policy{Subject: req.Sub, Object: req.Obj, Actions: []string{req.Act}})
		if err != nil {
			return authorizeRes{}, err
		}
		return authorizeRes{authorized: true}, err
	}
}

func issueEndpoint(svc clients.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(issueReq)
		if err := req.validate(); err != nil {
			return issueRes{}, err
		}

		tkn, err := svc.IssueToken(ctx, req.email, "")
		if err != nil {
			return issueRes{}, err
		}

		return issueRes{value: tkn.AccessToken}, nil
	}
}

func identifyEndpoint(svc clients.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(identityReq)
		if err := req.validate(); err != nil {
			return identityRes{}, err
		}

		ir, err := svc.Identify(ctx, req.token)
		if err != nil {
			return identityRes{}, err
		}

		ret := identityRes{
			id:    ir.ID,
			email: ir.Email,
		}
		return ret, nil
	}
}

func addPolicyEndpoint(svc policies.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(addPolicyReq)
		if err := req.validate(); err != nil {
			return addPolicyRes{}, err
		}

		err := svc.AddPolicy(ctx, req.Token, policies.Policy{Subject: req.Sub, Object: req.Obj, Actions: req.Act})
		if err != nil {
			return addPolicyRes{}, err
		}
		return addPolicyRes{authorized: true}, err
	}
}

func deletePolicyEndpoint(svc policies.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(policyReq)
		if err := req.validate(); err != nil {
			return deletePolicyRes{}, err
		}

		err := svc.DeletePolicy(ctx, "", policies.Policy{Subject: req.Sub, Object: req.Obj, Actions: []string{req.Act}})
		if err != nil {
			return deletePolicyRes{}, err
		}
		return deletePolicyRes{deleted: true}, nil
	}
}

func listPoliciesEndpoint(svc policies.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listPoliciesReq)

		page, err := svc.ListPolicy(ctx, "", policies.Page{Subject: req.Sub, Object: req.Obj, Action: req.Act})
		if err != nil {
			return deletePolicyRes{}, err
		}
		var objects []string
		for _, p := range page.Policies {
			objects = append(objects, p.Object)
		}
		return listPoliciesRes{objects: objects}, nil
	}
}
