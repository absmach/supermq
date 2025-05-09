// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/absmach/supermq/auth"
	"github.com/absmach/supermq/groups"
	"github.com/absmach/supermq/pkg/authn"
	smqauthz "github.com/absmach/supermq/pkg/authz"
	"github.com/absmach/supermq/pkg/callout"
	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	"github.com/absmach/supermq/pkg/policies"
	"github.com/absmach/supermq/pkg/roles"
	rmMW "github.com/absmach/supermq/pkg/roles/rolemanager/middleware"
	"github.com/absmach/supermq/pkg/svcutil"
)

var (
	errView                        = errors.New("not authorized to view group")
	errUpdate                      = errors.New("not authorized to update group")
	errEnable                      = errors.New("not authorized to enable group")
	errDisable                     = errors.New("not authorized to disable group")
	errDelete                      = errors.New("not authorized to delete group")
	errViewHierarchy               = errors.New("not authorized to view group parent/children hierarchy")
	errListChildrenGroups          = errors.New("not authorized to view chidden groups of group")
	errSetParentGroup              = errors.New("not authorized to set parent group to group")
	errRemoveParentGroup           = errors.New("not authorized to remove parent group from group")
	errSetChildrenGroups           = errors.New("not authorized to set children groups to group")
	errRemoveChildrenGroups        = errors.New("not authorized to remove children groups from group")
	errParentGroupSetChildGroup    = errors.New("not authorized to set child group in parent group")
	errParentGroupRemoveChildGroup = errors.New("not authorized to remove child group from parent group")
	errChildGroupSetParentGroup    = errors.New("not authorized to set parent group to child group")
	errDomainCreateGroups          = errors.New("not authorized to create groups in domain")
	errDomainListGroups            = errors.New("not authorized to list groups in domain")
)

var _ groups.Service = (*authorizationMiddleware)(nil)

type authorizationMiddleware struct {
	svc     groups.Service
	repo    groups.Repository
	authz   smqauthz.Authorization
	opp     svcutil.OperationPerm
	extOpp  svcutil.ExternalOperationPerm
	callout callout.Callout
	rmMW.RoleManagerAuthorizationMiddleware
}

// AuthorizationMiddleware adds authorization to the clients service.
func AuthorizationMiddleware(entityType string,
	svc groups.Service,
	repo groups.Repository,
	authz smqauthz.Authorization,
	groupsOpPerm, rolesOpPerm map[svcutil.Operation]svcutil.Permission,
	extOpPerm map[svcutil.ExternalOperation]svcutil.Permission,
	httpClient *http.Client,
	method string,
	urls []string,
	permissions []string,
) (groups.Service, error) {
	opp := groups.NewOperationPerm()
	if err := opp.AddOperationPermissionMap(groupsOpPerm); err != nil {
		return nil, err
	}
	if err := opp.Validate(); err != nil {
		return nil, err
	}

	extOpp := groups.NewExternalOperationPerm()
	if err := extOpp.AddOperationPermissionMap(extOpPerm); err != nil {
		return nil, err
	}
	if err := extOpp.Validate(); err != nil {
		return nil, err
	}

	ram, err := rmMW.NewRoleManagerAuthorizationMiddleware(entityType, svc, authz, rolesOpPerm)
	if err != nil {
		return nil, err
	}

	call, err := callout.NewCallback(httpClient, method, urls, permissions)
	if err != nil {
		return nil, err
	}

	return &authorizationMiddleware{
		svc:                                svc,
		repo:                               repo,
		authz:                              authz,
		opp:                                opp,
		extOpp:                             extOpp,
		RoleManagerAuthorizationMiddleware: ram,
		callout:                            call,
	}, nil
}

func (am *authorizationMiddleware) CreateGroup(ctx context.Context, session authn.Session, g groups.Group) (groups.Group, []roles.RoleProvision, error) {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.GroupsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.CreateOp,
			EntityID:         auth.AnyIDs,
		}); err != nil {
			return groups.Group{}, []roles.RoleProvision{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}

	if err := am.extAuthorize(ctx, groups.DomainOpCreateGroup, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		SubjectKind: policies.UsersKind,
		Subject:     session.DomainUserID,
		Object:      session.DomainID,
		ObjectType:  policies.DomainType,
	}); err != nil {
		return groups.Group{}, []roles.RoleProvision{}, errors.Wrap(errDomainCreateGroups, err)
	}

	if g.Parent != "" {
		if err := am.authorize(ctx, groups.OpAddChildrenGroups, smqauthz.PolicyReq{
			Domain:      session.DomainID,
			SubjectType: policies.UserType,
			SubjectKind: policies.UsersKind,
			Subject:     session.DomainUserID,
			Object:      g.Parent,
			ObjectType:  policies.GroupType,
		}); err != nil {
			return groups.Group{}, []roles.RoleProvision{}, errors.Wrap(errParentGroupSetChildGroup, err)
		}
	}
	if err := am.Callout(ctx, session, groups.OpCreateGroup.String(groups.OperationNames)); err != nil {
		return groups.Group{}, []roles.RoleProvision{}, err
	}

	return am.svc.CreateGroup(ctx, session, g)
}

