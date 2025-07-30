// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"maps"
	"time"

	"github.com/absmach/supermq/auth"
	"github.com/absmach/supermq/domains"
	"github.com/absmach/supermq/pkg/authn"
	"github.com/absmach/supermq/pkg/authz"
	smqauthz "github.com/absmach/supermq/pkg/authz"
	"github.com/absmach/supermq/pkg/callout"
	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	"github.com/absmach/supermq/pkg/policies"
	"github.com/absmach/supermq/pkg/roles"
	rmMW "github.com/absmach/supermq/pkg/roles/rolemanager/middleware"
	"github.com/absmach/supermq/pkg/svcutil"
)

var _ domains.Service = (*authorizationMiddleware)(nil)

// ErrMemberExist indicates that the user is already a member of the domain.
var ErrMemberExist = errors.New("user is already a member of the domain")

type authorizationMiddleware struct {
	svc         domains.Service
	authz       smqauthz.Authorization
	entitiesOps svcutil.EntitiesOperations[svcutil.Operation]
	rOps        svcutil.Operations[svcutil.RoleOperation]
	callout     callout.Callout
	rmMW.RoleManagerAuthorizationMiddleware
}

// AuthorizationMiddleware adds authorization to the clients service.
func AuthorizationMiddleware(entityType string, svc domains.Service, authz smqauthz.Authorization, entitiesOps svcutil.EntitiesOperations[svcutil.Operation], domainRoleOps svcutil.Operations[svcutil.RoleOperation], callout callout.Callout) (domains.Service, error) {

	if err := entitiesOps.Validate(); err != nil {
		return &authorizationMiddleware{}, err
	}

	ram, err := rmMW.NewRoleManagerAuthorizationMiddleware(entityType, svc, authz, domainRoleOps, callout)
	if err != nil {
		return &authorizationMiddleware{}, err
	}
	return &authorizationMiddleware{
		svc:                                svc,
		authz:                              authz,
		entitiesOps:                        entitiesOps,
		callout:                            callout,
		RoleManagerAuthorizationMiddleware: ram,
	}, nil
}

func (am *authorizationMiddleware) CreateDomain(ctx context.Context, session authn.Session, d domains.Domain) (domains.Domain, []roles.RoleProvision, error) {
	params := map[string]any{
		"domain": d,
	}
	if err := am.callOut(ctx, session, policies.DomainType, domains.OpCreateDomain, params); err != nil {
		return domains.Domain{}, nil, err
	}
	return am.svc.CreateDomain(ctx, session, d)
}

func (am *authorizationMiddleware) RetrieveDomain(ctx context.Context, session authn.Session, id string, withRoles bool) (domains.Domain, error) {
	if err := am.checkSuperAdmin(ctx, session.UserID); err == nil {
		session.SuperAdmin = true
		return am.svc.RetrieveDomain(ctx, session, id, withRoles)
	}

	if err := am.authorize(ctx, policies.DomainType, domains.OpRetrieveDomain, authz.PolicyReq{
		Subject:     session.DomainUserID,
		SubjectType: policies.UserType,
		SubjectKind: policies.UsersKind,
		Object:      id,
		ObjectType:  policies.DomainType,
	}); err != nil {
		return domains.Domain{}, err
	}
	params := map[string]any{
		"domain":     id,
		"with_roles": withRoles,
	}
	if err := am.callOut(ctx, session, policies.DomainType, domains.OpRetrieveDomain, params); err != nil {
		return domains.Domain{}, err
	}
	return am.svc.RetrieveDomain(ctx, session, id, withRoles)
}

func (am *authorizationMiddleware) UpdateDomain(ctx context.Context, session authn.Session, id string, d domains.DomainReq) (domains.Domain, error) {
	if err := am.authorize(ctx, policies.DomainType, domains.OpUpdateDomain, authz.PolicyReq{
		Subject:     session.DomainUserID,
		SubjectType: policies.UserType,
		SubjectKind: policies.UsersKind,
		Object:      id,
		ObjectType:  policies.DomainType,
	}); err != nil {
		return domains.Domain{}, err
	}
	params := map[string]any{
		"domain":     id,
		"domain_req": d,
	}
	if err := am.callOut(ctx, session, policies.DomainType, domains.OpUpdateDomain, params); err != nil {
		return domains.Domain{}, err
	}
	return am.svc.UpdateDomain(ctx, session, id, d)
}

func (am *authorizationMiddleware) EnableDomain(ctx context.Context, session authn.Session, id string) (domains.Domain, error) {
	if err := am.authorize(ctx, policies.DomainType, domains.OpEnableDomain, authz.PolicyReq{
		Subject:     session.DomainUserID,
		SubjectType: policies.UserType,
		SubjectKind: policies.UsersKind,
		Object:      id,
		ObjectType:  policies.DomainType,
	}); err != nil {
		return domains.Domain{}, err
	}
	params := map[string]any{
		"domain": id,
	}
	if err := am.callOut(ctx, session, policies.DomainType, domains.OpEnableDomain, params); err != nil {
		return domains.Domain{}, err
	}
	return am.svc.EnableDomain(ctx, session, id)
}

