// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"time"

	"github.com/absmach/supermq/auth"
	"github.com/absmach/supermq/clients"
	"github.com/absmach/supermq/pkg/authn"
	smqauthz "github.com/absmach/supermq/pkg/authz"
	"github.com/absmach/supermq/pkg/callout"
	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	"github.com/absmach/supermq/pkg/policies"
	"github.com/absmach/supermq/pkg/roles"
	rolemw "github.com/absmach/supermq/pkg/roles/rolemanager/middleware"
	"github.com/absmach/supermq/pkg/svcutil"
)

var (
	errView                    = errors.New("not authorized to view client")
	errUpdate                  = errors.New("not authorized to update client")
	errUpdateTags              = errors.New("not authorized to update client tags")
	errUpdateSecret            = errors.New("not authorized to update client secret")
	errEnable                  = errors.New("not authorized to enable client")
	errDisable                 = errors.New("not authorized to disable client")
	errDelete                  = errors.New("not authorized to delete client")
	errSetParentGroup          = errors.New("not authorized to set parent group to client")
	errRemoveParentGroup       = errors.New("not authorized to remove parent group from client")
	errDomainCreateClients     = errors.New("not authorized to create client in domain")
	errGroupSetChildClients    = errors.New("not authorized to set child client for group")
	errGroupRemoveChildClients = errors.New("not authorized to remove child client for group")
)

var _ clients.Service = (*authorizationMiddleware)(nil)

type authorizationMiddleware struct {
	svc     clients.Service
	repo    clients.Repository
	authz   smqauthz.Authorization
	opp     svcutil.OperationPerm
	extOpp  svcutil.ExternalOperationPerm
	callout callout.Callout
	rolemw.RoleManagerAuthorizationMiddleware
}

// NewAuthorization adds authorization to the clients service.
func NewAuthorization(
	entityType string,
	svc clients.Service,
	authz smqauthz.Authorization,
	repo clients.Repository,
	clientsOpPerm, rolesOpPerm map[svcutil.Operation]svcutil.Permission,
	extOpPerm map[svcutil.ExternalOperation]svcutil.Permission,
	callout callout.Callout,
) (clients.Service, error) {
	opp := clients.NewOperationPerm()
	if err := opp.AddOperationPermissionMap(clientsOpPerm); err != nil {
		return nil, err
	}
	if err := opp.Validate(); err != nil {
		return nil, err
	}
	ram, err := rolemw.NewAuthorization(policies.ClientType, svc, authz, rolesOpPerm, callout)
	if err != nil {
		return nil, err
	}
	extOpp := clients.NewExternalOperationPerm()
	if err := extOpp.AddOperationPermissionMap(extOpPerm); err != nil {
		return nil, err
	}
	if err := extOpp.Validate(); err != nil {
		return nil, err
	}

	return &authorizationMiddleware{
		svc:                                svc,
		authz:                              authz,
		repo:                               repo,
		opp:                                opp,
		extOpp:                             extOpp,
		RoleManagerAuthorizationMiddleware: ram,
		callout:                            callout,
	}, nil
}

func (am *authorizationMiddleware) CreateClients(ctx context.Context, session authn.Session, client ...clients.Client) ([]clients.Client, []roles.RoleProvision, error) {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.ClientsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.CreateOp,
			EntityID:         auth.AnyIDs,
		}); err != nil {
			return []clients.Client{}, []roles.RoleProvision{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}
	if err := am.extAuthorize(ctx, clients.DomainOpCreateClient, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.DomainType,
		Object:      session.DomainID,
	}); err != nil {
		return []clients.Client{}, []roles.RoleProvision{}, errors.Wrap(err, errDomainCreateClients)
	}

	params := map[string]any{
		"entities": client,
		"count":    len(client),
	}

	if err := am.callOut(ctx, session, clients.OpCreateClient.String(clients.OperationNames), "", params); err != nil {
		return []clients.Client{}, []roles.RoleProvision{}, err
	}

	return am.svc.CreateClients(ctx, session, client...)
}

func (am *authorizationMiddleware) View(ctx context.Context, session authn.Session, id string, withRoles bool) (clients.Client, error) {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.ClientsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.ReadOp,
			EntityID:         id,
		}); err != nil {
			return clients.Client{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}

	if err := am.authorize(ctx, clients.OpViewClient, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.ClientType,
		Object:      id,
	}); err != nil {
		return clients.Client{}, errors.Wrap(err, errView)
	}

	if err := am.callOut(ctx, session, clients.OpViewClient.String(clients.OperationNames), id, nil); err != nil {
		return clients.Client{}, err
	}

	return am.svc.View(ctx, session, id, withRoles)
}