func (am *authorizationMiddleware) UpdateGroup(ctx context.Context, session authn.Session, g groups.Group) (groups.Group, error) {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.GroupsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.UpdateOp,
			EntityID:         g.ID,
		}); err != nil {
			return groups.Group{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}

	if err := am.authorize(ctx, groups.OpUpdateGroup, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		SubjectKind: policies.UsersKind,
		Subject:     session.DomainUserID,
		Object:      g.ID,
		ObjectType:  policies.GroupType,
	}); err != nil {
		return groups.Group{}, errors.Wrap(errUpdate, err)
	}
	if err := am.Callout(ctx, session, groups.OpUpdateGroup.String(groups.OperationNames)); err != nil {
		return groups.Group{}, err
	}

	return am.svc.UpdateGroup(ctx, session, g)
}

func (am *authorizationMiddleware) ViewGroup(ctx context.Context, session authn.Session, id string, withRoles bool) (groups.Group, error) {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.GroupsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.ReadOp,
			EntityID:         id,
		}); err != nil {
			return groups.Group{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}

	if err := am.authorize(ctx, groups.OpViewGroup, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		SubjectKind: policies.UsersKind,
		Subject:     session.DomainUserID,
		Object:      id,
		ObjectType:  policies.GroupType,
	}); err != nil {
		return groups.Group{}, errors.Wrap(errView, err)
	}
	if err := am.Callout(ctx, session, groups.OpViewGroup.String(groups.OperationNames)); err != nil {
		return groups.Group{}, err
	}

	return am.svc.ViewGroup(ctx, session, id, withRoles)
}

func (am *authorizationMiddleware) ListGroups(ctx context.Context, session authn.Session, gm groups.PageMeta) (groups.Page, error) {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.GroupsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.ListOp,
			EntityID:         auth.AnyIDs,
		}); err != nil {
			return groups.Page{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}

	err := am.checkSuperAdmin(ctx, session.UserID)
	if err == nil {
		session.SuperAdmin = true
		return am.svc.ListGroups(ctx, session, gm)
	}
	if err := am.extAuthorize(ctx, groups.DomainOpListGroups, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		SubjectKind: policies.UsersKind,
		Subject:     session.DomainUserID,
		Object:      session.DomainID,
		ObjectType:  policies.DomainType,
	}); err != nil {
		return groups.Page{}, errors.Wrap(errDomainListGroups, err)
	}

	return am.svc.ListGroups(ctx, session, gm)
}

func (am *authorizationMiddleware) ListUserGroups(ctx context.Context, session authn.Session, userID string, pm groups.PageMeta) (groups.Page, error) {
	err := am.checkSuperAdmin(ctx, session.UserID)
	if err == nil {
		session.SuperAdmin = true
		return am.svc.ListGroups(ctx, session, pm)
	}
	if err := am.extAuthorize(ctx, groups.UserOpListGroups, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		SubjectKind: policies.UsersKind,
		Subject:     session.DomainUserID,
		Object:      session.DomainID,
		ObjectType:  policies.DomainType,
	}); err != nil {
		return groups.Page{}, errors.Wrap(errDomainListGroups, err)
	}
	return am.svc.ListUserGroups(ctx, session, userID, pm)
}

func (am *authorizationMiddleware) EnableGroup(ctx context.Context, session authn.Session, id string) (groups.Group, error) {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.GroupsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.UpdateOp,
			EntityID:         id,
		}); err != nil {
			return groups.Group{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}

	if err := am.authorize(ctx, groups.OpEnableGroup, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		Object:      id,
		ObjectType:  policies.GroupType,
	}); err != nil {
		return groups.Group{}, errors.Wrap(errEnable, err)
	}
	if err := am.Callout(ctx, session, groups.OpEnableGroup.String(groups.OperationNames)); err != nil {
		return groups.Group{}, err
	}

	return am.svc.EnableGroup(ctx, session, id)
}

