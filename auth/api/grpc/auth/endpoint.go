// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"

	"github.com/absmach/supermq/auth"
	"github.com/absmach/supermq/pkg/policies"
	"github.com/go-kit/kit/endpoint"
)

func authenticateEndpoint(svc auth.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(authenticateReq)
		if err := req.validate(); err != nil {
			return authenticateRes{}, err
		}

		key, err := svc.Identify(ctx, req.token)
		if err != nil {
			return authenticateRes{}, err
		}

		tokenType := auth.AccessTokenType
		if key.Type == auth.PersonalAccessToken {
			tokenType = auth.PersonalAccessTokenType
		}

		return authenticateRes{id: key.ID, userID: key.Subject, userRole: key.Role, verified: key.Verified, tokenType: tokenType}, nil
	}
}

func authorizeEndpoint(svc auth.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		req := request.(authReq)

		if err := req.validate(); err != nil {
			return authorizeRes{}, err
		}

		err := svc.Authorize(ctx, policies.Policy{
			TokenType:        req.TokenType,
			Domain:           req.Domain,
			SubjectType:      req.SubjectType,
			SubjectKind:      req.SubjectKind,
			Subject:          req.Subject,
			Relation:         req.Relation,
			Permission:       req.Permission,
			ObjectType:       req.ObjectType,
			Object:           req.Object,
			UserID:           req.UserID,
			PatID:            req.PatID,
			EntityType:       uint32(req.EntityType),
			OptionalDomainID: req.OptionalDomainID,
			Operation:        uint32(req.Operation),
			EntityID:         req.EntityID,
		})
		if err != nil {
			return authorizeRes{authorized: false}, err
		}
		return authorizeRes{authorized: true}, nil
	}
}