func (am *authorizationMiddleware) ListClients(ctx context.Context, session authn.Session, pm clients.Page) (clients.ClientsPage, error) {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.ClientsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.ListOp,
			EntityID:         auth.AnyIDs,
		}); err != nil {
			return clients.ClientsPage{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}

	if err := am.checkSuperAdmin(ctx, session); err == nil {
		session.SuperAdmin = true
	}

	params := map[string]any{
		"pagemeta": pm,
	}

	if err := am.callOut(ctx, session, clients.OpListClients.String(clients.OperationNames), "", params); err != nil {
		return clients.ClientsPage{}, err
	}

	return am.svc.ListClients(ctx, session, pm)
}

func (am *authorizationMiddleware) ListUserClients(ctx context.Context, session authn.Session, userID string, pm clients.Page) (clients.ClientsPage, error) {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.ClientsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.ListOp,
			EntityID:         auth.AnyIDs,
		}); err != nil {
			return clients.ClientsPage{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}

	if err := am.checkSuperAdmin(ctx, session); err != nil {
		return clients.ClientsPage{}, err
	}

	params := map[string]any{
		"user_id":  userID,
		"pagemeta": pm,
	}

	if err := am.callOut(ctx, session, clients.OpListUserClients.String(clients.OperationNames), "", params); err != nil {
		return clients.ClientsPage{}, err
	}

	return am.svc.ListUserClients(ctx, session, userID, pm)
}

func (am *authorizationMiddleware) Update(ctx context.Context, session authn.Session, client clients.Client) (clients.Client, error) {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.ClientsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.UpdateOp,
			EntityID:         client.ID,
		}); err != nil {
			return clients.Client{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}

	if err := am.authorize(ctx, clients.OpUpdateClient, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.ClientType,
		Object:      client.ID,
	}); err != nil {
		return clients.Client{}, errors.Wrap(err, errUpdate)
	}

	if err := am.callOut(ctx, session, clients.OpUpdateClient.String(clients.OperationNames), client.ID, nil); err != nil {
		return clients.Client{}, err
	}

	return am.svc.Update(ctx, session, client)
}

func (am *authorizationMiddleware) UpdateTags(ctx context.Context, session authn.Session, client clients.Client) (clients.Client, error) {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.ClientsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.UpdateOp,
			EntityID:         client.ID,
		}); err != nil {
			return clients.Client{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}

	if err := am.authorize(ctx, clients.OpUpdateClientTags, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.ClientType,
		Object:      client.ID,
	}); err != nil {
		return clients.Client{}, errors.Wrap(err, errUpdateTags)
	}

	if err := am.callOut(ctx, session, clients.OpUpdateClientTags.String(clients.OperationNames), client.ID, nil); err != nil {
		return clients.Client{}, err
	}

	return am.svc.UpdateTags(ctx, session, client)
}

func (am *authorizationMiddleware) UpdateSecret(ctx context.Context, session authn.Session, id, key string) (clients.Client, error) {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.ClientsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.UpdateOp,
			EntityID:         id,
		}); err != nil {
			return clients.Client{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}

	if err := am.authorize(ctx, clients.OpUpdateClientSecret, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.ClientType,
		Object:      id,
	}); err != nil {
		return clients.Client{}, errors.Wrap(err, errUpdateSecret)
	}

	if err := am.callOut(ctx, session, clients.OpUpdateClientSecret.String(clients.OperationNames), id, nil); err != nil {
		return clients.Client{}, err
	}

	return am.svc.UpdateSecret(ctx, session, id, key)
}

func (am *authorizationMiddleware) Enable(ctx context.Context, session authn.Session, id string) (clients.Client, error) {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.ClientsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.UpdateOp,
			EntityID:         id,
		}); err != nil {
			return clients.Client{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}

	if err := am.authorize(ctx, clients.OpEnableClient, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.ClientType,
		Object:      id,
	}); err != nil {
		return clients.Client{}, errors.Wrap(err, errEnable)
	}

	if err := am.callOut(ctx, session, clients.OpEnableClient.String(clients.OperationNames), id, nil); err != nil {
		return clients.Client{}, err
	}

	return am.svc.Enable(ctx, session, id)
}

func (am *authorizationMiddleware) Disable(ctx context.Context, session authn.Session, id string) (clients.Client, error) {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.ClientsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.UpdateOp,
			EntityID:         id,
		}); err != nil {
			return clients.Client{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}

	if err := am.authorize(ctx, clients.OpDisableClient, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.ClientType,
		Object:      id,
	}); err != nil {
		return clients.Client{}, errors.Wrap(err, errDisable)
	}

	if err := am.callOut(ctx, session, clients.OpDisableClient.String(clients.OperationNames), id, nil); err != nil {
		return clients.Client{}, err
	}

	return am.svc.Disable(ctx, session, id)
}

