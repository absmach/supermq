// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"time"

	"github.com/absmach/supermq/groups"
	"github.com/absmach/supermq/pkg/authn"
	"github.com/absmach/supermq/pkg/callout"
	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	"github.com/absmach/supermq/pkg/policies"
	"github.com/absmach/supermq/pkg/roles"
	rolemw "github.com/absmach/supermq/pkg/roles/rolemanager/middleware"
)

var _ groups.Service = (*calloutMiddleware)(nil)

type calloutMiddleware struct {
	svc     groups.Service
	repo    groups.Repository
	callout callout.Callout
	rolemw.RoleManagerCalloutMiddleware
}

func NewCallout(svc groups.Service, repo groups.Repository, callout callout.Callout) (groups.Service, error) {
	call, err := rolemw.NewCallout(policies.ClientType, svc, callout)
	if err != nil {
		return nil, err
	}

	return &calloutMiddleware{
		svc:                          svc,
		repo:                         repo,
		callout:                      callout,
		RoleManagerCalloutMiddleware: call,
	}, nil
}

func (cm *calloutMiddleware) CreateGroup(ctx context.Context, session authn.Session, g groups.Group) (groups.Group, []roles.RoleProvision, error) {
	params := map[string]any{
		"entities": []groups.Group{g},
		"count":    1,
	}

	if err := cm.callOut(ctx, session, groups.OpCreateGroup.String(groups.OperationNames), "", params); err != nil {
		return groups.Group{}, nil, err
	}

	return cm.svc.CreateGroup(ctx, session, g)
}

func (cm *calloutMiddleware) UpdateGroup(ctx context.Context, session authn.Session, group groups.Group) (groups.Group, error) {
	params := map[string]any{
		"group": group,
	}

	if err := cm.callOut(ctx, session, groups.OpUpdateGroup.String(groups.OperationNames), group.ID, params); err != nil {
		return groups.Group{}, err
	}

	return cm.svc.UpdateGroup(ctx, session, group)
}

func (cm *calloutMiddleware) UpdateGroupTags(ctx context.Context, session authn.Session, group groups.Group) (groups.Group, error) {
	params := map[string]any{
		"tags": group.Tags,
	}

	if err := cm.callOut(ctx, session, groups.OpUpdateGroupTags.String(groups.OperationNames), group.ID, params); err != nil {
		return groups.Group{}, err
	}

	return cm.svc.UpdateGroupTags(ctx, session, group)
}

func (cm *calloutMiddleware) ViewGroup(ctx context.Context, session authn.Session, id string, withRoles bool) (groups.Group, error) {
	if err := cm.callOut(ctx, session, groups.OpViewGroup.String(groups.OperationNames), id, nil); err != nil {
		return groups.Group{}, err
	}

	return cm.svc.ViewGroup(ctx, session, id, withRoles)
}

func (cm *calloutMiddleware) ListGroups(ctx context.Context, session authn.Session, gm groups.PageMeta) (groups.Page, error) {
	params := map[string]any{
		"pagemeta": gm,
	}

	if err := cm.callOut(ctx, session, groups.OpListGroups.String(groups.OperationNames), "", params); err != nil {
		return groups.Page{}, err
	}

	return cm.svc.ListGroups(ctx, session, gm)
}

func (cm *calloutMiddleware) ListUserGroups(ctx context.Context, session authn.Session, userID string, gm groups.PageMeta) (groups.Page, error) {
	params := map[string]any{
		"user_id":  userID,
		"pagemeta": gm,
	}

	if err := cm.callOut(ctx, session, groups.OpListUserGroups.String(groups.OperationNames), "", params); err != nil {
		return groups.Page{}, err
	}

	return cm.svc.ListUserGroups(ctx, session, userID, gm)
}

func (cm *calloutMiddleware) EnableGroup(ctx context.Context, session authn.Session, id string) (groups.Group, error) {
	if err := cm.callOut(ctx, session, groups.OpEnableGroup.String(groups.OperationNames), id, nil); err != nil {
		return groups.Group{}, err
	}

	return cm.svc.EnableGroup(ctx, session, id)
}

func (cm *calloutMiddleware) DisableGroup(ctx context.Context, session authn.Session, id string) (groups.Group, error) {
	if err := cm.callOut(ctx, session, groups.OpDisableGroup.String(groups.OperationNames), id, nil); err != nil {
		return groups.Group{}, err
	}

	return cm.svc.DisableGroup(ctx, session, id)
}

