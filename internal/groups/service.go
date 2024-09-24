// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package groups

import (
	"context"
	"fmt"
	"time"

	"github.com/absmach/magistrala"
	"github.com/absmach/magistrala/auth"
	grpcclient "github.com/absmach/magistrala/auth/api/grpc"
	"github.com/absmach/magistrala/pkg/apiutil"
	mgclients "github.com/absmach/magistrala/pkg/clients"
	"github.com/absmach/magistrala/pkg/entityroles"
	"github.com/absmach/magistrala/pkg/errors"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	"github.com/absmach/magistrala/pkg/groups"
	"github.com/absmach/magistrala/pkg/roles"
	"github.com/absmach/magistrala/pkg/svcutil"
	"golang.org/x/sync/errgroup"
)

var (
	errParentUnAuthz = errors.New("failed to authorize parent group")
	errMemberKind    = errors.New("invalid member kind")
	errGroupIDs      = errors.New("invalid group ids")
)

type identity struct {
	ID       string
	DomainID string
	UserID   string
}

type service struct {
	groups     groups.Repository
	auth       grpcclient.AuthServiceClient
	policy     magistrala.PolicyServiceClient
	idProvider magistrala.IDProvider
	opp        svcutil.OperationPerm
	entityroles.RolesSvc
}

// NewService returns a new Clients service implementation.
func NewService(g groups.Repository, idp magistrala.IDProvider, authClient magistrala.AuthServiceClient) (groups.Service, error) {
	rolesSvc, err := entityroles.NewRolesSvc("group", g, idp, authClient, groups.AvailableActions(), groups.BuiltInRoles(), groups.NewRolesOperationPermissionMap())
	if err != nil {
		return service{}, err
	}
	opp := groups.NewOperationPerm()
	if err := opp.AddOperationPermissionMap(groups.NewOperationPermissionMap()); err != nil {
		return service{}, err
	}
	return service{
		groups:     g,
		idProvider: idp,
		auth:       authClient,
		opp:        opp,
		RolesSvc:   rolesSvc,
	}, nil
}

func (svc service) CreateGroup(ctx context.Context, token, kind string, g groups.Group) (gr groups.Group, err error) {
	userInfo, err := svc.identify(ctx, token)
	if err != nil {
		return groups.Group{}, err
	}
	// If domain is disabled , then this authorization will fail for all non-admin domain users
	if err := svc.authorize(ctx, groups.OpCreateGroup, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Subject:     userInfo.ID,
		Object:      userInfo.DomainID,
		ObjectType:  auth.DomainType,
	}); err != nil {
		return groups.Group{}, err
	}
	groupID, err := svc.idProvider.ID()
	if err != nil {
		return groups.Group{}, err
	}
	if g.Status != mgclients.EnabledStatus && g.Status != mgclients.DisabledStatus {
		return groups.Group{}, svcerr.ErrInvalidStatus
	}

	g.ID = groupID
	g.CreatedAt = time.Now()
	g.Domain = userInfo.DomainID
	if g.Parent != "" {
		if err := svc.authorize(ctx, groups.OpAddChildrenGroups, &magistrala.AuthorizeReq{
			Domain:      userInfo.DomainID,
			SubjectType: auth.UserType,
			SubjectKind: auth.UsersKind,
			Subject:     userInfo.ID,
			Object:      g.Parent,
			ObjectType:  auth.GroupType,
		}); err != nil {
			return groups.Group{}, errors.Wrap(errParentUnAuthz, err)
		}
	}

	saved, err := svc.groups.Save(ctx, g)
	if err != nil {
		return groups.Group{}, errors.Wrap(svcerr.ErrCreateEntity, err)
	}
	defer func() {
		if err != nil {
			if errRollback := svc.groups.Delete(ctx, saved.ID); errRollback != nil {
				err = errors.Wrap(errors.Wrap(errors.ErrRollbackTx, errRollback), err)
			}
		}
	}()

	oprs := []roles.OptionalPolicy{}

	oprs = append(oprs, roles.OptionalPolicy{
		Namespace:   userInfo.DomainID,
		SubjectType: auth.DomainType,
		Subject:     userInfo.DomainID,
		Relation:    auth.DomainRelation,
		ObjectType:  auth.GroupType,
		Object:      saved.ID,
	})
	if saved.Parent != "" {
		oprs = append(oprs, roles.OptionalPolicy{
			Namespace:   userInfo.DomainID,
			SubjectType: auth.GroupType,
			Subject:     saved.Parent,
			Relation:    auth.ParentGroupRelation,
			ObjectType:  auth.GroupType,
			Object:      saved.ID,
		})
	}
	newBuiltInRoleMembers := map[roles.BuiltInRoleName][]roles.Member{
		groups.BuiltInRoleAdmin:      {roles.Member(userInfo.UserID)},
		groups.BuiltInRoleMembership: {},
	}
	if _, err := svc.AddNewEntityRoles(ctx, userInfo.ID, userInfo.DomainID, saved.ID, newBuiltInRoleMembers, oprs); err != nil {
		return groups.Group{}, errors.Wrap(svcerr.ErrAddPolicies, err)
	}

	return saved, nil
}

