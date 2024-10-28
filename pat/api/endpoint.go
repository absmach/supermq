// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package pats

import (
	"context"

	"github.com/absmach/magistrala/internal/api"
	"github.com/absmach/magistrala/pat"
	"github.com/absmach/magistrala/pkg/authn"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	"github.com/go-kit/kit/endpoint"
)

func createPATEndpoint(svc pat.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createPatReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthentication
		}

		pat, err := svc.CreatePAT(ctx, session, req.Name, req.Description, req.Duration, req.Scope)
		if err != nil {
			return nil, err
		}

		return createPatRes{pat}, nil
	}
}

func retrievePATEndpoint(svc pat.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(retrievePatReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthentication
		}

		pat, err := svc.RetrievePAT(ctx, session, req.id)
		if err != nil {
			return nil, err
		}

		return retrievePatRes{pat}, nil
	}
}

func updatePATNameEndpoint(svc pat.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updatePatNameReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthentication
		}
		pat, err := svc.UpdatePATName(ctx, session, req.id, req.Name)
		if err != nil {
			return nil, err
		}

		return updatePatNameRes{pat}, nil
	}
}

func updatePATDescriptionEndpoint(svc pat.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updatePatDescriptionReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthentication
		}

		pat, err := svc.UpdatePATDescription(ctx, session, req.id, req.Description)
		if err != nil {
			return nil, err
		}

		return updatePatDescriptionRes{pat}, nil
	}
}

func listPATSEndpoint(svc pat.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listPatsReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthentication
		}

		pm := pat.PATSPageMeta{
			Limit:  req.limit,
			Offset: req.offset,
		}
		patsPage, err := svc.ListPATS(ctx, session, pm)
		if err != nil {
			return nil, err
		}

		return listPatsRes{patsPage}, nil
	}
}

func deletePATEndpoint(svc pat.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(deletePatReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthentication
		}

		if err := svc.DeletePAT(ctx, session, req.id); err != nil {
			return nil, err
		}

		return deletePatRes{}, nil
	}
}

func resetPATSecretEndpoint(svc pat.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(resetPatSecretReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthentication
		}

		pat, err := svc.ResetPATSecret(ctx, session, req.id, req.Duration)
		if err != nil {
			return nil, err
		}

		return resetPatSecretRes{pat}, nil
	}
}

func revokePATSecretEndpoint(svc pat.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(revokePatSecretReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthentication
		}

		if err := svc.RevokePATSecret(ctx, session, req.id); err != nil {
			return nil, err
		}

		return revokePatSecretRes{}, nil
	}
}

func addPATScopeEntryEndpoint(svc pat.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(addPatScopeEntryReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthentication
		}

		scope, err := svc.AddPATScopeEntry(ctx, session, req.id, req.PlatformEntityType, req.OptionalDomainID, req.OptionalDomainEntityType, req.Operation, req.EntityIDs...)
		if err != nil {
			return nil, err
		}

		return addPatScopeEntryRes{scope}, nil
	}
}

func removePATScopeEntryEndpoint(svc pat.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(removePatScopeEntryReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthentication
		}

		scope, err := svc.RemovePATScopeEntry(ctx, session, req.id, req.PlatformEntityType, req.OptionalDomainID, req.OptionalDomainEntityType, req.Operation, req.EntityIDs...)
		if err != nil {
			return nil, err
		}
		return removePatScopeEntryRes{scope}, nil
	}
}

func clearPATAllScopeEntryEndpoint(svc pat.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(clearAllScopeEntryReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		session, ok := ctx.Value(api.SessionKey).(authn.Session)
		if !ok {
			return nil, svcerr.ErrAuthentication
		}

		if err := svc.ClearPATAllScopeEntry(ctx, session, req.id); err != nil {
			return nil, err
		}

		return clearAllScopeEntryRes{}, nil
	}
}

func authorizePATEndpoint(svc pat.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(authorizePATReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		if err := svc.AuthorizePAT(ctx, req.token, req.PlatformEntityType, req.OptionalDomainID, req.OptionalDomainEntityType, req.Operation, req.EntityIDs...); err != nil {
			return nil, err
		}

		return authorizePATRes{}, nil
	}
}