func (am *authorizationMiddleware) Delete(ctx context.Context, session authn.Session, id string) error {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.ClientsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.DeleteOp,
			EntityID:         id,
		}); err != nil {
			return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}
	if err := am.authorize(ctx, clients.OpDeleteClient, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.ClientType,
		Object:      id,
	}); err != nil {
		return errors.Wrap(err, errDelete)
	}

	if err := am.callOut(ctx, session, clients.OpDeleteClient.String(clients.OperationNames), id, nil); err != nil {
		return err
	}

	return am.svc.Delete(ctx, session, id)
}

func (am *authorizationMiddleware) SetParentGroup(ctx context.Context, session authn.Session, parentGroupID string, id string) error {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.ClientsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.UpdateOp,
			EntityID:         id,
		}); err != nil {
			return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}

	if err := am.authorize(ctx, clients.OpSetParentGroup, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.ClientType,
		Object:      id,
	}); err != nil {
		return errors.Wrap(err, errSetParentGroup)
	}

	if err := am.extAuthorize(ctx, clients.GroupOpSetChildClient, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.GroupType,
		Object:      parentGroupID,
	}); err != nil {
		return errors.Wrap(err, errGroupSetChildClients)
	}

	params := map[string]any{
		"parent_id": parentGroupID,
	}

	if err := am.callOut(ctx, session, clients.OpSetParentGroup.String(clients.OperationNames), id, params); err != nil {
		return err
	}

	return am.svc.SetParentGroup(ctx, session, parentGroupID, id)
}

func (am *authorizationMiddleware) RemoveParentGroup(ctx context.Context, session authn.Session, id string) error {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.ClientsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.DeleteOp,
			EntityID:         id,
		}); err != nil {
			return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}

	if err := am.authorize(ctx, clients.OpRemoveParentGroup, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.ClientType,
		Object:      id,
	}); err != nil {
		return errors.Wrap(err, errRemoveParentGroup)
	}

	th, err := am.repo.RetrieveByID(ctx, id)
	if err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}

	if th.ParentGroup != "" {
		if err := am.extAuthorize(ctx, clients.GroupOpSetChildClient, smqauthz.PolicyReq{
			Domain:      session.DomainID,
			SubjectType: policies.UserType,
			Subject:     session.DomainUserID,
			ObjectType:  policies.GroupType,
			Object:      th.ParentGroup,
		}); err != nil {
			return errors.Wrap(err, errGroupRemoveChildClients)
		}

		params := map[string]any{
			"parent_id": th.ParentGroup,
		}

		if err := am.callOut(ctx, session, clients.OpRemoveParentGroup.String(clients.OperationNames), id, params); err != nil {
			return err
		}

		return am.svc.RemoveParentGroup(ctx, session, id)
	}
	return nil
}

func (am *authorizationMiddleware) authorize(ctx context.Context, op svcutil.Operation, req smqauthz.PolicyReq) error {
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

func (am *authorizationMiddleware) extAuthorize(ctx context.Context, extOp svcutil.ExternalOperation, req smqauthz.PolicyReq) error {
	perm, err := am.extOpp.GetPermission(extOp)
	if err != nil {
		return err
	}

	req.Permission = perm.String()

	if err := am.authz.Authorize(ctx, req); err != nil {
		return err
	}

	return nil
}

func (am *authorizationMiddleware) checkSuperAdmin(ctx context.Context, session authn.Session) error {
	if session.Role != authn.AdminRole {
		return svcerr.ErrSuperAdminAction
	}
	if err := am.authz.Authorize(ctx, smqauthz.PolicyReq{
		SubjectType: policies.UserType,
		Subject:     session.UserID,
		Permission:  policies.AdminPermission,
		ObjectType:  policies.PlatformType,
		Object:      policies.SuperMQObject,
	}); err != nil {
		return err
	}
	return nil
}

func (am *authorizationMiddleware) callOut(ctx context.Context, session authn.Session, op, entityID string, pld map[string]any) error {
	req := callout.Request{
		BaseRequest: callout.BaseRequest{
			Operation:  op,
			EntityType: policies.ClientType,
			EntityID:   entityID,
			CallerID:   session.UserID,
			CallerType: policies.UserType,
			DomainID:   session.DomainID,
			Time:       time.Now().UTC(),
		},
		Payload: pld,
	}

	if err := am.callout.Callout(ctx, req); err != nil {
		return err
	}

	return nil
}