func (svc service) ViewGroup(ctx context.Context, token, id string) (groups.Group, error) {
	userInfo, err := svc.identify(ctx, token)
	if err != nil {
		return groups.Group{}, err
	}

	if err := svc.authorize(ctx, groups.OpViewGroup, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		SubjectType: auth.UserType,
		Subject:     userInfo.ID,
		Object:      id,
		ObjectType:  auth.GroupType,
	}); err != nil {
		return groups.Group{}, err
	}

	group, err := svc.groups.RetrieveByID(ctx, id)
	if err != nil {
		return groups.Group{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}

	return group, nil
}

func (svc service) ListGroups(ctx context.Context, token string, gm groups.Page) (groups.Page, error) {
	userInfo, err := svc.identify(ctx, token)
	if err != nil {
		return groups.Page{}, err
	}

	var ids []string

	gm.PageMeta.DomainID = userInfo.DomainID
	if err := svc.checkSuperAdmin(ctx, userInfo.ID); err != nil {
		// If domain is disabled , then this authorization will fail for all non-admin domain users
		if err := svc.authorize(ctx, groups.OpListGroups, &magistrala.AuthorizeReq{
			Domain:      userInfo.DomainID,
			SubjectType: auth.UserType,
			Subject:     userInfo.ID,
			Object:      userInfo.DomainID,
			ObjectType:  auth.DomainType,
		}); err != nil {
			return groups.Page{}, err
		}
		ids, err = svc.listAllGroupsOfUserID(ctx, userInfo.ID, gm.Permission)
		if err != nil {
			return groups.Page{}, err
		}
	}

	gp, err := svc.groups.RetrieveByIDs(ctx, gm, ids...)
	if err != nil {
		return groups.Page{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}

	if gm.ListPerms && len(gp.Groups) > 0 {
		g, ctx := errgroup.WithContext(ctx)

		for i := range gp.Groups {
			// Copying loop variable "i" to avoid "loop variable captured by func literal"
			iter := i
			g.Go(func() error {
				return svc.retrievePermissions(ctx, userInfo.ID, &gp.Groups[iter])
			})
		}

		if err := g.Wait(); err != nil {
			return groups.Page{}, err
		}
	}
	return gp, nil
}

// Experimental functions used for async calling of svc.listUserThingPermission. This might be helpful during listing of large number of entities.
func (svc service) retrievePermissions(ctx context.Context, userID string, group *groups.Group) error {
	permissions, err := svc.listUserGroupPermission(ctx, userID, group.ID)
	if err != nil {
		return err
	}
	group.Permissions = permissions
	return nil
}

func (svc service) listUserGroupPermission(ctx context.Context, userID, groupID string) ([]string, error) {
	lp, err := svc.policy.ListPermissions(ctx, &magistrala.ListPermissionsReq{
		SubjectType: auth.UserType,
		Subject:     userID,
		Object:      groupID,
		ObjectType:  auth.GroupType,
	})
	if err != nil {
		return []string{}, err
	}
	if len(lp.GetPermissions()) == 0 {
		return []string{}, svcerr.ErrAuthorization
	}
	return lp.GetPermissions(), nil
}

func (svc service) checkSuperAdmin(ctx context.Context, userID string) error {
	res, err := svc.auth.Authorize(ctx, &magistrala.AuthorizeReq{
		SubjectType: auth.UserType,
		Subject:     userID,
		Permission:  auth.AdminPermission,
		ObjectType:  auth.PlatformType,
		Object:      auth.MagistralaObject,
	})
	if err != nil {
		return errors.Wrap(svcerr.ErrAuthorization, err)
	}
	if !res.Authorized {
		return svcerr.ErrAuthorization
	}
	return nil
}

func (svc service) UpdateGroup(ctx context.Context, token string, g groups.Group) (groups.Group, error) {
	userInfo, err := svc.identify(ctx, token)
	if err != nil {
		return groups.Group{}, err
	}

	if err := svc.authorize(ctx, groups.OpUpdateGroup, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		SubjectType: auth.UserType,
		Subject:     userInfo.ID,
		Object:      g.ID,
		ObjectType:  auth.GroupType,
	}); err != nil {
		return groups.Group{}, err
	}

	g.UpdatedAt = time.Now()
	g.UpdatedBy = userInfo.UserID

	return svc.groups.Update(ctx, g)
}

