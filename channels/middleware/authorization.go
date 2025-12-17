// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"

	"github.com/absmach/supermq/auth"
	"github.com/absmach/supermq/channels"
	"github.com/absmach/supermq/clients"
	"github.com/absmach/supermq/domains"
	"github.com/absmach/supermq/groups"
	"github.com/absmach/supermq/pkg/authn"
	smqauthz "github.com/absmach/supermq/pkg/authz"
	"github.com/absmach/supermq/pkg/connections"
	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	"github.com/absmach/supermq/pkg/permissions"
	"github.com/absmach/supermq/pkg/policies"
	"github.com/absmach/supermq/pkg/roles"
	rolemgr "github.com/absmach/supermq/pkg/roles/rolemanager/middleware"
)

var (
	errView                     = errors.New("not authorized to view channel")
	errList                     = errors.New("not authorized to list user channels")
	errUpdate                   = errors.New("not authorized to update channel")
	errUpdateTags               = errors.New("not authorized to update channel tags")
	errEnable                   = errors.New("not authorized to enable channel")
	errDisable                  = errors.New("not authorized to disable channel")
	errDelete                   = errors.New("not authorized to delete channel")
	errConnect                  = errors.New("not authorized to connect to channel")
	errDisconnect               = errors.New("not authorized to disconnect from channel")
	errSetParentGroup           = errors.New("not authorized to set parent group to channel")
	errRemoveParentGroup        = errors.New("not authorized to remove parent group from channel")
	errDomainCreateChannels     = errors.New("not authorized to create channel in domain")
	errGroupSetChildChannels    = errors.New("not authorized to set child channel for group")
	errGroupRemoveChildChannels = errors.New("not authorized to remove child channel for group")
	errClientDisConnectChannels = errors.New("not authorized to disconnect channel for client")
	errClientConnectChannels    = errors.New("not authorized to connect channel for client")
)

var _ channels.Service = (*authorizationMiddleware)(nil)

type authorizationMiddleware struct {
	svc         channels.Service
	repo        channels.Repository
	authz       smqauthz.Authorization
	entitiesOps permissions.EntitiesOperations[permissions.Operation]
	rolemgr.RoleManagerAuthorizationMiddleware
}

// NewAuthorization adds authorization to the channels service.
func NewAuthorization(
	entityType string,
	svc channels.Service,
	authz smqauthz.Authorization,
	repo channels.Repository,
	entitiesOps permissions.EntitiesOperations[permissions.Operation],
	roleOps permissions.Operations[permissions.RoleOperation],
) (channels.Service, error) {
	if err := entitiesOps.Validate(); err != nil {
		return nil, err
	}
	ram, err := rolemgr.NewAuthorization(policies.ChannelType, svc, authz, roleOps)
	if err != nil {
		return nil, err
	}

	return &authorizationMiddleware{
		svc:                                svc,
		authz:                              authz,
		repo:                               repo,
		entitiesOps:                        entitiesOps,
		RoleManagerAuthorizationMiddleware: ram,
	}, nil
}

func (am *authorizationMiddleware) CreateChannels(ctx context.Context, session authn.Session, chs ...channels.Channel) ([]channels.Channel, []roles.RoleProvision, error) {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.ChannelsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.CreateOp,
			EntityID:         auth.AnyIDs,
		}); err != nil {
			return []channels.Channel{}, []roles.RoleProvision{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}
	if err := am.authorize(ctx, policies.DomainType, domains.OpCreateDomainChannels, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.DomainType,
		Object:      session.DomainID,
	}); err != nil {
		return []channels.Channel{}, []roles.RoleProvision{}, errors.Wrap(err, errDomainCreateChannels)
	}

	return am.svc.CreateChannels(ctx, session, chs...)
}

func (am *authorizationMiddleware) ViewChannel(ctx context.Context, session authn.Session, id string, withRoles bool) (channels.Channel, error) {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.ChannelsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.ReadOp,
			EntityID:         id,
		}); err != nil {
			return channels.Channel{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}

	if err := am.authorize(ctx, policies.ChannelType, channels.OpViewChannel, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.ChannelType,
		Object:      id,
	}); err != nil {
		return channels.Channel{}, errors.Wrap(err, errView)
	}

	return am.svc.ViewChannel(ctx, session, id, withRoles)
}

func (am *authorizationMiddleware) ListChannels(ctx context.Context, session authn.Session, pm channels.Page) (channels.ChannelsPage, error) {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.ChannelsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.ListOp,
			EntityID:         auth.AnyIDs,
		}); err != nil {
			return channels.ChannelsPage{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}

	if err := am.checkSuperAdmin(ctx, session); err == nil {
		session.SuperAdmin = true
	}

	return am.svc.ListChannels(ctx, session, pm)
}

