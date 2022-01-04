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

var _ groups.GroupRepository = (*groupsRepository)(nil)

type groupsRepository struct {
	tracer opentracing.Tracer
	repo   groups.GroupRepository
}

// New returns Tracing repository that tracks request and their latency, and adds spans to context.
func New(tracer opentracing.Tracer, gr groups.GroupRepository) groups.GroupRepository {
	return groupsRepository{
		tracer: tracer,
		repo:   gr,
	}
}

func (grm groupsRepository) Save(ctx context.Context, g groups.Group) (groups.Group, error) {
	span := createSpan(ctx, grm.tracer, saveGroup)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.Save(ctx, g)
}

func (grm groupsRepository) Update(ctx context.Context, g groups.Group) (groups.Group, error) {
	span := createSpan(ctx, grm.tracer, updateGroup)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.Update(ctx, g)
}

func (grm groupsRepository) Delete(ctx context.Context, groupID string) error {
	span := createSpan(ctx, grm.tracer, deleteGroup)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.Delete(ctx, groupID)
}

func (grm groupsRepository) RetrieveByID(ctx context.Context, id string) (groups.Group, error) {
	span := createSpan(ctx, grm.tracer, retrieveByID)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.RetrieveByID(ctx, id)
}

func (grm groupsRepository) RetrieveAllParents(ctx context.Context, groupID string, pm groups.PageMetadata) (groups.GroupPage, error) {
	span := createSpan(ctx, grm.tracer, retrieveAllParents)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.RetrieveAllParents(ctx, groupID, pm)
}

func (grm groupsRepository) RetrieveAllChildren(ctx context.Context, groupID string, pm groups.PageMetadata) (groups.GroupPage, error) {
	span := createSpan(ctx, grm.tracer, retrieveAllChildren)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.RetrieveAllChildren(ctx, groupID, pm)
}

func (grm groupsRepository) RetrieveAll(ctx context.Context, pm groups.PageMetadata) (groups.GroupPage, error) {
	span := createSpan(ctx, grm.tracer, retrieveAll)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.RetrieveAll(ctx, pm)
}

func (grm groupsRepository) Memberships(ctx context.Context, memberID string, pm groups.PageMetadata) (groups.GroupPage, error) {
	span := createSpan(ctx, grm.tracer, memberships)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.Memberships(ctx, memberID, pm)
}

func (grm groupsRepository) Assign(ctx context.Context, groupID string, memberIDs ...string) error {
	span := createSpan(ctx, grm.tracer, assign)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.Assign(ctx, groupID, memberIDs...)
}

func (grm groupsRepository) Unassign(ctx context.Context, groupID string, memberIDs ...string) error {
	span := createSpan(ctx, grm.tracer, unassign)
	defer span.Finish()
	ctx = opentracing.ContextWithSpan(ctx, span)

	return grm.repo.Unassign(ctx, groupID, memberIDs...)
}

func createSpan(ctx context.Context, tracer opentracing.Tracer, opName string) opentracing.Span {
	if parentSpan := opentracing.SpanFromContext(ctx); parentSpan != nil {
		return tracer.StartSpan(
			opName,
			opentracing.ChildOf(parentSpan.Context()),
		)
	}

	return tracer.StartSpan(opName)
}
