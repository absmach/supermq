// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"

	smqauth "github.com/absmach/supermq/auth"
	"github.com/absmach/supermq/channels"
	"github.com/absmach/supermq/pkg/authn"
	"github.com/absmach/supermq/pkg/connections"
	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	smqpat "github.com/absmach/supermq/pkg/pat"
	"github.com/absmach/supermq/pkg/roles"
)

var _ channels.Service = (*patMiddleware)(nil)

type patMiddleware struct {
	svc channels.Service
	pat smqpat.Authorization
}

func PATMiddleware(svc channels.Service, pat smqpat.Authorization) channels.Service {
	return &patMiddleware{
		svc: svc,
		pat: pat,
	}
}

func (pm *patMiddleware) authorizePAT(ctx context.Context, session authn.Session, platformEntityType smqauth.PlatformEntityType, optionalDomainEntityType smqauth.DomainEntityType, OptionalDomainID string, operation smqauth.OperationType, entityIDs []string) error {
	if session.Type != authn.PersonalAccessToken {
		return nil
	}
	if session.PatID == "" || session.UserID == "" {
		return errors.Wrap(svcerr.ErrAuthentication, errors.New("invalid PAT credentials"))
	}

	if err := pm.pat.AuthorizePAT(ctx, smqpat.PatReq{
		UserID:                   session.UserID,
		PatID:                    session.PatID,
		PlatformEntityType:       platformEntityType,
		OptionalDomainEntityType: optionalDomainEntityType,
		OptionalDomainID:         OptionalDomainID,
		Operation:                operation,
		EntityIDs:                entityIDs,
	}); err != nil {
		return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return nil
}

func (pm *patMiddleware) CreateChannels(ctx context.Context, session authn.Session, chs ...channels.Channel) ([]channels.Channel, []roles.RoleProvision, error) {
	if err := pm.authorizePAT(ctx, session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainChannelsScope,
		session.DomainID,
		smqauth.CreateOp,
		smqauth.AnyIDs{}.Values(),
	); err != nil {
		return []channels.Channel{}, []roles.RoleProvision{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.CreateChannels(ctx, session, chs...)
}

func (pm *patMiddleware) ViewChannel(ctx context.Context, session authn.Session, id string) (channels.Channel, error) {
	if err := pm.authorizePAT(ctx, session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainChannelsScope,
		session.DomainID,
		smqauth.ReadOp,
		[]string{id},
	); err != nil {
		return channels.Channel{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.ViewChannel(ctx, session, id)
}

func (pm *patMiddleware) ListChannels(ctx context.Context, session authn.Session, pg channels.PageMetadata) (channels.Page, error) {
	if err := pm.authorizePAT(ctx, session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainChannelsScope,
		session.DomainID,
		smqauth.ListOp,
		smqauth.AnyIDs{}.Values(),
	); err != nil {
		return channels.Page{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.ListChannels(ctx, session, pg)
}

func (pm *patMiddleware) ListUserChannels(ctx context.Context, session authn.Session, userID string, pg channels.PageMetadata) (channels.Page, error) {
	if err := pm.authorizePAT(ctx, session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainChannelsScope,
		session.DomainID,
		smqauth.ListOp,
		smqauth.AnyIDs{}.Values(),
	); err != nil {
		return channels.Page{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.ListUserChannels(ctx, session, userID, pg)
}

func (pm *patMiddleware) UpdateChannel(ctx context.Context, session authn.Session, channel channels.Channel) (channels.Channel, error) {
	if err := pm.authorizePAT(ctx, session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainChannelsScope,
		session.DomainID,
		smqauth.UpdateOp,
		[]string{channel.ID},
	); err != nil {
		return channels.Channel{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.UpdateChannel(ctx, session, channel)
}

func (pm *patMiddleware) UpdateChannelTags(ctx context.Context, session authn.Session, channel channels.Channel) (channels.Channel, error) {
	if err := pm.authorizePAT(ctx, session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainChannelsScope,
		session.DomainID,
		smqauth.UpdateOp,
		[]string{channel.ID},
	); err != nil {
		return channels.Channel{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.UpdateChannelTags(ctx, session, channel)
}

func (pm *patMiddleware) EnableChannel(ctx context.Context, session authn.Session, id string) (channels.Channel, error) {
	if err := pm.authorizePAT(ctx, session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainChannelsScope,
		session.DomainID,
		smqauth.UpdateOp,
		[]string{id},
	); err != nil {
		return channels.Channel{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.EnableChannel(ctx, session, id)
}

func (pm *patMiddleware) DisableChannel(ctx context.Context, session authn.Session, id string) (channels.Channel, error) {
	if err := pm.authorizePAT(ctx, session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainChannelsScope,
		session.DomainID,
		smqauth.UpdateOp,
		[]string{id},
	); err != nil {
		return channels.Channel{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.DisableChannel(ctx, session, id)
}

func (pm *patMiddleware) RemoveChannel(ctx context.Context, session authn.Session, id string) error {
	if err := pm.authorizePAT(ctx, session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainChannelsScope,
		session.DomainID,
		smqauth.DeleteOp,
		[]string{id},
	); err != nil {
		return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.RemoveChannel(ctx, session, id)
}

func (pm *patMiddleware) Connect(ctx context.Context, session authn.Session, chIDs, thIDs []string, connTypes []connections.ConnType) error {
	if err := pm.authorizePAT(ctx, session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainChannelsScope,
		session.DomainID,
		smqauth.CreateOp,
		chIDs,
	); err != nil {
		return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	if err := pm.authorizePAT(ctx, session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainClientsScope,
		session.DomainID,
		smqauth.CreateOp,
		thIDs,
	); err != nil {
		return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.Connect(ctx, session, chIDs, thIDs, connTypes)
}

func (pm *patMiddleware) Disconnect(ctx context.Context, session authn.Session, chIDs, thIDs []string, connTypes []connections.ConnType) error {
	if err := pm.authorizePAT(ctx, session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainChannelsScope,
		session.DomainID,
		smqauth.DeleteOp,
		chIDs,
	); err != nil {
		return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	if err := pm.authorizePAT(ctx, session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainClientsScope,
		session.DomainID,
		smqauth.DeleteOp,
		thIDs,
	); err != nil {
		return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.Disconnect(ctx, session, chIDs, thIDs, connTypes)
}

func (pm *patMiddleware) SetParentGroup(ctx context.Context, session authn.Session, parentGroupID string, id string) error {
	if err := pm.authorizePAT(ctx, session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainChannelsScope,
		session.DomainID,
		smqauth.UpdateOp,
		[]string{id},
	); err != nil {
		return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.SetParentGroup(ctx, session, parentGroupID, id)
}

func (pm *patMiddleware) RemoveParentGroup(ctx context.Context, session authn.Session, id string) error {
	if err := pm.authorizePAT(ctx, session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainChannelsScope,
		session.DomainID,
		smqauth.DeleteOp,
		[]string{id},
	); err != nil {
		return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.RemoveParentGroup(ctx, session, id)
}

func (pm *patMiddleware) AddRole(ctx context.Context, session authn.Session, entityID string, roleName string, optionalActions []string, optionalMembers []string) (roles.RoleProvision, error) {
	if err := pm.authorizePAT(ctx,
		session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainChannelsScope,
		session.DomainID,
		smqauth.CreateOp,
		[]string{entityID},
	); err != nil {
		return roles.RoleProvision{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.AddRole(ctx, session, entityID, roleName, optionalActions, optionalMembers)
}

func (pm *patMiddleware) ListAvailableActions(ctx context.Context, session authn.Session) ([]string, error) {
	if err := pm.authorizePAT(ctx,
		session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainChannelsScope,
		session.DomainID,
		smqauth.ListOp,
		smqauth.AnyIDs{}.Values(),
	); err != nil {
		return []string{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.ListAvailableActions(ctx, session)
}

func (pm *patMiddleware) RemoveMemberFromAllRoles(ctx context.Context, session authn.Session, member string) error {
	if err := pm.authorizePAT(ctx,
		session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainChannelsScope,
		session.DomainID,
		smqauth.DeleteOp,
		smqauth.AnyIDs{}.Values(),
	); err != nil {
		return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.RemoveMemberFromAllRoles(ctx, session, member)
}

func (pm *patMiddleware) RemoveRole(ctx context.Context, session authn.Session, entityID, roleID string) error {
	if err := pm.authorizePAT(ctx,
		session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainChannelsScope,
		session.DomainID,
		smqauth.DeleteOp,
		[]string{entityID},
	); err != nil {
		return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.RemoveRole(ctx, session, entityID, roleID)
}

func (pm *patMiddleware) RetrieveAllRoles(ctx context.Context, session authn.Session, entityID string, limit, offset uint64) (roles.RolePage, error) {
	if err := pm.authorizePAT(ctx,
		session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainChannelsScope,
		session.DomainID,
		smqauth.ListOp,
		[]string{entityID},
	); err != nil {
		return roles.RolePage{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.RetrieveAllRoles(ctx, session, entityID, limit, offset)
}

func (pm *patMiddleware) RetrieveRole(ctx context.Context, session authn.Session, entityID, roleID string) (roles.Role, error) {
	if err := pm.authorizePAT(ctx,
		session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainChannelsScope,
		session.DomainID,
		smqauth.ReadOp,
		[]string{entityID},
	); err != nil {
		return roles.Role{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.RetrieveRole(ctx, session, entityID, roleID)
}

func (pm *patMiddleware) ListEntityMembers(ctx context.Context, session authn.Session, entityID string, pq roles.MembersRolePageQuery) (roles.MembersRolePage, error) {
	if err := pm.authorizePAT(ctx,
		session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainChannelsScope,
		session.DomainID,
		smqauth.ListOp,
		[]string{entityID},
	); err != nil {
		return roles.MembersRolePage{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.ListEntityMembers(ctx, session, entityID, pq)
}

func (pm *patMiddleware) RemoveEntityMembers(ctx context.Context, session authn.Session, entityID string, members []string) error {
	if err := pm.authorizePAT(ctx,
		session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainChannelsScope,
		session.DomainID,
		smqauth.DeleteOp,
		[]string{entityID},
	); err != nil {
		return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.RemoveEntityMembers(ctx, session, entityID, members)
}

func (pm *patMiddleware) RoleAddActions(ctx context.Context, session authn.Session, entityID, roleID string, actions []string) ([]string, error) {
	if err := pm.authorizePAT(ctx,
		session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainChannelsScope,
		session.DomainID,
		smqauth.UpdateOp,
		[]string{entityID},
	); err != nil {
		return []string{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.RoleAddActions(ctx, session, entityID, roleID, actions)
}

func (pm *patMiddleware) RoleListActions(ctx context.Context, session authn.Session, entityID, roleID string) ([]string, error) {
	if err := pm.authorizePAT(ctx,
		session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainChannelsScope,
		session.DomainID,
		smqauth.ReadOp,
		[]string{entityID},
	); err != nil {
		return []string{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.RoleListActions(ctx, session, entityID, roleID)
}

func (pm *patMiddleware) RoleCheckActionsExists(ctx context.Context, session authn.Session, entityID, roleID string, actions []string) (bool, error) {
	if err := pm.authorizePAT(ctx,
		session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainChannelsScope,
		session.DomainID,
		smqauth.ReadOp,
		[]string{entityID},
	); err != nil {
		return false, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.RoleCheckActionsExists(ctx, session, entityID, roleID, actions)
}

func (pm *patMiddleware) RoleRemoveActions(ctx context.Context, session authn.Session, entityID, roleID string, actions []string) error {
	if err := pm.authorizePAT(ctx,
		session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainChannelsScope,
		session.DomainID,
		smqauth.DeleteOp,
		[]string{entityID},
	); err != nil {
		return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.RoleRemoveActions(ctx, session, entityID, roleID, actions)
}

func (pm *patMiddleware) RoleRemoveAllActions(ctx context.Context, session authn.Session, entityID, roleID string) error {
	if err := pm.authorizePAT(ctx,
		session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainChannelsScope,
		session.DomainID,
		smqauth.DeleteOp,
		[]string{entityID},
	); err != nil {
		return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.RoleRemoveAllActions(ctx, session, entityID, roleID)
}

func (pm *patMiddleware) RoleAddMembers(ctx context.Context, session authn.Session, entityID, roleID string, members []string) ([]string, error) {
	if err := pm.authorizePAT(ctx,
		session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainChannelsScope,
		session.DomainID,
		smqauth.UpdateOp,
		[]string{entityID},
	); err != nil {
		return []string{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.RoleAddMembers(ctx, session, entityID, roleID, members)
}

func (pm *patMiddleware) RoleListMembers(ctx context.Context, session authn.Session, entityID, roleID string, limit, offset uint64) (roles.MembersPage, error) {
	if err := pm.authorizePAT(ctx,
		session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainChannelsScope,
		session.DomainID,
		smqauth.ReadOp,
		[]string{entityID},
	); err != nil {
		return roles.MembersPage{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.RoleListMembers(ctx, session, entityID, roleID, limit, offset)
}

func (pm *patMiddleware) RoleCheckMembersExists(ctx context.Context, session authn.Session, entityID, roleID string, members []string) (bool, error) {
	if err := pm.authorizePAT(ctx,
		session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainChannelsScope,
		session.DomainID,
		smqauth.ReadOp,
		[]string{entityID},
	); err != nil {
		return false, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.RoleCheckMembersExists(ctx, session, entityID, roleID, members)
}

func (pm *patMiddleware) RoleRemoveMembers(ctx context.Context, session authn.Session, entityID, roleID string, members []string) error {
	if err := pm.authorizePAT(ctx,
		session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainChannelsScope,
		session.DomainID,
		smqauth.DeleteOp,
		[]string{entityID},
	); err != nil {
		return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.RoleRemoveMembers(ctx, session, entityID, roleID, members)
}

func (pm *patMiddleware) RoleRemoveAllMembers(ctx context.Context, session authn.Session, entityID, roleID string) error {
	if err := pm.authorizePAT(ctx,
		session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainChannelsScope,
		session.DomainID,
		smqauth.DeleteOp,
		[]string{entityID},
	); err != nil {
		return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.RoleRemoveAllMembers(ctx, session, entityID, roleID)
}

func (pm *patMiddleware) UpdateRoleName(ctx context.Context, session authn.Session, entityID, roleID, newRoleName string) (roles.Role, error) {
	if err := pm.authorizePAT(ctx,
		session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainChannelsScope,
		session.DomainID,
		smqauth.UpdateOp,
		[]string{entityID},
	); err != nil {
		return roles.Role{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.UpdateRoleName(ctx, session, entityID, roleID, newRoleName)
}