func (svc service) EnableGroup(ctx context.Context, token, id string) (groups.Group, error) {
	userInfo, err := svc.identify(ctx, token)
	if err != nil {
		return groups.Group{}, err
	}

	if err := svc.authorize(ctx, groups.OpEnableGroup, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		SubjectType: auth.UserType,
		Subject:     userInfo.ID,
		Object:      id,
		ObjectType:  auth.GroupType,
	}); err != nil {
		return groups.Group{}, err
	}

	group := groups.Group{
		ID:        id,
		Status:    mgclients.EnabledStatus,
		UpdatedAt: time.Now(),
	}
	group, err = svc.changeGroupStatus(ctx, userInfo, group)
	if err != nil {
		return groups.Group{}, err
	}
	return group, nil
}

func (svc service) DisableGroup(ctx context.Context, token, id string) (groups.Group, error) {
	userInfo, err := svc.identify(ctx, token)
	if err != nil {
		return groups.Group{}, err
	}

	if err := svc.authorize(ctx, groups.OpEnableGroup, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		SubjectType: auth.UserType,
		Subject:     userInfo.ID,
		Object:      id,
		ObjectType:  auth.GroupType,
	}); err != nil {
		return groups.Group{}, err
	}

	group := groups.Group{
		ID:        id,
		Status:    mgclients.DisabledStatus,
		UpdatedAt: time.Now(),
	}
	group, err = svc.changeGroupStatus(ctx, userInfo, group)
	if err != nil {
		return groups.Group{}, err
	}
	return group, nil
}

func (svc service) ListParentGroups(ctx context.Context, token, id string, gm groups.Page) (groups.Page, error) {
	userInfo, err := svc.identify(ctx, token)
	if err != nil {
		return groups.Page{}, err
	}

	if err := svc.authorize(ctx, groups.OpListParentGroups, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		SubjectType: auth.UserType,
		Subject:     userInfo.ID,
		Object:      id,
		ObjectType:  auth.GroupType,
	}); err != nil {
		return groups.Page{}, err
	}

	return groups.Page{}, nil
}

func (svc service) AddParentGroup(ctx context.Context, token, id, parentID string) error {
	userInfo, err := svc.identify(ctx, token)
	if err != nil {
		return err
	}

	if err := svc.authorize(ctx, groups.OpAddParentGroup, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		SubjectType: auth.UserType,
		Subject:     userInfo.ID,
		Object:      id,
		ObjectType:  auth.GroupType,
	}); err != nil {
		return err
	}

	if err := svc.authorize(ctx, groups.OpAddChildrenGroups, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		SubjectType: auth.UserType,
		Subject:     userInfo.ID,
		Object:      parentID,
		ObjectType:  auth.GroupType,
	}); err != nil {
		return err
	}

	group, err := svc.groups.RetrieveByID(ctx, id)
	if err != nil {
		return errors.Wrap(svcerr.ErrViewEntity, err)
	}

	var addPolicies magistrala.AddPoliciesReq
	var deletePolicies magistrala.DeletePoliciesReq
	if group.Parent != "" {
		return errors.Wrap(svcerr.ErrConflict, fmt.Errorf("%s group already have parent", group.ID))
	}
	addPolicies.AddPoliciesReq = append(addPolicies.AddPoliciesReq, &magistrala.AddPolicyReq{
		Domain:      userInfo.DomainID,
		SubjectType: auth.GroupType,
		Subject:     parentID,
		Relation:    auth.ParentGroupRelation,
		ObjectType:  auth.GroupType,
		Object:      group.ID,
	})
	deletePolicies.DeletePoliciesReq = append(deletePolicies.DeletePoliciesReq, &magistrala.DeletePolicyReq{
		Domain:      userInfo.DomainID,
		SubjectType: auth.GroupType,
		Subject:     parentID,
		Relation:    auth.ParentGroupRelation,
		ObjectType:  auth.GroupType,
		Object:      group.ID,
	})

	if _, err := svc.auth.AddPolicies(ctx, &addPolicies); err != nil {
		return errors.Wrap(svcerr.ErrAddPolicies, err)
	}
	defer func() {
		if err != nil {
			if _, errRollback := svc.auth.DeletePolicies(ctx, &deletePolicies); errRollback != nil {
				err = errors.Wrap(err, errors.Wrap(apiutil.ErrRollbackTx, errRollback))
			}
		}
	}()

	if err := svc.groups.AssignParentGroup(ctx, parentID, group.ID); err != nil {
		return err
	}
	return nil
}

