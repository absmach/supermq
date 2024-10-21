// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"

	"github.com/absmach/magistrala/channels"
	"github.com/absmach/magistrala/pkg/authn"
	"github.com/absmach/magistrala/pkg/authz"
	mgauthz "github.com/absmach/magistrala/pkg/authz"
	"github.com/absmach/magistrala/pkg/policies"
	rmMW "github.com/absmach/magistrala/pkg/roles/rolemanager/middleware"
	"github.com/absmach/magistrala/pkg/svcutil"
)

var _ channels.Service = (*authorizationMiddleware)(nil)

type authorizationMiddleware struct {
	svc   channels.Service
	authz mgauthz.Authorization
	opp   svcutil.OperationPerm
	rmMW.RoleManagerAuthorizationMiddleware
}

// AuthorizationMiddleware adds authorization to the channels service.
func AuthorizationMiddleware(svc channels.Service, authz mgauthz.Authorization, channelsOpPerm, rolesOpPerm map[svcutil.Operation]svcutil.Permission) (channels.Service, error) {
	opp := channels.NewOperationPerm()
	if err := opp.AddOperationPermissionMap(channelsOpPerm); err != nil {
		return nil, err
	}
	if err := opp.Validate(); err != nil {
		return nil, err
	}
	ram, err := rmMW.NewRoleManagerAuthorizationMiddleware(policies.ChannelType, svc, authz, rolesOpPerm)
	if err != nil {
		return nil, err
	}
	return &authorizationMiddleware{
		svc:                                svc,
		authz:                              authz,
		RoleManagerAuthorizationMiddleware: ram,
		opp:                                opp,
	}, nil
}

func (am *authorizationMiddleware) CreateChannels(ctx context.Context, session authn.Session, chs ...channels.Channel) ([]channels.Channel, error) {
	// If domain is disabled , then this authorization will fail for all non-admin domain users
	if err := am.authorize(ctx, channels.OpCreateChannel, authz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.DomainType,
		Object:      session.DomainID,
	}); err != nil {
		return []channels.Channel{}, err
	}
	return am.svc.CreateChannels(ctx, session, chs...)
}

func (am *authorizationMiddleware) ViewChannel(ctx context.Context, session authn.Session, id string) (channels.Channel, error) {
	if err := am.authorize(ctx, channels.OpViewChannel, authz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.ChannelType,
		Object:      id,
	}); err != nil {
		return channels.Channel{}, err
	}
	return am.svc.ViewChannel(ctx, session, id)
}

func (am *authorizationMiddleware) ListChannels(ctx context.Context, session authn.Session, pm channels.PageMetadata) (channels.Page, error) {
	if err := am.checkSuperAdmin(ctx, session.UserID); err != nil {
		if err := am.authorize(ctx, channels.OpListChannel, authz.PolicyReq{
			Domain:      session.DomainID,
			SubjectType: policies.UserType,
			Subject:     session.DomainUserID,
			ObjectType:  policies.DomainType,
			Object:      session.DomainID,
		}); err != nil {
			return channels.Page{}, err
		}
		return am.svc.ListChannels(ctx, session, pm)
	}
	session.SuperAdmin = true
	return am.svc.ListChannels(ctx, session, pm)
}

func (am *authorizationMiddleware) ListChannelsByThing(ctx context.Context, session authn.Session, thingID string, pm channels.PageMetadata) (channels.Page, error) {

	return am.svc.ListChannelsByThing(ctx, session, thingID, pm)
}

func (am *authorizationMiddleware) UpdateChannel(ctx context.Context, session authn.Session, channel channels.Channel) (channels.Channel, error) {
	if err := am.authorize(ctx, channels.OpUpdateChannel, authz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.ChannelType,
		Object:      channel.ID,
	}); err != nil {
		return channels.Channel{}, err
	}
	return am.svc.UpdateChannel(ctx, session, channel)
}

func (am *authorizationMiddleware) UpdateChannelTags(ctx context.Context, session authn.Session, channel channels.Channel) (channels.Channel, error) {
	if err := am.authorize(ctx, channels.OpUpdateChannelTags, authz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.ChannelType,
		Object:      channel.ID,
	}); err != nil {
		return channels.Channel{}, err
	}
	return am.svc.UpdateChannelTags(ctx, session, channel)
}

func (am *authorizationMiddleware) EnableChannel(ctx context.Context, session authn.Session, id string) (channels.Channel, error) {

	return am.svc.EnableChannel(ctx, session, id)
}

func (am *authorizationMiddleware) DisableChannel(ctx context.Context, session authn.Session, id string) (channels.Channel, error) {

	return am.svc.DisableChannel(ctx, session, id)
}

func (am *authorizationMiddleware) RemoveChannel(ctx context.Context, session authn.Session, id string) error {
	if err := am.authorize(ctx, channels.OpDeleteChannel, authz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.ChannelType,
		Object:      id,
	}); err != nil {
		return err
	}
	return am.svc.RemoveChannel(ctx, session, id)
}

func (am *authorizationMiddleware) Connect(ctx context.Context, session authn.Session, chIDs, thIDs []string) error {
	//ToDo: This authorization will be changed with Bulk Authorization. For this we need to add bulk authorization API in policies.
	for _, chID := range chIDs {
		if err := am.authorize(ctx, channels.OpConnectThingChannel, authz.PolicyReq{
			Domain:      session.DomainID,
			SubjectType: policies.UserType,
			Subject:     session.DomainUserID,
			ObjectType:  policies.ChannelType,
			Object:      chID,
		}); err != nil {
			return err
		}
	}

	for _, thID := range thIDs {
		if err := am.authorize(ctx, channels.OpConnectThingChannel, authz.PolicyReq{
			Domain:      session.DomainID,
			SubjectType: policies.UserType,
			Subject:     session.DomainUserID,
			ObjectType:  policies.ChannelType,
			Object:      thID,
		}); err != nil {
			return err
		}
	}
	return am.svc.Connect(ctx, session, chIDs, thIDs)
}
func (am *authorizationMiddleware) Disconnect(ctx context.Context, session authn.Session, chIDs, thIDs []string) error {
	//ToDo: This authorization will be changed with Bulk Authorization. For this we need to add bulk authorization API in policies.
	for _, chID := range chIDs {
		if err := am.authorize(ctx, channels.OpDisconnectThingChannel, authz.PolicyReq{
			Domain:      session.DomainID,
			SubjectType: policies.UserType,
			Subject:     session.DomainUserID,
			ObjectType:  policies.ChannelType,
			Object:      chID,
		}); err != nil {
			return err
		}
	}

	for _, thID := range thIDs {
		if err := am.authorize(ctx, channels.OpDisconnectThingChannel, authz.PolicyReq{
			Domain:      session.DomainID,
			SubjectType: policies.UserType,
			Subject:     session.DomainUserID,
			ObjectType:  policies.ChannelType,
			Object:      thID,
		}); err != nil {
			return err
		}
	}
	return am.svc.Disconnect(ctx, session, chIDs, thIDs)
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