func (am *authorizationMiddleware) ListUserChannels(ctx context.Context, session authn.Session, userID string, pm channels.Page) (channels.ChannelsPage, error) {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.ChannelsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.ListOp,
			EntityID:         auth.AnyIDs,
		}); err != nil {
			return channels.ChannelsPage{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}
	if err := am.checkSuperAdmin(ctx, session); err != nil {
		return channels.ChannelsPage{}, errors.Wrap(err, errList)
	}

	return am.svc.ListUserChannels(ctx, session, userID, pm)
}

func (am *authorizationMiddleware) UpdateChannel(ctx context.Context, session authn.Session, channel channels.Channel) (channels.Channel, error) {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.ChannelsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.UpdateOp,
			EntityID:         channel.ID,
		}); err != nil {
			return channels.Channel{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}

	if err := am.authorize(ctx, policies.ChannelType, channels.OpUpdateChannel, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.ChannelType,
		Object:      channel.ID,
	}); err != nil {
		return channels.Channel{}, errors.Wrap(err, errUpdate)
	}

	return am.svc.UpdateChannel(ctx, session, channel)
}

func (am *authorizationMiddleware) UpdateChannelTags(ctx context.Context, session authn.Session, channel channels.Channel) (channels.Channel, error) {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.ChannelsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.UpdateOp,
			EntityID:         channel.ID,
		}); err != nil {
			return channels.Channel{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}

	if err := am.authorize(ctx, policies.ChannelType, channels.OpUpdateChannelTags, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.ChannelType,
		Object:      channel.ID,
	}); err != nil {
		return channels.Channel{}, errors.Wrap(err, errUpdateTags)
	}

	return am.svc.UpdateChannelTags(ctx, session, channel)
}

func (am *authorizationMiddleware) EnableChannel(ctx context.Context, session authn.Session, id string) (channels.Channel, error) {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.ChannelsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.UpdateOp,
			EntityID:         id,
		}); err != nil {
			return channels.Channel{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}

	if err := am.authorize(ctx, policies.ChannelType, channels.OpEnableChannel, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.ChannelType,
		Object:      id,
	}); err != nil {
		return channels.Channel{}, errors.Wrap(err, errEnable)
	}

	return am.svc.EnableChannel(ctx, session, id)
}

func (am *authorizationMiddleware) DisableChannel(ctx context.Context, session authn.Session, id string) (channels.Channel, error) {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.ChannelsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.UpdateOp,
			EntityID:         id,
		}); err != nil {
			return channels.Channel{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}

	if err := am.authorize(ctx, policies.ChannelType, channels.OpDisableChannel, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.ChannelType,
		Object:      id,
	}); err != nil {
		return channels.Channel{}, errors.Wrap(err, errDisable)
	}

	return am.svc.DisableChannel(ctx, session, id)
}

func (am *authorizationMiddleware) RemoveChannel(ctx context.Context, session authn.Session, id string) error {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.ChannelsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.DeleteOp,
			EntityID:         id,
		}); err != nil {
			return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}
	if err := am.authorize(ctx, policies.ChannelType, channels.OpDeleteChannel, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.ChannelType,
		Object:      id,
	}); err != nil {
		return errors.Wrap(err, errDelete)
	}

	return am.svc.RemoveChannel(ctx, session, id)
}

func (am *authorizationMiddleware) Connect(ctx context.Context, session authn.Session, chIDs, thIDs []string, connTypes []connections.ConnType) error {
	if session.Type == authn.PersonalAccessToken {
		for _, chID := range chIDs {
			if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
				UserID:           session.UserID,
				PatID:            session.PatID,
				EntityType:       auth.ChannelsType,
				OptionalDomainID: session.DomainID,
				Operation:        auth.CreateOp,
				EntityID:         chID,
			}); err != nil {
				return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
			}
		}
		for _, thID := range thIDs {
			if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
				UserID:           session.UserID,
				PatID:            session.PatID,
				EntityType:       auth.ClientsType,
				OptionalDomainID: session.DomainID,
				Operation:        auth.CreateOp,
				EntityID:         thID,
			}); err != nil {
				return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
			}
		}
	}
	for _, chID := range chIDs {
		if err := am.authorize(ctx, policies.ChannelType, channels.OpConnectClient, smqauthz.PolicyReq{
			Domain:      session.DomainID,
			SubjectType: policies.UserType,
			Subject:     session.DomainUserID,
			ObjectType:  policies.ChannelType,
			Object:      chID,
		}); err != nil {
			return errors.Wrap(err, errConnect)
		}
	}

	for _, thID := range thIDs {
		if err := am.authorize(ctx, policies.ClientType, clients.OpConnectToChannel, smqauthz.PolicyReq{
			Domain:      session.DomainID,
			SubjectType: policies.UserType,
			Subject:     session.DomainUserID,
			ObjectType:  policies.ClientType,
			Object:      thID,
		}); err != nil {
			return errors.Wrap(err, errClientConnectChannels)
		}
	}

	return am.svc.Connect(ctx, session, chIDs, thIDs, connTypes)
}