func (svc service) RemoveParentGroup(ctx context.Context, token, id string) error {
	userInfo, err := svc.identify(ctx, token)
	if err != nil {
		return err
	}

	if err := svc.authorize(ctx, groups.OpRemoveParentGroup, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		SubjectType: auth.UserType,
		Subject:     userInfo.ID,
		Object:      id,
		ObjectType:  auth.GroupType,
	}); err != nil {
		return err
	}

	group, err := svc.groups.RetrieveByID(ctx, id)
	if err != nil {
		return errors.Wrap(svcerr.ErrViewEntity, err)
	}

	if group.Parent != "" {
		if err := svc.authorize(ctx, groups.OpRemoveChildrenGroups, &magistrala.AuthorizeReq{
			Domain:      userInfo.DomainID,
			SubjectType: auth.UserType,
			Subject:     userInfo.ID,
			Object:      group.Parent,
			ObjectType:  auth.GroupType,
		}); err != nil {
			return err
		}

		var addPolicies magistrala.AddPoliciesReq
		var deletePolicies magistrala.DeletePoliciesReq

		addPolicies.AddPoliciesReq = append(addPolicies.AddPoliciesReq, &magistrala.AddPolicyReq{
			Domain:      userInfo.DomainID,
			SubjectType: auth.GroupType,
			Subject:     group.Parent,
			Relation:    auth.ParentGroupRelation,
			ObjectType:  auth.GroupType,
			Object:      group.ID,
		})
		deletePolicies.DeletePoliciesReq = append(deletePolicies.DeletePoliciesReq, &magistrala.DeletePolicyReq{
			Domain:      userInfo.DomainID,
			SubjectType: auth.GroupType,
			Subject:     group.Parent,
			Relation:    auth.ParentGroupRelation,
			ObjectType:  auth.GroupType,
			Object:      group.ID,
		})

		if _, err := svc.auth.DeletePolicies(ctx, &deletePolicies); err != nil {
			return errors.Wrap(svcerr.ErrDeletePolicies, err)
		}
		defer func() {
			if err != nil {
				if _, errRollback := svc.auth.AddPolicies(ctx, &addPolicies); errRollback != nil {
					err = errors.Wrap(err, errors.Wrap(apiutil.ErrRollbackTx, errRollback))
				}
			}
		}()

		return svc.groups.UnassignParentGroup(ctx, group.Parent, group.ID)
	}

	return nil
}

func (svc service) ViewParentGroup(ctx context.Context, token, id string) (groups.Group, error) {
	userInfo, err := svc.identify(ctx, token)
	if err != nil {
		return groups.Group{}, err
	}

	if err := svc.authorize(ctx, groups.OpViewParentGroup, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		SubjectType: auth.UserType,
		Subject:     userInfo.ID,
		Object:      id,
		ObjectType:  auth.GroupType,
	}); err != nil {
		return groups.Group{}, err
	}

	g, err := svc.groups.RetrieveByID(ctx, id)
	if err != nil {
		return groups.Group{}, err
	}
	if err := svc.authorize(ctx, groups.OpViewGroup, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		SubjectType: auth.UserType,
		Subject:     userInfo.ID,
		Object:      g.Parent,
		ObjectType:  auth.GroupType,
	}); err != nil {
		return groups.Group{}, err
	}

	pg, err := svc.groups.RetrieveByID(ctx, g.Parent)
	if err != nil {
		return groups.Group{}, err
	}
	return pg, nil

}

func (svc service) AddChildrenGroups(ctx context.Context, token, parentGroupID string, childrenGroupIDs []string) error {
	userInfo, err := svc.identify(ctx, token)
	if err != nil {
		return err
	}

	if err := svc.authorize(ctx, groups.OpAddChildrenGroups, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		SubjectType: auth.UserType,
		Subject:     userInfo.ID,
		Object:      parentGroupID,
		ObjectType:  auth.GroupType,
	}); err != nil {
		return err
	}

	childrenGroupsPage, err := svc.groups.RetrieveByIDs(ctx, groups.Page{PageMeta: groups.PageMeta{Limit: 1<<63 - 1}}, childrenGroupIDs...)
	if err != nil {
		return errors.Wrap(svcerr.ErrViewEntity, err)
	}
	if len(childrenGroupsPage.Groups) == 0 {
		return errGroupIDs
	}

	for _, childGroup := range childrenGroupsPage.Groups {
		if err := svc.authorize(ctx, groups.OpAddParentGroup, &magistrala.AuthorizeReq{
			Domain:      userInfo.DomainID,
			SubjectType: auth.UserType,
			Subject:     userInfo.ID,
			Object:      childGroup.ID,
			ObjectType:  auth.GroupType,
		}); err != nil {
			return err
		}
	}

	var addPolicies magistrala.AddPoliciesReq
	var deletePolicies magistrala.DeletePoliciesReq
	for _, childGroup := range childrenGroupsPage.Groups {
		if childGroup.Parent != "" {
			return errors.Wrap(svcerr.ErrConflict, fmt.Errorf("%s group already have parent", childGroup.ID))
		}
		addPolicies.AddPoliciesReq = append(addPolicies.AddPoliciesReq, &magistrala.AddPolicyReq{
			Domain:      userInfo.DomainID,
			SubjectType: auth.GroupType,
			Subject:     parentGroupID,
			Relation:    auth.ParentGroupRelation,
			ObjectType:  auth.GroupType,
			Object:      childGroup.ID,
		})
		deletePolicies.DeletePoliciesReq = append(deletePolicies.DeletePoliciesReq, &magistrala.DeletePolicyReq{
			Domain:      userInfo.DomainID,
			SubjectType: auth.GroupType,
			Subject:     parentGroupID,
			Relation:    auth.ParentGroupRelation,
			ObjectType:  auth.GroupType,
			Object:      childGroup.ID,
		})
	}

	if _, err := svc.policy.AddPolicies(ctx, &addPolicies); err != nil {
		return errors.Wrap(svcerr.ErrAddPolicies, err)
	}
	defer func() {
		if err != nil {
			if _, errRollback := svc.policy.DeletePolicies(ctx, &deletePolicies); errRollback != nil {
				err = errors.Wrap(err, errors.Wrap(apiutil.ErrRollbackTx, errRollback))
			}
		}
	}()

	return svc.groups.AssignParentGroup(ctx, parentGroupID, childrenGroupIDs...)
}