func (am *authorizationMiddleware) DisableDomain(ctx context.Context, session authn.Session, id string) (domains.Domain, error) {
	if err := am.authorize(ctx, policies.DomainType, domains.OpDisableDomain, authz.PolicyReq{
		Subject:     session.DomainUserID,
		SubjectType: policies.UserType,
		SubjectKind: policies.UsersKind,
		Object:      id,
		ObjectType:  policies.DomainType,
	}); err != nil {
		return domains.Domain{}, err
	}
	params := map[string]any{
		"domain": id,
	}
	if err := am.callOut(ctx, session, policies.DomainType, domains.OpDisableDomain, params); err != nil {
		return domains.Domain{}, err
	}
	return am.svc.DisableDomain(ctx, session, id)
}

func (am *authorizationMiddleware) FreezeDomain(ctx context.Context, session authn.Session, id string) (domains.Domain, error) {
	// Only SuperAdmin can freeze the domain
	if err := am.authz.Authorize(ctx, authz.PolicyReq{
		Subject:     session.UserID,
		SubjectType: policies.UserType,
		SubjectKind: policies.UsersKind,
		Permission:  policies.AdminPermission,
		Object:      policies.SuperMQObject,
		ObjectType:  policies.PlatformType,
	}); err != nil {
		return domains.Domain{}, err
	}
	params := map[string]any{
		"domain": id,
	}
	if err := am.callOut(ctx, session, policies.DomainType, domains.OpFreezeDomain, params); err != nil {
		return domains.Domain{}, err
	}
	return am.svc.FreezeDomain(ctx, session, id)
}

func (am *authorizationMiddleware) ListDomains(ctx context.Context, session authn.Session, page domains.Page) (domains.DomainsPage, error) {
	if err := am.checkSuperAdmin(ctx, session.UserID); err == nil {
		session.SuperAdmin = true
	}
	params := map[string]any{
		"page": page,
	}
	if err := am.callOut(ctx, session, policies.DomainType, domains.OpListDomains, params); err != nil {
		return domains.DomainsPage{}, err
	}
	return am.svc.ListDomains(ctx, session, page)
}

func (am *authorizationMiddleware) SendInvitation(ctx context.Context, session authn.Session, invitation domains.Invitation) (err error) {
	domainUserId := auth.EncodeDomainUserID(invitation.DomainID, invitation.InviteeUserID)
	req := authz.PolicyReq{
		SubjectType: policies.UserType,
		SubjectKind: policies.UsersKind,
		Subject:     domainUserId,
		Permission:  policies.MembershipPermission,
		ObjectType:  policies.DomainType,
		Object:      invitation.DomainID,
	}
	if err := am.authz.Authorize(ctx, req); err != nil {
		// return error if the user is already a member of the domain
		return errors.Wrap(svcerr.ErrConflict, ErrMemberExist)
	}

	if err := am.authorize(ctx, policies.DomainType, domains.OpSendDomainInvitation, smqauthz.PolicyReq{
		SubjectType: policies.UserType,
		SubjectKind: policies.UsersKind,
		Subject:     session.DomainUserID,
		ObjectType:  policies.DomainType,
		Object:      session.DomainID,
	}); err != nil {
		return err
	}

	params := map[string]any{
		"invitation": invitation,
		"domain":     invitation.DomainID,
	}
	if err := am.callOut(ctx, session, policies.DomainType, domains.OpSendDomainInvitation, params); err != nil {
		return err
	}

	return am.svc.SendInvitation(ctx, session, invitation)
}

func (am *authorizationMiddleware) DeleteInvitation(ctx context.Context, session authn.Session, inviteeUserID, domainID string) (err error) {
	session.DomainUserID = auth.EncodeDomainUserID(session.DomainID, session.UserID)
	if err := am.authorize(ctx, policies.DomainType, domains.OpDeleteDomainInvitation, smqauthz.PolicyReq{
		SubjectType: policies.UserType,
		SubjectKind: policies.UsersKind,
		Subject:     session.DomainUserID,
		ObjectType:  policies.DomainType,
		Object:      session.DomainID,
	}); err != nil {
		return err
	}

	params := map[string]any{
		"invitee_user_id": inviteeUserID,
		"domain":          domainID,
	}
	if err := am.callOut(ctx, session, policies.DomainType, domains.OpDeleteDomainInvitation, params); err != nil {
		return err
	}

	return am.svc.DeleteInvitation(ctx, session, inviteeUserID, domainID)
}

func (am *authorizationMiddleware) ListDomainInvitations(ctx context.Context, session authn.Session, page domains.InvitationPageMeta) (invs domains.InvitationPage, err error) {
	if err := am.authorize(ctx, policies.DomainType, domains.OpListDomainInvitations, smqauthz.PolicyReq{
		SubjectType: policies.UserType,
		SubjectKind: policies.UsersKind,
		Subject:     session.DomainUserID,
		ObjectType:  policies.DomainType,
		Object:      session.DomainID,
	}); err != nil {
		return domains.InvitationPage{}, err
	}

	params := map[string]any{
		"page": page,
	}
	if err := am.callOut(ctx, session, policies.DomainType, domains.OpListDomainInvitations, params); err != nil {
		return domains.InvitationPage{}, err
	}

	return am.svc.ListDomainInvitations(ctx, session, page)
}