func (am *authorizationMiddleware) Disconnect(ctx context.Context, session authn.Session, chIDs, thIDs []string, connTypes []connections.ConnType) error {
	if session.Type == authn.PersonalAccessToken {
		for _, chID := range chIDs {
			if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
				UserID:           session.UserID,
				PatID:            session.PatID,
				EntityType:       auth.ChannelsType,
				OptionalDomainID: session.DomainID,
				Operation:        auth.DeleteOp,
				EntityID:         chID,
			}); err != nil {
				return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
			}
		}
		for _, thID := range thIDs {
			if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
				UserID:           session.UserID,
				PatID:            session.PatID,
				EntityType:       auth.ClientsType,
				OptionalDomainID: session.DomainID,
				Operation:        auth.DeleteOp,
				EntityID:         thID,
			}); err != nil {
				return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
			}
		}
	}

	for _, chID := range chIDs {
		if err := am.authorize(ctx, policies.ChannelType, channels.OpDisconnectClient, smqauthz.PolicyReq{
			Domain:      session.DomainID,
			SubjectType: policies.UserType,
			Subject:     session.DomainUserID,
			ObjectType:  policies.ChannelType,
			Object:      chID,
		}); err != nil {
			return errors.Wrap(err, errDisconnect)
		}
	}

	for _, thID := range thIDs {
		if err := am.authorize(ctx, policies.ClientType, clients.OpDisconnectFromChannel, smqauthz.PolicyReq{
			Domain:      session.DomainID,
			SubjectType: policies.UserType,
			Subject:     session.DomainUserID,
			ObjectType:  policies.ClientType,
			Object:      thID,
		}); err != nil {
			return errors.Wrap(err, errClientDisConnectChannels)
		}
	}

	return am.svc.Disconnect(ctx, session, chIDs, thIDs, connTypes)
}

func (am *authorizationMiddleware) SetParentGroup(ctx context.Context, session authn.Session, parentGroupID string, id string) error {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.ChannelsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.UpdateOp,
			EntityID:         id,
		}); err != nil {
			return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}

	if err := am.authorize(ctx, policies.ChannelType, channels.OpSetParentGroup, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.ChannelType,
		Object:      id,
	}); err != nil {
		return errors.Wrap(err, errSetParentGroup)
	}

	if err := am.authorize(ctx, policies.GroupType, groups.OpGroupSetChildChannel, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.GroupType,
		Object:      parentGroupID,
	}); err != nil {
		return errors.Wrap(err, errGroupSetChildChannels)
	}

	return am.svc.SetParentGroup(ctx, session, parentGroupID, id)
}

func (am *authorizationMiddleware) RemoveParentGroup(ctx context.Context, session authn.Session, id string) error {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.ChannelsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.DeleteOp,
			EntityID:         id,
		}); err != nil {
			return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}

	if err := am.authorize(ctx, policies.ChannelType, channels.OpRemoveParentGroup, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		ObjectType:  policies.ChannelType,
		Object:      id,
	}); err != nil {
		return errors.Wrap(err, errRemoveParentGroup)
	}

	ch, err := am.repo.RetrieveByID(ctx, id)
	if err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}

	if ch.ParentGroup != "" {
		if err := am.authorize(ctx, policies.GroupType, groups.OpGroupRemoveChildChannel, smqauthz.PolicyReq{
			Domain:      session.DomainID,
			SubjectType: policies.UserType,
			Subject:     session.DomainUserID,
			ObjectType:  policies.GroupType,
			Object:      ch.ParentGroup,
		}); err != nil {
			return errors.Wrap(err, errGroupRemoveChildChannels)
		}

		return am.svc.RemoveParentGroup(ctx, session, id)
	}
	return nil
}

func (am *authorizationMiddleware) authorize(ctx context.Context, entityType string, op permissions.Operation, req smqauthz.PolicyReq) error {
	perm, err := am.entitiesOps.GetPermission(entityType, op)
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
