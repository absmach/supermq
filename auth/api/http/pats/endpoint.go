// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package pats

import (
	"context"

	"github.com/absmach/supermq/auth"
	"github.com/go-kit/kit/endpoint"
)

func createPATEndpoint(svc auth.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createPatReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		pat, err := svc.CreatePAT(ctx, req.token, req.Name, req.Description, req.Duration)
		if err != nil {
			return nil, err
		}

		return createPatRes{pat}, nil
	}
}

func retrievePATEndpoint(svc auth.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(retrievePatReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		pat, err := svc.RetrievePAT(ctx, req.token, req.id)
		if err != nil {
			return nil, err
		}

		return retrievePatRes{pat}, nil
	}
}

func updatePATNameEndpoint(svc auth.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updatePatNameReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		pat, err := svc.UpdatePATName(ctx, req.token, req.id, req.Name)
		if err != nil {
			return nil, err
		}

		return updatePatNameRes{pat}, nil
	}
}

func updatePATDescriptionEndpoint(svc auth.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updatePatDescriptionReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		pat, err := svc.UpdatePATDescription(ctx, req.token, req.id, req.Description)
		if err != nil {
			return nil, err
		}

		return updatePatDescriptionRes{pat}, nil
	}
}

func listPATSEndpoint(svc auth.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listPatsReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		pm := auth.PATSPageMeta{
			Limit:  req.limit,
			Offset: req.offset,
		}
		patsPage, err := svc.ListPATS(ctx, req.token, pm)
		if err != nil {
			return nil, err
		}

		return listPatsRes{patsPage}, nil
	}
}

func deletePATEndpoint(svc auth.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(deletePatReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		if err := svc.DeletePAT(ctx, req.token, req.id); err != nil {
			return nil, err
		}

		return deletePatRes{}, nil
	}
}

func resetPATSecretEndpoint(svc auth.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(resetPatSecretReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		pat, err := svc.ResetPATSecret(ctx, req.token, req.id, req.Duration)
		if err != nil {
			return nil, err
		}

		return resetPatSecretRes{pat}, nil
	}
}

func revokePATSecretEndpoint(svc auth.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(revokePatSecretReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		if err := svc.RevokePATSecret(ctx, req.token, req.id); err != nil {
			return nil, err
		}

		return revokePatSecretRes{}, nil
	}
}

func addScopeEntryEndpoint(svc auth.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(scopeEntryReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		err := svc.AddScopeEntry(ctx, req.token, req.id, req.Scopes)
		if err != nil {
			return nil, err
		}

		return scopeEntryRes{}, nil
	}
}

func removeScopeEntryEndpoint(svc auth.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(scopeEntryReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		err := svc.RemoveScopeEntry(ctx, req.token, req.id, req.Scopes)
		if err != nil {
			return nil, err
		}
		return scopeEntryRes{}, nil
	}
}

func clearAllScopeEntryEndpoint(svc auth.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(clearAllScopeEntryReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		if err := svc.ClearAllScopeEntry(ctx, req.token, req.id); err != nil {
			return nil, err
		}

		return clearAllScopeEntryRes{}, nil
	}
}

func listScopesEndpoint(svc auth.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listScopesReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		pm := auth.ScopesPageMeta{
			Limit:  req.limit,
			Offset: req.offset,
			PatID:  req.patID,
		}
		scopesPage, err := svc.ListScopes(ctx, req.token, pm)
		if err != nil {
			return nil, err
		}

		return listScopeRes{scopesPage}, nil
	}
}