func (svc service) RemoveChildrenGroups(ctx context.Context, token, parentGroupID string, childrenGroupIDs []string) error {
	userInfo, err := svc.identify(ctx, token)
	if err != nil {
		return err
	}

	if err := svc.authorize(ctx, groups.OpRemoveChildrenGroups, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		SubjectType: auth.UserType,
		Subject:     userInfo.ID,
		Object:      parentGroupID,
		ObjectType:  auth.GroupType,
	}); err != nil {
		return err
	}

	childrenGroupsPage, err := svc.groups.RetrieveByIDs(ctx, groups.Page{PageMeta: groups.PageMeta{Limit: 1<<63 - 1}}, childrenGroupIDs...)
	if err != nil {
		return errors.Wrap(svcerr.ErrViewEntity, err)
	}
	if len(childrenGroupsPage.Groups) == 0 {
		return errGroupIDs
	}

	var addPolicies magistrala.AddPoliciesReq
	var deletePolicies magistrala.DeletePoliciesReq
	for _, group := range childrenGroupsPage.Groups {
		if group.Parent != "" && group.Parent != parentGroupID {
			return errors.Wrap(svcerr.ErrConflict, fmt.Errorf("%s group doesn't have same parent", group.ID))
		}
		addPolicies.AddPoliciesReq = append(addPolicies.AddPoliciesReq, &magistrala.AddPolicyReq{
			Domain:      userInfo.DomainID,
			SubjectType: auth.GroupType,
			Subject:     parentGroupID,
			Relation:    auth.ParentGroupRelation,
			ObjectType:  auth.GroupType,
			Object:      group.ID,
		})
		deletePolicies.DeletePoliciesReq = append(deletePolicies.DeletePoliciesReq, &magistrala.DeletePolicyReq{
			Domain:      userInfo.DomainID,
			SubjectType: auth.GroupType,
			Subject:     parentGroupID,
			Relation:    auth.ParentGroupRelation,
			ObjectType:  auth.GroupType,
			Object:      group.ID,
		})
	}

	if _, err := svc.policy.DeletePolicies(ctx, &deletePolicies); err != nil {
		return errors.Wrap(svcerr.ErrDeletePolicies, err)
	}
	defer func() {
		if err != nil {
			if _, errRollback := svc.policy.AddPolicies(ctx, &addPolicies); errRollback != nil {
				err = errors.Wrap(err, errors.Wrap(apiutil.ErrRollbackTx, errRollback))
			}
		}
	}()

	return svc.groups.UnassignParentGroup(ctx, parentGroupID, childrenGroupIDs...)
}

func (svc service) RemoveAllChildrenGroups(ctx context.Context, token, id string) error {
	userInfo, err := svc.identify(ctx, token)
	if err != nil {
		return err
	}

	if err := svc.authorize(ctx, groups.OpRemoveAllChildrenGroups, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		SubjectType: auth.UserType,
		Subject:     userInfo.ID,
		Object:      id,
		ObjectType:  auth.GroupType,
	}); err != nil {
		return err
	}

	policy := magistrala.DeletePolicyFilterReq{
		Domain:      userInfo.DomainID,
		SubjectType: auth.GroupType,
		Subject:     id,
		Relation:    auth.ParentGroupRelation,
		ObjectType:  auth.GroupType,
	}

	if _, err := svc.auth.DeletePolicyFilter(ctx, &policy); err != nil {
		return errors.Wrap(svcerr.ErrDeletePolicies, err)
	}

	return svc.groups.UnassignAllChildrenGroup(ctx, id)
}

func (svc service) ListChildrenGroups(ctx context.Context, token, id string, gm groups.Page) (groups.Page, error) {
	userInfo, err := svc.identify(ctx, token)
	if err != nil {
		return groups.Page{}, err
	}

	if err := svc.authorize(ctx, groups.OpListChildrenGroups, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		SubjectType: auth.UserType,
		Subject:     userInfo.ID,
		Object:      id,
		ObjectType:  auth.GroupType,
	}); err != nil {
		return groups.Page{}, err
	}
	return groups.Page{}, nil
}