func (am *authorizationMiddleware) ViewDomainInvitation(ctx context.Context, session authn.Session, inviteeUserID, domain string) (invitation domains.Invitation, err error) {
	session.DomainUserID = auth.EncodeDomainUserID(session.DomainID, session.UserID)
	if session.UserID != inviteeUserID {
		if err := am.checkAdmin(ctx, session); err != nil {
			return domains.Invitation{}, err
		}
	}

	params := map[string]any{
		"invitee_user_id": inviteeUserID,
		"domain":          domain,
	}
	if err := am.callOut(ctx, session, policies.DomainType, domains.OpViewInvitation, params); err != nil {
		return domains.Invitation{}, err
	}

	return am.svc.ViewInvitation(ctx, session, inviteeUserID, domain)
}

func (am *authorizationMiddleware) ListInvitations(ctx context.Context, session authn.Session, page domains.InvitationPageMeta) (invs domains.InvitationPage, err error) {
	params := map[string]any{
		"page": page,
	}
	if err := am.callOut(ctx, session, policies.DomainType, domains.OpListInvitations, params); err != nil {
		return domains.InvitationPage{}, err
	}

	return am.svc.ListInvitations(ctx, session, page)
}

func (am *authorizationMiddleware) ViewInvitation(ctx context.Context, session authn.Session, domainID string) (invitation domains.Invitation, err error) {
	session.DomainUserID = auth.EncodeDomainUserID(session.DomainID, session.UserID)

	params := map[string]any{
		"invitee_user_id": session.,
		"domain":          domainID,
	}
	if err := am.callOut(ctx, session, policies.DomainType, domains.OpViewInvitation, params); err != nil {
		return domains.Invitation{}, err
	}

	return am.svc.ViewInvitation(ctx, session, inviteeUserID, domain)
}

func (am *authorizationMiddleware) AcceptInvitation(ctx context.Context, session authn.Session, domainID string) (inv domains.Invitation, err error) {
	params := map[string]any{
		"domain": domainID,
	}
	if err := am.callOut(ctx, session, policies.DomainType, domains.OpAcceptInvitation, params); err != nil {
		return domains.Invitation{}, err
	}
	return am.svc.AcceptInvitation(ctx, session, domainID)
}

func (am *authorizationMiddleware) RejectInvitation(ctx context.Context, session authn.Session, domainID string) (err error) {
	params := map[string]any{
		"domain": domainID,
	}
	if err := am.callOut(ctx, session, policies.DomainType, domains.OpRejectInvitation, params); err != nil {
		return err
	}
	return am.svc.RejectInvitation(ctx, session, domainID)
}

func (am *authorizationMiddleware) authorize(ctx context.Context, entityType string, op svcutil.Operation, authReq authz.PolicyReq) error {
	perm, err := am.entitiesOps.GetPermission(entityType, op)
	if err != nil {
		return err
	}
	authReq.Permission = perm.String()

	if err := am.authz.Authorize(ctx, authReq); err != nil {
		return err
	}

	return nil
}

// checkAdmin checks if the given user is a domain or platform administrator.
func (am *authorizationMiddleware) checkAdmin(ctx context.Context, session authn.Session) error {
	req := smqauthz.PolicyReq{
		SubjectType: policies.UserType,
		SubjectKind: policies.UsersKind,
		Subject:     session.DomainUserID,
		Permission:  policies.AdminPermission,
		ObjectType:  policies.DomainType,
		Object:      session.DomainID,
	}
	if err := am.authz.Authorize(ctx, req); err == nil {
		return nil
	}

	req = smqauthz.PolicyReq{
		SubjectType: policies.UserType,
		SubjectKind: policies.UsersKind,
		Subject:     session.UserID,
		Permission:  policies.AdminPermission,
		ObjectType:  policies.PlatformType,
		Object:      policies.SuperMQObject,
	}

	if err := am.authz.Authorize(ctx, req); err == nil {
		return nil
	}

	return svcerr.ErrAuthorization
}

func (am *authorizationMiddleware) checkSuperAdmin(ctx context.Context, userID string) error {
	if err := am.authz.Authorize(ctx, smqauthz.PolicyReq{
		SubjectType: policies.UserType,
		Subject:     userID,
		Permission:  policies.AdminPermission,
		ObjectType:  policies.PlatformType,
		Object:      policies.SuperMQObject,
	}); err != nil {
		return err
	}
	return nil
}

func (am *authorizationMiddleware) callOut(ctx context.Context, session authn.Session, entityType string, op svcutil.Operation, params map[string]interface{}) error {
	pl := map[string]any{
		"entity_type":  entityType,
		"subject_type": policies.UserType,
		"subject_id":   session.UserID,
		"time":         time.Now().UTC(),
	}

	maps.Copy(params, pl)

	if err := am.callout.Callout(ctx, am.entitiesOps.OperationName(entityType, op), params); err != nil {
		return err
	}

	return nil
}