func (am *authorizationMiddleware) DisableGroup(ctx context.Context, session authn.Session, id string) (groups.Group, error) {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.GroupsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.UpdateOp,
			EntityID:         id,
		}); err != nil {
			return groups.Group{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}

	if err := am.authorize(ctx, groups.OpDisableGroup, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		Object:      id,
		ObjectType:  policies.GroupType,
	}); err != nil {
		return groups.Group{}, errors.Wrap(errDisable, err)
	}
	if err := am.Callout(ctx, session, groups.OpDisableGroup.String(groups.OperationNames)); err != nil {
		return groups.Group{}, err
	}

	return am.svc.DisableGroup(ctx, session, id)
}

func (am *authorizationMiddleware) DeleteGroup(ctx context.Context, session authn.Session, id string) error {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.GroupsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.DeleteOp,
			EntityID:         id,
		}); err != nil {
			return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}
	if err := am.authorize(ctx, groups.OpDeleteGroup, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		Object:      id,
		ObjectType:  policies.GroupType,
	}); err != nil {
		return errors.Wrap(errDelete, err)
	}
	if err := am.Callout(ctx, session, groups.OpDeleteGroup.String(groups.OperationNames)); err != nil {
		return err
	}

	return am.svc.DeleteGroup(ctx, session, id)
}

func (am *authorizationMiddleware) RetrieveGroupHierarchy(ctx context.Context, session authn.Session, id string, hm groups.HierarchyPageMeta) (groups.HierarchyPage, error) {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.GroupsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.ListOp,
			EntityID:         id,
		}); err != nil {
			return groups.HierarchyPage{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}

	if err := am.authorize(ctx, groups.OpRetrieveGroupHierarchy, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		Object:      id,
		ObjectType:  policies.GroupType,
	}); err != nil {
		return groups.HierarchyPage{}, errors.Wrap(errViewHierarchy, err)
	}
	if err := am.Callout(ctx, session, groups.OpRetrieveGroupHierarchy.String(groups.OperationNames)); err != nil {
		return groups.HierarchyPage{}, err
	}
	return am.svc.RetrieveGroupHierarchy(ctx, session, id, hm)
}

func (am *authorizationMiddleware) AddParentGroup(ctx context.Context, session authn.Session, id, parentID string) error {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.GroupsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.UpdateOp,
			EntityID:         id,
		}); err != nil {
			return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}

	if err := am.authorize(ctx, groups.OpAddParentGroup, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		Object:      id,
		ObjectType:  policies.GroupType,
	}); err != nil {
		return errors.Wrap(errSetParentGroup, err)
	}

	if err := am.authorize(ctx, groups.OpAddChildrenGroups, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		Object:      parentID,
		ObjectType:  policies.GroupType,
	}); err != nil {
		return errors.Wrap(errParentGroupSetChildGroup, err)
	}
	if err := am.Callout(ctx, session, groups.OpAddParentGroup.String(groups.OperationNames)); err != nil {
		return err
	}
	return am.svc.AddParentGroup(ctx, session, id, parentID)
}

func (am *authorizationMiddleware) RemoveParentGroup(ctx context.Context, session authn.Session, id string) error {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.GroupsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.DeleteOp,
			EntityID:         id,
		}); err != nil {
			return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}

	if err := am.authorize(ctx, groups.OpRemoveParentGroup, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		Object:      id,
		ObjectType:  policies.GroupType,
	}); err != nil {
		return errors.Wrap(errRemoveParentGroup, err)
	}

	group, err := am.repo.RetrieveByID(ctx, id)
	if err != nil {
		return errors.Wrap(svcerr.ErrViewEntity, err)
	}

	if group.Parent != "" {
		if err := am.authorize(ctx, groups.OpRemoveParentGroup, smqauthz.PolicyReq{
			Domain:      session.DomainID,
			SubjectType: policies.UserType,
			Subject:     session.DomainUserID,
			Object:      group.Parent,
			ObjectType:  policies.GroupType,
		}); err != nil {
			return errors.Wrap(errParentGroupRemoveChildGroup, err)
		}
	}
	if err := am.Callout(ctx, session, groups.OpRemoveParentGroup.String(groups.OperationNames)); err != nil {
		return err
	}
	return am.svc.RemoveParentGroup(ctx, session, id)
}