// func (svc service) AddChannels(ctx context.Context, token, id string, channelIDs []string) error {
// 	userInfo, err := svc.identify(ctx, token)
// 	if err != nil {
// 		return err
// 	}

// 	if err := svc.authorize(ctx, groups.OpAddChannels, &magistrala.AuthorizeReq{
// 		Domain:      userInfo.DomainID,
// 		SubjectType: auth.UserType,
// 		Subject:     userInfo.ID,
// 		Object:      id,
// 		ObjectType:  auth.GroupType,
// 	}); err != nil {
// 		return err
// 	}

// 	policies := magistrala.AddPoliciesReq{}

// 	for _, channelID := range channelIDs {
// 		policies.AddPoliciesReq = append(policies.AddPoliciesReq, &magistrala.AddPolicyReq{
// 			Domain:      userInfo.DomainID,
// 			SubjectType: auth.GroupType,
// 			SubjectKind: auth.ChannelsKind,
// 			Subject:     id,
// 			Relation:    auth.ParentGroupRelation,
// 			ObjectType:  auth.ThingType,
// 			Object:      channelID,
// 		})
// 	}

// 	if _, err := svc.auth.AddPolicies(ctx, &policies); err != nil {
// 		return errors.Wrap(svcerr.ErrAddPolicies, err)
// 	}

// 	return nil
// }

// func (svc service) RemoveChannels(ctx context.Context, token, id string, channelIDs []string) error {
// 	userInfo, err := svc.identify(ctx, token)
// 	if err != nil {
// 		return err
// 	}

// 	if err := svc.authorize(ctx, groups.OpAddChannels, &magistrala.AuthorizeReq{
// 		Domain:      userInfo.DomainID,
// 		SubjectType: auth.UserType,
// 		Subject:     userInfo.ID,
// 		Object:      id,
// 		ObjectType:  auth.GroupType,
// 	}); err != nil {
// 		return err
// 	}
// 	policies := magistrala.DeletePoliciesReq{}

// 	for _, channelID := range channelIDs {
// 		policies.DeletePoliciesReq = append(policies.DeletePoliciesReq, &magistrala.DeletePolicyReq{
// 			Domain:      userInfo.DomainID,
// 			SubjectType: auth.GroupType,
// 			Subject:     id,
// 			Relation:    auth.ParentGroupRelation,
// 			ObjectType:  auth.ChannelType,
// 			Object:      channelID,
// 		})
// 	}
// 	if _, err := svc.auth.DeletePolicies(ctx, &policies); err != nil {
// 		return errors.Wrap(svcerr.ErrDeletePolicies, err)
// 	}

// 	return nil
// }

// func (svc service) ListChannels(ctx context.Context, token, id, gm groups.Page) (groups.Page, error) {
// 	return groups.Page{}, nil
// }

// func (svc service) AddThings(ctx context.Context, token, id string, thingIDs []string) error {
// 	userInfo, err := svc.identify(ctx, token)
// 	if err != nil {
// 		return err
// 	}

// 	if err := svc.authorize(ctx, groups.OpAddChannels, &magistrala.AuthorizeReq{
// 		Domain:      userInfo.DomainID,
// 		SubjectType: auth.UserType,
// 		Subject:     userInfo.ID,
// 		Object:      id,
// 		ObjectType:  auth.GroupType,
// 	}); err != nil {
// 		return err
// 	}
// 	policies := magistrala.AddPoliciesReq{}

// 	for _, thingID := range thingIDs {
// 		policies.AddPoliciesReq = append(policies.AddPoliciesReq, &magistrala.AddPolicyReq{
// 			Domain:      userInfo.DomainID,
// 			SubjectType: auth.GroupType,
// 			SubjectKind: auth.ChannelsKind,
// 			Subject:     id,
// 			Relation:    auth.ParentGroupRelation,
// 			ObjectType:  auth.ThingType,
// 			Object:      thingID,
// 		})
// 	}

// 	if _, err := svc.auth.AddPolicies(ctx, &policies); err != nil {
// 		return errors.Wrap(svcerr.ErrAddPolicies, err)
// 	}

// 	return nil
// }

// func (svc service) RemoveThings(ctx context.Context, token, id string, thingIDs []string) error {
// 	userInfo, err := svc.identify(ctx, token)
// 	if err != nil {
// 		return err
// 	}

// 	if err := svc.authorize(ctx, groups.OpRemoveAllChannels, &magistrala.AuthorizeReq{
// 		Domain:      userInfo.DomainID,
// 		SubjectType: auth.UserType,
// 		Subject:     userInfo.ID,
// 		Object:      id,
// 		ObjectType:  auth.GroupType,
// 	}); err != nil {
// 		return err
// 	}
// 	policies := magistrala.DeletePoliciesReq{}

