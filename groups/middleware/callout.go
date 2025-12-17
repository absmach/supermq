// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"time"

	"github.com/absmach/supermq/domains"
	"github.com/absmach/supermq/groups"
	"github.com/absmach/supermq/pkg/authn"
	"github.com/absmach/supermq/pkg/callout"
	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	"github.com/absmach/supermq/pkg/permissions"
	"github.com/absmach/supermq/pkg/policies"
	"github.com/absmach/supermq/pkg/roles"
	rolemgr "github.com/absmach/supermq/pkg/roles/rolemanager/middleware"
)

var _ groups.Service = (*calloutMiddleware)(nil)

type calloutMiddleware struct {
	svc         groups.Service
	repo        groups.Repository
	callout     callout.Callout
	entitiesOps permissions.EntitiesOperations[permissions.Operation]
	rolemgr.RoleManagerCalloutMiddleware
}

func NewCallout(svc groups.Service, repo groups.Repository, entitiesOps permissions.EntitiesOperations[permissions.Operation], roleOps permissions.Operations[permissions.RoleOperation], callout callout.Callout) (groups.Service, error) {
	call, err := rolemgr.NewCallout(policies.GroupType, svc, callout, roleOps)
	if err != nil {
		return nil, err
	}

	if err := entitiesOps.Validate(); err != nil {
		return nil, err
	}

	return &calloutMiddleware{
		svc:                          svc,
		repo:                         repo,
		callout:                      callout,
		entitiesOps:                  entitiesOps,
		RoleManagerCalloutMiddleware: call,
	}, nil
}

func (cm *calloutMiddleware) CreateGroup(ctx context.Context, session authn.Session, g groups.Group) (groups.Group, []roles.RoleProvision, error) {
	params := map[string]any{
		"entities": []groups.Group{g},
		"count":    1,
	}

	if err := cm.callOut(ctx, session, policies.DomainType, domains.OpCreateDomainGroups, params); err != nil {
		return groups.Group{}, nil, err
	}

	return cm.svc.CreateGroup(ctx, session, g)
}

func (cm *calloutMiddleware) UpdateGroup(ctx context.Context, session authn.Session, group groups.Group) (groups.Group, error) {
	params := map[string]any{
		"entity_id": group.ID,
		"group":     group,
	}

	if err := cm.callOut(ctx, session, policies.GroupType, groups.OpUpdateGroup, params); err != nil {
		return groups.Group{}, err
	}

	return cm.svc.UpdateGroup(ctx, session, group)
}

func (cm *calloutMiddleware) UpdateGroupTags(ctx context.Context, session authn.Session, group groups.Group) (groups.Group, error) {
	params := map[string]any{
		"entity_id": group.ID,
		"tags":      group.Tags,
	}

	if err := cm.callOut(ctx, session, policies.GroupType, groups.OpUpdateGroupTags, params); err != nil {
		return groups.Group{}, err
	}

	return cm.svc.UpdateGroupTags(ctx, session, group)
}

func (cm *calloutMiddleware) ViewGroup(ctx context.Context, session authn.Session, id string, withRoles bool) (groups.Group, error) {
	params := map[string]any{
		"entity_id": id,
	}

	if err := cm.callOut(ctx, session, policies.GroupType, groups.OpViewGroup, params); err != nil {
		return groups.Group{}, err
	}

	return cm.svc.ViewGroup(ctx, session, id, withRoles)
}

func (cm *calloutMiddleware) ListGroups(ctx context.Context, session authn.Session, gm groups.PageMeta) (groups.Page, error) {
	params := map[string]any{
		"pagemeta": gm,
	}

	if err := cm.callOut(ctx, session, policies.DomainType, domains.OpListDomainGroups, params); err != nil {
		return groups.Page{}, err
	}

	return cm.svc.ListGroups(ctx, session, gm)
}

func (cm *calloutMiddleware) ListUserGroups(ctx context.Context, session authn.Session, userID string, gm groups.PageMeta) (groups.Page, error) {
	params := map[string]any{
		"user_id":  userID,
		"pagemeta": gm,
	}

	if err := cm.callOut(ctx, session, policies.GroupType, groups.OpListUserGroups, params); err != nil {
		return groups.Page{}, err
	}

	return cm.svc.ListUserGroups(ctx, session, userID, gm)
}

func (cm *calloutMiddleware) EnableGroup(ctx context.Context, session authn.Session, id string) (groups.Group, error) {
	params := map[string]any{
		"entity_id": id,
	}

	if err := cm.callOut(ctx, session, policies.GroupType, groups.OpEnableGroup, params); err != nil {
		return groups.Group{}, err
	}

	return cm.svc.EnableGroup(ctx, session, id)
}

