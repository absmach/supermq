// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"time"

	"github.com/absmach/magistrala/pat"
	"github.com/absmach/magistrala/pkg/authn"
	mgauthz "github.com/absmach/magistrala/pkg/authz"
)

var _ pat.Service = (*authorizationMiddleware)(nil)

type authorizationMiddleware struct {
	svc   pat.Service
	authz mgauthz.Authorization
}

// AuthorizationMiddleware adds authorization to the clients service.
func AuthorizationMiddleware(entityType string, svc pat.Service, authz mgauthz.Authorization) (pat.Service, error) {
	return &authorizationMiddleware{
		svc:   svc,
		authz: authz,
	}, nil
}

func (am *authorizationMiddleware) CreatePAT(ctx context.Context, session authn.Session, name, description string, duration time.Duration, scope pat.Scope) (pat.PAT, error) {
	return am.svc.CreatePAT(ctx, session, name, description, duration, scope)
}

func (am *authorizationMiddleware) UpdatePATName(ctx context.Context, session authn.Session, patID, name string) (pat.PAT, error) {
	return am.svc.UpdatePATName(ctx, session, patID, name)
}

func (am *authorizationMiddleware) UpdatePATDescription(ctx context.Context, session authn.Session, patID, description string) (pat.PAT, error) {
	return am.svc.UpdatePATDescription(ctx, session, patID, description)
}

func (am *authorizationMiddleware) RetrievePAT(ctx context.Context, session authn.Session, patID string) (pat.PAT, error) {
	return am.svc.RetrievePAT(ctx, session, patID)
}

func (am *authorizationMiddleware) ListPATS(ctx context.Context, session authn.Session, pm pat.PATSPageMeta) (pat.PATSPage, error) {
	return am.svc.ListPATS(ctx, session, pm)
}

func (am *authorizationMiddleware) DeletePAT(ctx context.Context, session authn.Session, patID string) error {
	return am.svc.DeletePAT(ctx, session, patID)
}

func (am *authorizationMiddleware) ResetPATSecret(ctx context.Context, session authn.Session, patID string, duration time.Duration) (pat.PAT, error) {
	return am.svc.ResetPATSecret(ctx, session, patID, duration)
}

func (am *authorizationMiddleware) RevokePATSecret(ctx context.Context, session authn.Session, patID string) error {
	return am.svc.RevokePATSecret(ctx, session, patID)
}

func (am *authorizationMiddleware) AddPATScopeEntry(ctx context.Context, session authn.Session, patID string, platformEntityType pat.PlatformEntityType, optionalDomainID string, optionalDomainEntityType pat.DomainEntityType, operation pat.OperationType, entityIDs ...string) (pat.Scope, error) {
	return am.svc.AddPATScopeEntry(ctx, session, patID, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
}

func (am *authorizationMiddleware) RemovePATScopeEntry(ctx context.Context, session authn.Session, patID string, platformEntityType pat.PlatformEntityType, optionalDomainID string, optionalDomainEntityType pat.DomainEntityType, operation pat.OperationType, entityIDs ...string) (pat.Scope, error) {
	return am.svc.RemovePATScopeEntry(ctx, session, patID, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
}

func (am *authorizationMiddleware) ClearPATAllScopeEntry(ctx context.Context, session authn.Session, patID string) error {
	return am.svc.ClearPATAllScopeEntry(ctx, session, patID)
}

func (am *authorizationMiddleware) IdentifyPAT(ctx context.Context, secret string) (pat.PAT, error) {
	return am.svc.IdentifyPAT(ctx, secret)
}

func (am *authorizationMiddleware) AuthorizePAT(ctx context.Context, paToken string, platformEntityType pat.PlatformEntityType, optionalDomainID string, optionalDomainEntityType pat.DomainEntityType, operation pat.OperationType, entityIDs ...string) error {
	return am.svc.AuthorizePAT(ctx, paToken, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
}

func (am *authorizationMiddleware) CheckPAT(ctx context.Context, userID, patID string, platformEntityType pat.PlatformEntityType, optionalDomainID string, optionalDomainEntityType pat.DomainEntityType, operation pat.OperationType, entityIDs ...string) error {
	return am.svc.CheckPAT(ctx, userID, patID, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
}