// 	for _, thingID := range thingIDs {
// 		policies.DeletePoliciesReq = append(policies.DeletePoliciesReq, &magistrala.DeletePolicyReq{
// 			Domain:      userInfo.DomainID,
// 			SubjectType: auth.GroupType,
// 			Subject:     id,
// 			Relation:    auth.ParentGroupRelation,
// 			ObjectType:  auth.ThingType,
// 			Object:      thingID,
// 		})
// 	}
// 	if _, err := svc.auth.DeletePolicies(ctx, &policies); err != nil {
// 		return errors.Wrap(svcerr.ErrDeletePolicies, err)
// 	}

// 	return nil
// }

// func (svc service) RemoveAllThings(ctx context.Context, token, id string) error {
// 	userInfo, err := svc.identify(ctx, token)
// 	if err != nil {
// 		return err
// 	}

// 	if err := svc.authorize(ctx, groups.OpRemoveAllThings, &magistrala.AuthorizeReq{
// 		Domain:      userInfo.DomainID,
// 		SubjectType: auth.UserType,
// 		Subject:     userInfo.ID,
// 		Object:      id,
// 		ObjectType:  auth.GroupType,
// 	}); err != nil {
// 		return err
// 	}

// 	policy := magistrala.DeletePolicyFilterReq{
// 		Domain:      userInfo.DomainID,
// 		SubjectType: auth.GroupType,
// 		Subject:     id,
// 		Relation:    auth.ParentGroupRelation,
// 		ObjectType:  auth.ThingType,
// 	}

// 	if _, err := svc.auth.DeletePolicyFilter(ctx, &policy); err != nil {
// 		return errors.Wrap(svcerr.ErrDeletePolicies, err)
// 	}
// 	return nil
// }

// func (svc service) ListThings(ctx context.Context, token, id, gm groups.Page) (groups.Page, error) {
// 	return groups.Page{}, nil
// }

func (svc service) DeleteGroup(ctx context.Context, token, id string) error {
	userInfo, err := svc.identify(ctx, token)
	if err != nil {
		return err
	}

	if err := svc.authorize(ctx, groups.OpDeleteGroup, &magistrala.AuthorizeReq{
		Domain:      userInfo.DomainID,
		SubjectType: auth.UserType,
		Subject:     userInfo.ID,
		Object:      id,
		ObjectType:  auth.GroupType,
	}); err != nil {
		return err
	}

	deleteRes, err := svc.policy.DeleteEntityPolicies(ctx, &magistrala.DeleteEntityPoliciesReq{
		EntityType: auth.GroupType,
		Id:         id,
	})
	if err != nil {
		return errors.Wrap(svcerr.ErrDeletePolicies, err)
	}
	if !deleteRes.Deleted {
		return svcerr.ErrAuthorization
	}

	if err := svc.groups.Delete(ctx, id); err != nil {
		return err
	}

	return nil
}

func (svc service) filterAllowedGroupIDsOfUserID(ctx context.Context, userID, permission string, groupIDs []string) ([]string, error) {
	var ids []string
	allowedIDs, err := svc.listAllGroupsOfUserID(ctx, userID, permission)
	if err != nil {
		return []string{}, err
	}

	for _, gid := range groupIDs {
		for _, id := range allowedIDs {
			if id == gid {
				ids = append(ids, id)
			}
		}
	}
	return ids, nil
}

func (svc service) listAllGroupsOfUserID(ctx context.Context, userID, permission string) ([]string, error) {
	allowedIDs, err := svc.policy.ListAllObjects(ctx, &magistrala.ListObjectsReq{
		SubjectType: auth.UserType,
		Subject:     userID,
		Permission:  permission,
		ObjectType:  auth.GroupType,
	})
	if err != nil {
		return []string{}, err
	}
	return allowedIDs.Policies, nil
}