func (cm *calloutMiddleware) DisableGroup(ctx context.Context, session authn.Session, id string) (groups.Group, error) {
	params := map[string]any{
		"entity_id": id,
	}

	if err := cm.callOut(ctx, session, policies.GroupType, groups.OpDisableGroup, params); err != nil {
		return groups.Group{}, err
	}

	return cm.svc.DisableGroup(ctx, session, id)
}

func (cm *calloutMiddleware) DeleteGroup(ctx context.Context, session authn.Session, id string) error {
	params := map[string]any{
		"entity_id": id,
	}

	if err := cm.callOut(ctx, session, policies.GroupType, groups.OpDeleteGroup, params); err != nil {
		return err
	}

	return cm.svc.DeleteGroup(ctx, session, id)
}

func (cm *calloutMiddleware) RetrieveGroupHierarchy(ctx context.Context, session authn.Session, id string, hm groups.HierarchyPageMeta) (groups.HierarchyPage, error) {
	params := map[string]any{
		"entity_id":          id,
		"hierarchy_pagemeta": hm,
	}

	if err := cm.callOut(ctx, session, policies.GroupType, groups.OpRetrieveGroupHierarchy, params); err != nil {
		return groups.HierarchyPage{}, err
	}

	return cm.svc.RetrieveGroupHierarchy(ctx, session, id, hm)
}

func (cm *calloutMiddleware) AddParentGroup(ctx context.Context, session authn.Session, id, parentID string) error {
	params := map[string]any{
		"entity_id": id,
		"parent_id": parentID,
	}

	if err := cm.callOut(ctx, session, policies.GroupType, groups.OpAddParentGroup, params); err != nil {
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
		"entity_id": id,
		"parent_id": group.Parent,
	}

	if err := cm.callOut(ctx, session, policies.GroupType, groups.OpRemoveParentGroup, params); err != nil {
		return err
	}

	return cm.svc.RemoveParentGroup(ctx, session, id)
}

func (cm *calloutMiddleware) AddChildrenGroups(ctx context.Context, session authn.Session, id string, childrenGroupIDs []string) error {
	params := map[string]any{
		"entity_id":          id,
		"children_group_ids": childrenGroupIDs,
	}

	if err := cm.callOut(ctx, session, policies.GroupType, groups.OpAddChildrenGroups, params); err != nil {
		return err
	}

	return cm.svc.AddChildrenGroups(ctx, session, id, childrenGroupIDs)
}

func (cm *calloutMiddleware) RemoveChildrenGroups(ctx context.Context, session authn.Session, id string, childrenGroupIDs []string) error {
	params := map[string]any{
		"entity_id":          id,
		"children_group_ids": childrenGroupIDs,
	}

	if err := cm.callOut(ctx, session, policies.GroupType, groups.OpRemoveChildrenGroups, params); err != nil {
		return err
	}

	return cm.svc.RemoveChildrenGroups(ctx, session, id, childrenGroupIDs)
}

func (cm *calloutMiddleware) RemoveAllChildrenGroups(ctx context.Context, session authn.Session, id string) error {
	params := map[string]any{
		"entity_id": id,
	}

	if err := cm.callOut(ctx, session, policies.GroupType, groups.OpRemoveAllChildrenGroups, params); err != nil {
		return err
	}

	return cm.svc.RemoveAllChildrenGroups(ctx, session, id)
}

func (cm *calloutMiddleware) ListChildrenGroups(ctx context.Context, session authn.Session, id string, startLevel, endLevel int64, pm groups.PageMeta) (groups.Page, error) {
	params := map[string]any{
		"entity_id":   id,
		"start_level": startLevel,
		"end_level":   endLevel,
		"pagemeta":    pm,
	}

	if err := cm.callOut(ctx, session, policies.GroupType, groups.OpListChildrenGroups, params); err != nil {
		return groups.Page{}, err
	}

	return cm.svc.ListChildrenGroups(ctx, session, id, startLevel, endLevel, pm)
}

func (cm *calloutMiddleware) callOut(ctx context.Context, session authn.Session, entityType string, op permissions.Operation, pld map[string]any) error {
	var entityID string
	if id, ok := pld["entity_id"].(string); ok {
		entityID = id
	}

	req := callout.Request{
		BaseRequest: callout.BaseRequest{
			Operation:  cm.entitiesOps.OperationName(entityType, op),
			EntityType: entityType,
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