func (cm *calloutMiddleware) DeleteGroup(ctx context.Context, session authn.Session, id string) error {
	if err := cm.callOut(ctx, session, groups.OpDeleteGroup.String(groups.OperationNames), id, nil); err != nil {
		return err
	}

	return cm.svc.DeleteGroup(ctx, session, id)
}

func (cm *calloutMiddleware) RetrieveGroupHierarchy(ctx context.Context, session authn.Session, id string, hm groups.HierarchyPageMeta) (groups.HierarchyPage, error) {
	params := map[string]any{
		"hierarchy_pagemeta": hm,
	}

	if err := cm.callOut(ctx, session, groups.OpRetrieveGroupHierarchy.String(groups.OperationNames), id, params); err != nil {
		return groups.HierarchyPage{}, err
	}

	return cm.svc.RetrieveGroupHierarchy(ctx, session, id, hm)
}

func (cm *calloutMiddleware) AddParentGroup(ctx context.Context, session authn.Session, id, parentID string) error {
	params := map[string]any{
		"parent_id": parentID,
	}

	if err := cm.callOut(ctx, session, groups.OpAddParentGroup.String(groups.OperationNames), id, params); err != nil {
		return err
	}

	return cm.svc.AddParentGroup(ctx, session, id, parentID)
}

func (cm *calloutMiddleware) RemoveParentGroup(ctx context.Context, session authn.Session, id string) error {
	group, err := cm.repo.RetrieveByID(ctx, id)
	if err != nil {
		return errors.Wrap(svcerr.ErrViewEntity, err)
	}

	params := map[string]any{
		"parent_id": group.Parent,
	}

	if err := cm.callOut(ctx, session, groups.OpRemoveParentGroup.String(groups.OperationNames), id, params); err != nil {
		return err
	}

	return cm.svc.RemoveParentGroup(ctx, session, id)
}

func (cm *calloutMiddleware) AddChildrenGroups(ctx context.Context, session authn.Session, id string, childrenGroupIDs []string) error {
	params := map[string]any{
		"children_group_ids": childrenGroupIDs,
	}

	if err := cm.callOut(ctx, session, groups.OpAddChildrenGroups.String(groups.OperationNames), id, params); err != nil {
		return err
	}

	return cm.svc.AddChildrenGroups(ctx, session, id, childrenGroupIDs)
}

func (cm *calloutMiddleware) RemoveChildrenGroups(ctx context.Context, session authn.Session, id string, childrenGroupIDs []string) error {
	params := map[string]any{
		"children_group_ids": childrenGroupIDs,
	}

	if err := cm.callOut(ctx, session, groups.OpRemoveChildrenGroups.String(groups.OperationNames), id, params); err != nil {
		return err
	}

	return cm.svc.RemoveChildrenGroups(ctx, session, id, childrenGroupIDs)
}

func (cm *calloutMiddleware) RemoveAllChildrenGroups(ctx context.Context, session authn.Session, id string) error {
	if err := cm.callOut(ctx, session, groups.OpRemoveAllChildrenGroups.String(groups.OperationNames), id, nil); err != nil {
		return err
	}

	return cm.svc.RemoveAllChildrenGroups(ctx, session, id)
}

func (cm *calloutMiddleware) ListChildrenGroups(ctx context.Context, session authn.Session, id string, startLevel, endLevel int64, pm groups.PageMeta) (groups.Page, error) {
	params := map[string]any{
		"start_level": startLevel,
		"end_level":   endLevel,
		"pagemeta":    pm,
	}

	if err := cm.callOut(ctx, session, groups.OpListChildrenGroups.String(groups.OperationNames), id, params); err != nil {
		return groups.Page{}, err
	}

	return cm.svc.ListChildrenGroups(ctx, session, id, startLevel, endLevel, pm)
}

func (cm *calloutMiddleware) callOut(ctx context.Context, session authn.Session, op, entityID string, pld map[string]any) error {
	req := callout.Request{
		BaseRequest: callout.BaseRequest{
			Operation:  op,
			EntityType: policies.GroupType,
			EntityID:   entityID,
			CallerID:   session.UserID,
			CallerType: policies.UserType,
			DomainID:   session.DomainID,
			Time:       time.Now().UTC(),
		},
		Payload: pld,
	}

	if err := cm.callout.Callout(ctx, req); err != nil {
		return err
	}

	return nil
}
