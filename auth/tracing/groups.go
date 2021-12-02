// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// Package tracing contains middlewares that will add spans to existing traces.
package tracing

import (
	"context"

	"github.com/mainflux/mainflux/internal/groups"
	opentracing "github.com/opentracing/opentracing-go"
)

const (
	assign              = "assign"
	saveGroup           = "save_group"
	deleteGroup         = "delete_group"
	updateGroup         = "update_group"
	retrieveByID        = "retrieve_by_id"
	retrieveAllParents  = "retrieve_all_parents"
	retrieveAllChildren = "retrieve_all_children"
	retrieveAll         = "retrieve_all_groups"
	memberships         = "memberships"
	members             = "members"
	unassign            = "unassign"
)

var _ groups.GroupRepository = (*groupRepositoryMiddleware)(nil)

type groupRepositoryMiddleware struct {
	tracer opentracing.Tracer
	repo   groups.GroupRepository
}

// GroupRepositoryMiddleware tracks request and their latency, and adds spans to context.
func GroupRepositoryMiddleware(tracer opentracing.Tracer, gr groups.GroupRepository) groups.GroupRepository {
	return groupRepositoryMiddleware{
		tracer: tracer,
		repo:   gr,
	}
}

func (grm groupRepositoryMiddleware) Save(ctx context.Context, g groups.Group) (groups.Group, error) {
	span := createSpan(ctx, grm.tracer, saveGroup)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.Save(ctx, g)
}

func (grm groupRepositoryMiddleware) Update(ctx context.Context, g groups.Group) (groups.Group, error) {
	span := createSpan(ctx, grm.tracer, updateGroup)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.Update(ctx, g)
}

func (grm groupRepositoryMiddleware) Delete(ctx context.Context, groupID string) error {
	span := createSpan(ctx, grm.tracer, deleteGroup)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.Delete(ctx, groupID)
}

func (grm groupRepositoryMiddleware) RetrieveByID(ctx context.Context, id string) (groups.Group, error) {
	span := createSpan(ctx, grm.tracer, retrieveByID)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.RetrieveByID(ctx, id)
}

func (grm groupRepositoryMiddleware) RetrieveAllParents(ctx context.Context, groupID string, pm groups.PageMetadata) (groups.GroupPage, error) {
	span := createSpan(ctx, grm.tracer, retrieveAllParents)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.RetrieveAllParents(ctx, groupID, pm)
}

func (grm groupRepositoryMiddleware) RetrieveAllChildren(ctx context.Context, groupID string, pm groups.PageMetadata) (groups.GroupPage, error) {
	span := createSpan(ctx, grm.tracer, retrieveAllChildren)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.RetrieveAllChildren(ctx, groupID, pm)
}

func (grm groupRepositoryMiddleware) RetrieveAll(ctx context.Context, pm groups.PageMetadata) (groups.GroupPage, error) {
	span := createSpan(ctx, grm.tracer, retrieveAll)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.RetrieveAll(ctx, pm)
}

func (grm groupRepositoryMiddleware) Memberships(ctx context.Context, memberID string, pm groups.PageMetadata) (groups.GroupPage, error) {
	span := createSpan(ctx, grm.tracer, memberships)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.Memberships(ctx, memberID, pm)
}

func (grm groupRepositoryMiddleware) Members(ctx context.Context, groupID, groupType string, pm groups.PageMetadata) (groups.MemberPage, error) {
	span := createSpan(ctx, grm.tracer, members)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.Members(ctx, groupID, groupType, pm)
}

func (grm groupRepositoryMiddleware) Assign(ctx context.Context, groupID, groupType string, memberIDs ...string) error {
	span := createSpan(ctx, grm.tracer, assign)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.Assign(ctx, groupID, groupType, memberIDs...)
}

func (grm groupRepositoryMiddleware) Unassign(ctx context.Context, groupID string, memberIDs ...string) error {
	span := createSpan(ctx, grm.tracer, unassign)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.Unassign(ctx, groupID, memberIDs...)
}
