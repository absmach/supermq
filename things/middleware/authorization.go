// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"

	"github.com/absmach/magistrala/pkg/authn"
	"github.com/absmach/magistrala/pkg/authz"
	mgauthz "github.com/absmach/magistrala/pkg/authz"
	"github.com/absmach/magistrala/pkg/clients"
	mgclients "github.com/absmach/magistrala/pkg/clients"
	"github.com/absmach/magistrala/pkg/policies"
	rmMW "github.com/absmach/magistrala/pkg/roles/rolemanager/middleware"
	"github.com/absmach/magistrala/pkg/svcutil"
	"github.com/absmach/magistrala/things"
)

var _ things.Service = (*authorizationMiddleware)(nil)

type authorizationMiddleware struct {
	svc   things.Service
	authz mgauthz.Authorization
	opp   svcutil.OperationPerm
	rmMW.RoleManagerAuthorizationMiddleware
}

// AuthorizationMiddleware adds authorization to the clients service.
func AuthorizationMiddleware(entityType string, svc things.Service, authz mgauthz.Authorization, thingsOpPerm, rolesOpPerm map[svcutil.Operation]svcutil.Permission) (things.Service, error) {
	opp := things.NewOperationPerm()
	if err := opp.AddOperationPermissionMap(thingsOpPerm); err != nil {
		return nil, err
	}
	if err := opp.Validate(); err != nil {
		return nil, err
	}
	ram, err := rmMW.NewRoleManagerAuthorizationMiddleware(policies.ThingType, svc, authz, rolesOpPerm)
	if err != nil {
		return nil, err
	}

	return &authorizationMiddleware{
		svc:                                svc,
		authz:                              authz,
		opp:                                opp,
		RoleManagerAuthorizationMiddleware: ram,
	}, nil
}

func (am *authorizationMiddleware) CreateThings(ctx context.Context, session authn.Session, client ...clients.Client) ([]clients.Client, error) {
	if err := am.authorize(ctx, things.OpCreateThing, authz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.DomainType,
		Object:      session.DomainID,
	}); err != nil {
		return []clients.Client{}, err
	}

	return am.svc.CreateThings(ctx, session, client...)
}

func (am *authorizationMiddleware) ViewClient(ctx context.Context, session authn.Session, id string) (clients.Client, error) {
	if err := am.authorize(ctx, things.OpViewThing, authz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.ThingType,
		Object:      id,
	}); err != nil {
		return clients.Client{}, err
	}
	return am.svc.ViewClient(ctx, session, id)
}

func (am *authorizationMiddleware) ListClients(ctx context.Context, session authn.Session, reqUserID string, pm clients.Page) (clients.ClientsPage, error) {
	if err := am.checkSuperAdmin(ctx, session.UserID); err != nil {
		session.SuperAdmin = true
	}

	return am.svc.ListClients(ctx, session, reqUserID, pm)
}

func (am *authorizationMiddleware) UpdateClient(ctx context.Context, session authn.Session, client clients.Client) (clients.Client, error) {
	if err := am.authorize(ctx, things.OpUpdateThing, authz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.ThingType,
		Object:      client.ID,
	}); err != nil {
		return clients.Client{}, err
	}

	return am.svc.UpdateClient(ctx, session, client)
}

func (am *authorizationMiddleware) UpdateClientTags(ctx context.Context, session authn.Session, client clients.Client) (clients.Client, error) {
	if err := am.authorize(ctx, things.OpUpdateClientTags, authz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.ThingType,
		Object:      client.ID,
	}); err != nil {
		return clients.Client{}, err
	}

	return am.svc.UpdateClientTags(ctx, session, client)
}

func (am *authorizationMiddleware) UpdateClientSecret(ctx context.Context, session authn.Session, id, key string) (clients.Client, error) {
	if err := am.authorize(ctx, things.OpUpdateClientSecret, authz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.ThingType,
		Object:      id,
	}); err != nil {
		return clients.Client{}, err
	}
	return am.svc.UpdateClientSecret(ctx, session, id, key)
}

func (am *authorizationMiddleware) EnableClient(ctx context.Context, session authn.Session, id string) (clients.Client, error) {
	if err := am.authorize(ctx, things.OpEnableThing, authz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.ThingType,
		Object:      id,
	}); err != nil {
		return clients.Client{}, err
	}

	return am.svc.EnableClient(ctx, session, id)
}

func (am *authorizationMiddleware) DisableClient(ctx context.Context, session authn.Session, id string) (clients.Client, error) {
	if err := am.authorize(ctx, things.OpDisableThing, authz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.ThingType,
		Object:      id,
	}); err != nil {
		return clients.Client{}, err
	}
	return am.svc.DisableClient(ctx, session, id)
}

func (am *authorizationMiddleware) Identify(ctx context.Context, key string) (string, error) {
	return am.svc.Identify(ctx, key)
}

func (am *authorizationMiddleware) Authorize(ctx context.Context, req things.AuthzReq) (string, error) {
	return am.svc.Authorize(ctx, req)
}

func (am *authorizationMiddleware) DeleteClient(ctx context.Context, session authn.Session, id string) error {
	if err := am.authorize(ctx, things.OpDeleteThing, authz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.ThingType,
		Object:      id,
	}); err != nil {
		return err
	}

	return am.svc.DeleteClient(ctx, session, id)
}

func (am *authorizationMiddleware) authorize(ctx context.Context, op svcutil.Operation, req authz.PolicyReq) error {
	perm, err := am.opp.GetPermission(op)
	if err != nil {
		return err
	}

	req.Permission = perm.String()

	if err := am.authz.Authorize(ctx, req); err != nil {
		return err
	}

	return nil
}

func (am *authorizationMiddleware) checkSuperAdmin(ctx context.Context, userID string) error {
	if err := am.authz.Authorize(ctx, mgauthz.PolicyReq{
		SubjectType: policies.UserType,
		Subject:     userID,
		Permission:  policies.AdminPermission,
		ObjectType:  policies.PlatformType,
		Object:      policies.MagistralaObject,
	}); err != nil {
		return err
	}
	return nil
}

func (am *authorizationMiddleware) RetrieveById(ctx context.Context, id string) (mgclients.Client, error) {
	return am.svc.RetrieveById(ctx, id)
}

func (am *authorizationMiddleware) RetrieveByIds(ctx context.Context, ids []string) (mgclients.ClientsPage, error) {
	return am.svc.RetrieveByIds(ctx, ids)
}

func (am *authorizationMiddleware) AddConnections(ctx context.Context, conns []things.Connection) (err error) {
	return am.svc.AddConnections(ctx, conns)
}
func (am *authorizationMiddleware) RemoveConnections(ctx context.Context, conns []things.Connection) (err error) {
	return am.svc.RemoveConnections(ctx, conns)
}