func (am *authorizationMiddleware) AddChildrenGroups(ctx context.Context, session authn.Session, id string, childrenGroupIDs []string) error {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.GroupsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.UpdateOp,
			EntityID:         id,
		}); err != nil {
			return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}

	if err := am.authorize(ctx, groups.OpAddChildrenGroups, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		Object:      id,
		ObjectType:  policies.GroupType,
	}); err != nil {
		return errors.Wrap(errSetChildrenGroups, err)
	}

	for _, childID := range childrenGroupIDs {
		if err := am.authorize(ctx, groups.OpAddParentGroup, smqauthz.PolicyReq{
			Domain:      session.DomainID,
			SubjectType: policies.UserType,
			Subject:     session.DomainUserID,
			Object:      childID,
			ObjectType:  policies.GroupType,
		}); err != nil {
			return errors.Wrap(errChildGroupSetParentGroup, errors.Wrap(fmt.Errorf("child group id: %s", childID), err))
		}
	}
	if err := am.Callout(ctx, session, groups.OpAddChildrenGroups.String(groups.OperationNames)); err != nil {
		return err
	}

	return am.svc.AddChildrenGroups(ctx, session, id, childrenGroupIDs)
}

func (am *authorizationMiddleware) RemoveChildrenGroups(ctx context.Context, session authn.Session, id string, childrenGroupIDs []string) error {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.GroupsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.DeleteOp,
			EntityID:         id,
		}); err != nil {
			return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}

	if err := am.authorize(ctx, groups.OpRemoveChildrenGroups, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		Object:      id,
		ObjectType:  policies.GroupType,
	}); err != nil {
		return errors.Wrap(errRemoveChildrenGroups, err)
	}
	if err := am.Callout(ctx, session, groups.OpRemoveChildrenGroups.String(groups.OperationNames)); err != nil {
		return err
	}

	return am.svc.RemoveChildrenGroups(ctx, session, id, childrenGroupIDs)
}

func (am *authorizationMiddleware) RemoveAllChildrenGroups(ctx context.Context, session authn.Session, id string) error {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.GroupsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.DeleteOp,
			EntityID:         id,
		}); err != nil {
			return errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}

	if err := am.authorize(ctx, groups.OpRemoveAllChildrenGroups, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		Object:      id,
		ObjectType:  policies.GroupType,
	}); err != nil {
		return err
	}
	if err := am.Callout(ctx, session, groups.OpRemoveAllChildrenGroups.String(groups.OperationNames)); err != nil {
		return err
	}

	return am.svc.RemoveAllChildrenGroups(ctx, session, id)
}

func (am *authorizationMiddleware) ListChildrenGroups(ctx context.Context, session authn.Session, id string, startLevel, endLevel int64, pm groups.PageMeta) (groups.Page, error) {
	if session.Type == authn.PersonalAccessToken {
		if err := am.authz.AuthorizePAT(ctx, smqauthz.PatReq{
			UserID:           session.UserID,
			PatID:            session.PatID,
			EntityType:       auth.GroupsType,
			OptionalDomainID: session.DomainID,
			Operation:        auth.ListOp,
			EntityID:         id,
		}); err != nil {
			return groups.Page{}, errors.Wrap(svcerr.ErrUnauthorizedPAT, err)
		}
	}

	if err := am.authorize(ctx, groups.OpListChildrenGroups, smqauthz.PolicyReq{
		Domain:      session.DomainID,
		SubjectType: policies.UserType,
		Subject:     session.DomainUserID,
		Object:      id,
		ObjectType:  policies.GroupType,
	}); err != nil {
		return groups.Page{}, errors.Wrap(errListChildrenGroups, err)
	}

	if err := am.Callout(ctx, session, groups.OpListChildrenGroups.String(groups.OperationNames)); err != nil {
		return groups.Page{}, err
	}

	return am.svc.ListChildrenGroups(ctx, session, id, startLevel, endLevel, pm)
}

func (am *authorizationMiddleware) checkSuperAdmin(ctx context.Context, adminID string) error {
	if err := am.authz.Authorize(ctx, smqauthz.PolicyReq{
		SubjectType: policies.UserType,
		Subject:     adminID,
		Permission:  policies.AdminPermission,
		ObjectType:  policies.PlatformType,
		Object:      policies.SuperMQObject,
	}); err != nil {
		return err
	}
	return nil
}

func (am *authorizationMiddleware) authorize(ctx context.Context, op svcutil.Operation, pr smqauthz.PolicyReq) error {
	perm, err := am.opp.GetPermission(op)
	if err != nil {
		return err
	}
	pr.Permission = perm.String()

	if err := am.authz.Authorize(ctx, pr); err != nil {
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

func (am *authorizationMiddleware) Callout(ctx context.Context, session authn.Session, op string) error {
	pl := map[string]interface{}{
		"entity_type": policies.GroupType,
		"sender":      session.UserID,
		"domain":      session.DomainID,
		"time":        time.Now().String(),
	}

	if err := am.callout.Callout(ctx, op, pl); err != nil {
		return err
	}

	return nil
}
