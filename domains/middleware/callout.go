// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"time"

	"github.com/absmach/supermq/domains"
	"github.com/absmach/supermq/pkg/authn"
	"github.com/absmach/supermq/pkg/callout"
	"github.com/absmach/supermq/pkg/policies"
	"github.com/absmach/supermq/pkg/roles"
	rolemw "github.com/absmach/supermq/pkg/roles/rolemanager/middleware"
)

var _ domains.Service = (*calloutMiddleware)(nil)

type calloutMiddleware struct {
	svc     domains.Service
	callout callout.Callout
	rolemw.RoleManagerCalloutMiddleware
}

func NewCallout(svc domains.Service, callout callout.Callout) (domains.Service, error) {
	call, err := rolemw.NewCallout(policies.ClientType, svc, callout)
	if err != nil {
		return nil, err
	}

	return &calloutMiddleware{
		svc:                          svc,
		callout:                      callout,
		RoleManagerCalloutMiddleware: call,
	}, nil
}

func (cm *calloutMiddleware) CreateDomain(ctx context.Context, session authn.Session, d domains.Domain) (domains.Domain, []roles.RoleProvision, error) {
	if err := cm.callOut(ctx, session, domains.OpCreateDomain.String(domains.OperationNames), d.ID, nil); err != nil {
		return domains.Domain{}, nil, err
	}

	return cm.svc.CreateDomain(ctx, session, d)
}

func (cm *calloutMiddleware) RetrieveDomain(ctx context.Context, session authn.Session, id string, withRoles bool) (domains.Domain, error) {
	params := map[string]any{
		"with_roles": withRoles,
	}

	if err := cm.callOut(ctx, session, domains.OpRetrieveDomain.String(domains.OperationNames), id, params); err != nil {
		return domains.Domain{}, err
	}

	return cm.svc.RetrieveDomain(ctx, session, id, withRoles)
}

func (cm *calloutMiddleware) UpdateDomain(ctx context.Context, session authn.Session, id string, d domains.DomainReq) (domains.Domain, error) {
	params := map[string]any{
		"domain_req": d,
	}

	if err := cm.callOut(ctx, session, domains.OpUpdateDomain.String(domains.OperationNames), id, params); err != nil {
		return domains.Domain{}, err
	}

	return cm.svc.UpdateDomain(ctx, session, id, d)
}

func (cm *calloutMiddleware) EnableDomain(ctx context.Context, session authn.Session, id string) (domains.Domain, error) {
	if err := cm.callOut(ctx, session, domains.OpEnableDomain.String(domains.OperationNames), id, nil); err != nil {
		return domains.Domain{}, err
	}

	return cm.svc.EnableDomain(ctx, session, id)
}

func (cm *calloutMiddleware) DisableDomain(ctx context.Context, session authn.Session, id string) (domains.Domain, error) {
	if err := cm.callOut(ctx, session, domains.OpDisableDomain.String(domains.OperationNames), id, nil); err != nil {
		return domains.Domain{}, err
	}

	return cm.svc.DisableDomain(ctx, session, id)
}

func (cm *calloutMiddleware) FreezeDomain(ctx context.Context, session authn.Session, id string) (domains.Domain, error) {
	if err := cm.callOut(ctx, session, domains.OpFreezeDomain.String(domains.OperationNames), id, nil); err != nil {
		return domains.Domain{}, err
	}

	return cm.svc.FreezeDomain(ctx, session, id)
}

func (cm *calloutMiddleware) ListDomains(ctx context.Context, session authn.Session, page domains.Page) (domains.DomainsPage, error) {
	params := map[string]any{
		"page": page,
	}

	if err := cm.callOut(ctx, session, domains.OpListDomains.String(domains.OperationNames), "", params); err != nil {
		return domains.DomainsPage{}, err
	}

	return cm.svc.ListDomains(ctx, session, page)
}

func (cm *calloutMiddleware) SendInvitation(ctx context.Context, session authn.Session, invitation domains.Invitation) (err error) {
	params := map[string]any{
		"invitation": invitation,
	}

	// While entity here is technically an invitation, Domain is used as
	// the entity in callout since the invitation refers to the domain.
	if err := cm.callOut(ctx, session, domains.OpSendInvitation.String(domains.OperationNames), invitation.DomainID, params); err != nil {
		return err
	}

	return cm.svc.SendInvitation(ctx, session, invitation)
}

func (cm *calloutMiddleware) ListInvitations(ctx context.Context, session authn.Session, page domains.InvitationPageMeta) (domains.InvitationPage, error) {
	params := map[string]any{
		"page": page,
	}

	if err := cm.callOut(ctx, session, domains.OpListInvitations.String(domains.OperationNames), "", params); err != nil {
		return domains.InvitationPage{}, err
	}

	return cm.svc.ListInvitations(ctx, session, page)
}

func (cm *calloutMiddleware) ListDomainInvitations(ctx context.Context, session authn.Session, page domains.InvitationPageMeta) (domains.InvitationPage, error) {
	params := map[string]any{
		"page": page,
	}

	if err := cm.callOut(ctx, session, domains.OpListDomainInvitations.String(domains.OperationNames), page.DomainID, params); err != nil {
		return domains.InvitationPage{}, err
	}

	return cm.svc.ListDomainInvitations(ctx, session, page)
}

func (cm *calloutMiddleware) AcceptInvitation(ctx context.Context, session authn.Session, domainID string) (domains.Invitation, error) {
	// Similar to sending an invitation, Domain is used as the
	// entity in callout since the invitation refers to the domain.
	if err := cm.callOut(ctx, session, domains.OpAcceptInvitation.String(domains.OperationNames), domainID, nil); err != nil {
		return domains.Invitation{}, err
	}

	return cm.svc.AcceptInvitation(ctx, session, domainID)
}

func (cm *calloutMiddleware) RejectInvitation(ctx context.Context, session authn.Session, domainID string) error {
	// Similar to sending and accepting, Domain is used as
	// the entity in callout since the invitation refers to the domain.
	if err := cm.callOut(ctx, session, domains.OpRejectInvitation.String(domains.OperationNames), domainID, nil); err != nil {
		return err
	}

	return cm.svc.RejectInvitation(ctx, session, domainID)
}

func (cm *calloutMiddleware) DeleteInvitation(ctx context.Context, session authn.Session, inviteeUserID, domainID string) error {
	params := map[string]any{
		"invitee_user_id": inviteeUserID,
	}

	if err := cm.callOut(ctx, session, domains.OpDeleteInvitation.String(domains.OperationNames), domainID, params); err != nil {
		return err
	}

	return cm.svc.DeleteInvitation(ctx, session, inviteeUserID, domainID)
}

func (cm *calloutMiddleware) callOut(ctx context.Context, session authn.Session, op, entityID string, pld map[string]any) error {
	req := callout.Request{
		BaseRequest: callout.BaseRequest{
			Operation:  op,
			EntityType: policies.DomainType,
			EntityID:   entityID,
			CallerID:   session.UserID,
			CallerType: policies.UserType,
			DomainID:   entityID,
			Time:       time.Now().UTC(),
		},
		Payload: pld,
	}

	if err := cm.callout.Callout(ctx, req); err != nil {
		return err
	}

	return nil
}
