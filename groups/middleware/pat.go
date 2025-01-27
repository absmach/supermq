// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"

	smqauth "github.com/absmach/supermq/auth"
	"github.com/absmach/supermq/groups"
	"github.com/absmach/supermq/pkg/authn"
	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	smqpat "github.com/absmach/supermq/pkg/pat"
	"github.com/absmach/supermq/pkg/roles"
)

var _ groups.Service = (*patMiddleware)(nil)

type patMiddleware struct {
	svc groups.Service
	pat smqpat.Authorization
}

func PATMiddleware(svc groups.Service, pat smqpat.Authorization) groups.Service {
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

func (pm *patMiddleware) CreateGroup(ctx context.Context, session authn.Session, g groups.Group) (groups.Group, []roles.RoleProvision, error) {
	if err := pm.authorizePAT(ctx, session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainGroupsScope,
		session.DomainID,
		smqauth.CreateOp,
		smqauth.AnyIDs{}.Values(),
	); err != nil {
		return groups.Group{}, []roles.RoleProvision{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.CreateGroup(ctx, session, g)
}

func (pm *patMiddleware) UpdateGroup(ctx context.Context, session authn.Session, g groups.Group) (groups.Group, error) {
	if err := pm.authorizePAT(ctx, session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainGroupsScope,
		session.DomainID,
		smqauth.UpdateOp,
		[]string{g.ID},
	); err != nil {
		return groups.Group{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.UpdateGroup(ctx, session, g)
}

func (pm *patMiddleware) ViewGroup(ctx context.Context, session authn.Session, id string) (groups.Group, error) {
	if err := pm.authorizePAT(ctx, session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainGroupsScope,
		session.DomainID,
		smqauth.ReadOp,
		[]string{id},
	); err != nil {
		return groups.Group{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.ViewGroup(ctx, session, id)
}

func (pm *patMiddleware) ListGroups(ctx context.Context, session authn.Session, gm groups.PageMeta) (groups.Page, error) {
	if err := pm.authorizePAT(ctx, session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainGroupsScope,
		session.DomainID,
		smqauth.ListOp,
		smqauth.AnyIDs{}.Values(),
	); err != nil {
		return groups.Page{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.ListGroups(ctx, session, gm)
}

func (pm *patMiddleware) ListUserGroups(ctx context.Context, session authn.Session, userID string, pg groups.PageMeta) (groups.Page, error) {
	if err := pm.authorizePAT(ctx, session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainGroupsScope,
		session.DomainID,
		smqauth.ListOp,
		smqauth.AnyIDs{}.Values(),
	); err != nil {
		return groups.Page{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.ListUserGroups(ctx, session, userID, pg)
}

func (pm *patMiddleware) EnableGroup(ctx context.Context, session authn.Session, id string) (groups.Group, error) {
	if err := pm.authorizePAT(ctx, session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainGroupsScope,
		session.DomainID,
		smqauth.UpdateOp,
		[]string{id},
	); err != nil {
		return groups.Group{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.EnableGroup(ctx, session, id)
}

func (pm *patMiddleware) DisableGroup(ctx context.Context, session authn.Session, id string) (groups.Group, error) {
	if err := pm.authorizePAT(ctx, session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainGroupsScope,
		session.DomainID,
		smqauth.UpdateOp,
		[]string{id},
	); err != nil {
		return groups.Group{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.DisableGroup(ctx, session, id)
}

func (pm *patMiddleware) DeleteGroup(ctx context.Context, session authn.Session, id string) error {
	if err := pm.authorizePAT(ctx, session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainGroupsScope,
		session.DomainID,
		smqauth.DeleteOp,
		[]string{id},
	); err != nil {
		return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.DeleteGroup(ctx, session, id)
}

func (pm *patMiddleware) RetrieveGroupHierarchy(ctx context.Context, session authn.Session, id string, hm groups.HierarchyPageMeta) (groups.HierarchyPage, error) {
	if err := pm.authorizePAT(ctx, session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainGroupsScope,
		session.DomainID,
		smqauth.ListOp,
		[]string{id},
	); err != nil {
		return groups.HierarchyPage{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.RetrieveGroupHierarchy(ctx, session, id, hm)
}

func (pm *patMiddleware) AddParentGroup(ctx context.Context, session authn.Session, id, parentID string) error {
	if err := pm.authorizePAT(ctx, session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainGroupsScope,
		session.DomainID,
		smqauth.UpdateOp,
		[]string{id},
	); err != nil {
		return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.AddParentGroup(ctx, session, id, parentID)
}

func (pm *patMiddleware) RemoveParentGroup(ctx context.Context, session authn.Session, id string) error {
	if err := pm.authorizePAT(ctx, session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainGroupsScope,
		session.DomainID,
		smqauth.DeleteOp,
		[]string{id},
	); err != nil {
		return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.RemoveParentGroup(ctx, session, id)
}

func (pm *patMiddleware) AddChildrenGroups(ctx context.Context, session authn.Session, id string, childrenGroupIDs []string) error {
	if err := pm.authorizePAT(ctx, session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainGroupsScope,
		session.DomainID,
		smqauth.UpdateOp,
		[]string{id},
	); err != nil {
		return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.AddChildrenGroups(ctx, session, id, childrenGroupIDs)
}

func (pm *patMiddleware) RemoveChildrenGroups(ctx context.Context, session authn.Session, id string, childrenGroupIDs []string) error {
	if err := pm.authorizePAT(ctx, session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainGroupsScope,
		session.DomainID,
		smqauth.DeleteOp,
		[]string{id},
	); err != nil {
		return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.RemoveChildrenGroups(ctx, session, id, childrenGroupIDs)
}

func (pm *patMiddleware) RemoveAllChildrenGroups(ctx context.Context, session authn.Session, id string) error {
	if err := pm.authorizePAT(ctx, session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainGroupsScope,
		session.DomainID,
		smqauth.DeleteOp,
		[]string{id},
	); err != nil {
		return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.RemoveAllChildrenGroups(ctx, session, id)
}

func (pm *patMiddleware) ListChildrenGroups(ctx context.Context, session authn.Session, id string, startLevel, endLevel int64, pg groups.PageMeta) (groups.Page, error) {
	if err := pm.authorizePAT(ctx, session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainGroupsScope,
		session.DomainID,
		smqauth.ListOp,
		[]string{id},
	); err != nil {
		return groups.Page{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.ListChildrenGroups(ctx, session, id, startLevel, endLevel, pg)
}

func (pm *patMiddleware) AddRole(ctx context.Context, session authn.Session, entityID string, roleName string, optionalActions []string, optionalMembers []string) (roles.RoleProvision, error) {
	if err := pm.authorizePAT(ctx,
		session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainGroupsScope,
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
		smqauth.DomainGroupsScope,
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
		smqauth.DomainGroupsScope,
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
		smqauth.DomainGroupsScope,
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
		smqauth.DomainGroupsScope,
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
		smqauth.DomainGroupsScope,
		session.DomainID,
		smqauth.ReadOp,
		[]string{entityID},
	); err != nil {
		return roles.Role{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.RetrieveRole(ctx, session, entityID, roleID)
}

func (pm *patMiddleware) RoleAddActions(ctx context.Context, session authn.Session, entityID, roleID string, actions []string) ([]string, error) {
	if err := pm.authorizePAT(ctx,
		session,
		smqauth.PlatformDomainsScope,
		smqauth.DomainGroupsScope,
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
		smqauth.DomainGroupsScope,
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
		smqauth.DomainGroupsScope,
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
		smqauth.DomainGroupsScope,
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
		smqauth.DomainGroupsScope,
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
		smqauth.DomainGroupsScope,
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
		smqauth.DomainGroupsScope,
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
		smqauth.DomainGroupsScope,
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
		smqauth.DomainGroupsScope,
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
		smqauth.DomainGroupsScope,
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
		smqauth.DomainGroupsScope,
		session.DomainID,
		smqauth.UpdateOp,
		[]string{entityID},
	); err != nil {
		return roles.Role{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
	}

	return pm.svc.UpdateRoleName(ctx, session, entityID, roleID, newRoleName)
}