func (svc service) changeGroupStatus(ctx context.Context, userInfo identity, group groups.Group) (groups.Group, error) {
	dbGroup, err := svc.groups.RetrieveByID(ctx, group.ID)
	if err != nil {
		return groups.Group{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	if dbGroup.Status == group.Status {
		return groups.Group{}, errors.ErrStatusAlreadyAssigned
	}

	group.UpdatedBy = userInfo.UserID
	return svc.groups.ChangeStatus(ctx, group)
}

// func (svc service) authorizeToken(ctx context.Context, subjectType, subject, permission, objectType, object string) (string, error) {
// 	req := &magistrala.AuthorizeReq{
// 		SubjectType: subjectType,
// 		SubjectKind: auth.TokenKind,
// 		Subject:     subject,
// 		Permission:  permission,
// 		Object:      object,
// 		ObjectType:  objectType,
// 	}
// 	res, err := svc.auth.Authorize(ctx, req)
// 	if err != nil {
// 		return "", errors.Wrap(svcerr.ErrAuthorization, err)
// 	}
// 	if !res.GetAuthorized() {
// 		return "", svcerr.ErrAuthorization
// 	}
// 	return res.GetId(), nil
// }

// func (svc service) authorizeKind(ctx context.Context, domainID, subjectType, subjectKind, subject, permission, objectType, object string) (string, error) {
// 	req := &magistrala.AuthorizeReq{
// 		Domain:      domainID,
// 		SubjectType: subjectType,
// 		SubjectKind: subjectKind,
// 		Subject:     subject,
// 		Permission:  permission,
// 		Object:      object,
// 		ObjectType:  objectType,
// 	}
// 	res, err := svc.auth.Authorize(ctx, req)
// 	if err != nil {
// 		return "", errors.Wrap(svcerr.ErrAuthorization, err)
// 	}
// 	if !res.GetAuthorized() {
// 		return "", svcerr.ErrAuthorization
// 	}
// 	return res.GetId(), nil
// }

func (svc service) addGroupPolicy(ctx context.Context, userID, domainID, id, parentID, kind string) error {
	policies := magistrala.AddPoliciesReq{}
	policies.AddPoliciesReq = append(policies.AddPoliciesReq, &magistrala.AddPolicyReq{
		Domain:      domainID,
		SubjectType: auth.UserType,
		Subject:     userID,
		Relation:    auth.AdministratorRelation,
		ObjectKind:  kind,
		ObjectType:  auth.GroupType,
		Object:      id,
	})
	policies.AddPoliciesReq = append(policies.AddPoliciesReq, &magistrala.AddPolicyReq{
		Domain:      domainID,
		SubjectType: auth.DomainType,
		Subject:     domainID,
		Relation:    auth.DomainRelation,
		ObjectType:  auth.GroupType,
		Object:      id,
	})
	if parentID != "" {
		policies.AddPoliciesReq = append(policies.AddPoliciesReq, &magistrala.AddPolicyReq{
			Domain:      domainID,
			SubjectType: auth.GroupType,
			Subject:     parentID,
			Relation:    auth.ParentGroupRelation,
			ObjectKind:  kind,
			ObjectType:  auth.GroupType,
			Object:      id,
		})
	}
	if _, err := svc.policy.AddPolicies(ctx, &policies); err != nil {
		return errors.Wrap(svcerr.ErrAddPolicies, err)
	}

	return nil
}

// func (svc service) addGroupPolicyRollback(ctx context.Context, userID, domainID, id, parentID, kind string) error {
// 	policies := magistrala.DeletePoliciesReq{}
// 	policies.DeletePoliciesReq = append(policies.DeletePoliciesReq, &magistrala.DeletePolicyReq{
// 		Domain:      domainID,
// 		SubjectType: auth.UserType,
// 		Subject:     userID,
// 		Relation:    auth.AdministratorRelation,
// 		ObjectKind:  kind,
// 		ObjectType:  auth.GroupType,
// 		Object:      id,
// 	})
// 	policies.DeletePoliciesReq = append(policies.DeletePoliciesReq, &magistrala.DeletePolicyReq{
// 		Domain:      domainID,
// 		SubjectType: auth.DomainType,
// 		Subject:     domainID,
// 		Relation:    auth.DomainRelation,
// 		ObjectType:  auth.GroupType,
// 		Object:      id,
// 	})
// 	if parentID != "" {
// 		policies.DeletePoliciesReq = append(policies.DeletePoliciesReq, &magistrala.DeletePolicyReq{
// 			Domain:      domainID,
// 			SubjectType: auth.GroupType,
// 			Subject:     parentID,
// 			Relation:    auth.ParentGroupRelation,
// 			ObjectKind:  kind,
// 			ObjectType:  auth.GroupType,
// 			Object:      id,
// 		})
// 	}
// 	if _, err := svc.auth.DeletePolicies(ctx, &policies); err != nil {
// 		return errors.Wrap(svcerr.ErrDeletePolicies, err)
// 	}

// 	return nil
// }

func (svc service) identify(ctx context.Context, token string) (identity, error) {
	resp, err := svc.auth.Identify(ctx, &magistrala.IdentityReq{Token: token})
	if err != nil {
		return identity{}, errors.Wrap(svcerr.ErrAuthentication, err)
	}
	return identity{ID: resp.GetId(), DomainID: resp.GetDomainId(), UserID: resp.GetUserId()}, nil
}

func (svc service) authorize(ctx context.Context, op svcutil.Operation, authReq *magistrala.AuthorizeReq) error {
	perm, err := svc.opp.GetPermission(op)
	if err != nil {
		return err
	}
	authReq.Permission = perm.String()
	resp, err := svc.auth.Authorize(ctx, authReq)
	if err != nil {
		return errors.Wrap(svcerr.ErrAuthorization, err)
	}
	if !resp.Authorized {
		return errors.Wrap(svcerr.ErrAuthorization, err)
	}
	return nil
}
