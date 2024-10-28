// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package tracing

import (
	"context"
	"time"

	"github.com/absmach/magistrala/pat"
	"github.com/absmach/magistrala/pkg/authn"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var _ pat.Service = (*tracingMiddleware)(nil)

type tracingMiddleware struct {
	tracer trace.Tracer
	svc    pat.Service
}

// New returns a new group service with tracing capabilities.
func New(svc pat.Service, tracer trace.Tracer) pat.Service {
	return &tracingMiddleware{tracer, svc}
}

func (tm *tracingMiddleware) CreatePAT(ctx context.Context, session authn.Session, name, description string, duration time.Duration, scope pat.Scope) (pat.PAT, error) {
	ctx, span := tm.tracer.Start(ctx, "create_pat", trace.WithAttributes(
		attribute.String("name", name),
		attribute.String("description", description),
		attribute.String("duration", duration.String()),
		attribute.String("scope", scope.String()),
	))
	defer span.End()
	return tm.svc.CreatePAT(ctx, session, name, description, duration, scope)
}

func (tm *tracingMiddleware) UpdatePATName(ctx context.Context, session authn.Session, patID, name string) (pat.PAT, error) {
	ctx, span := tm.tracer.Start(ctx, "update_pat_name", trace.WithAttributes(
		attribute.String("pat_id", patID),
		attribute.String("name", name),
	))
	defer span.End()
	return tm.svc.UpdatePATName(ctx, session, patID, name)
}

func (tm *tracingMiddleware) UpdatePATDescription(ctx context.Context, session authn.Session, patID, description string) (pat.PAT, error) {
	ctx, span := tm.tracer.Start(ctx, "update_pat_description", trace.WithAttributes(
		attribute.String("pat_id", patID),
		attribute.String("description", description),
	))
	defer span.End()
	return tm.svc.UpdatePATDescription(ctx, session, patID, description)
}

func (tm *tracingMiddleware) RetrievePAT(ctx context.Context, session authn.Session, patID string) (pat.PAT, error) {
	ctx, span := tm.tracer.Start(ctx, "retrieve_pat", trace.WithAttributes(
		attribute.String("pat_id", patID),
	))
	defer span.End()
	return tm.svc.RetrievePAT(ctx, session, patID)
}

func (tm *tracingMiddleware) ListPATS(ctx context.Context, session authn.Session, pm pat.PATSPageMeta) (pat.PATSPage, error) {
	ctx, span := tm.tracer.Start(ctx, "list_pat", trace.WithAttributes(
		attribute.Int64("limit", int64(pm.Limit)),
		attribute.Int64("offset", int64(pm.Offset)),
	))
	defer span.End()
	return tm.svc.ListPATS(ctx, session, pm)
}

func (tm *tracingMiddleware) DeletePAT(ctx context.Context, session authn.Session, patID string) error {
	ctx, span := tm.tracer.Start(ctx, "delete_pat", trace.WithAttributes(
		attribute.String("pat_id", patID),
	))
	defer span.End()
	return tm.svc.DeletePAT(ctx, session, patID)
}

func (tm *tracingMiddleware) ResetPATSecret(ctx context.Context, session authn.Session, patID string, duration time.Duration) (pat.PAT, error) {
	ctx, span := tm.tracer.Start(ctx, "reset_pat_secret", trace.WithAttributes(
		attribute.String("pat_id", patID),
		attribute.String("duration", duration.String()),
	))
	defer span.End()
	return tm.svc.ResetPATSecret(ctx, session, patID, duration)
}

func (tm *tracingMiddleware) RevokePATSecret(ctx context.Context, session authn.Session, patID string) error {
	ctx, span := tm.tracer.Start(ctx, "revoke_pat_secret", trace.WithAttributes(
		attribute.String("pat_id", patID),
	))
	defer span.End()
	return tm.svc.RevokePATSecret(ctx, session, patID)
}

func (tm *tracingMiddleware) AddPATScopeEntry(ctx context.Context, session authn.Session, patID string, platformEntityType pat.PlatformEntityType, optionalDomainID string, optionalDomainEntityType pat.DomainEntityType, operation pat.OperationType, entityIDs ...string) (pat.Scope, error) {
	ctx, span := tm.tracer.Start(ctx, "add_pat_scope_entry", trace.WithAttributes(
		attribute.String("pat_id", patID),
		attribute.String("platform_entity", platformEntityType.String()),
		attribute.String("optional_domain_id", optionalDomainID),
		attribute.String("optional_domain_entity", optionalDomainEntityType.String()),
		attribute.String("operation", operation.String()),
		attribute.StringSlice("entities", entityIDs),
	))
	defer span.End()
	return tm.svc.AddPATScopeEntry(ctx, session, patID, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
}

func (tm *tracingMiddleware) RemovePATScopeEntry(ctx context.Context, session authn.Session, patID string, platformEntityType pat.PlatformEntityType, optionalDomainID string, optionalDomainEntityType pat.DomainEntityType, operation pat.OperationType, entityIDs ...string) (pat.Scope, error) {
	ctx, span := tm.tracer.Start(ctx, "remove_pat_scope_entry", trace.WithAttributes(
		attribute.String("pat_id", patID),
		attribute.String("platform_entity", platformEntityType.String()),
		attribute.String("optional_domain_id", optionalDomainID),
		attribute.String("optional_domain_entity", optionalDomainEntityType.String()),
		attribute.String("operation", operation.String()),
		attribute.StringSlice("entities", entityIDs),
	))
	defer span.End()
	return tm.svc.RemovePATScopeEntry(ctx, session, patID, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
}

func (tm *tracingMiddleware) ClearPATAllScopeEntry(ctx context.Context, session authn.Session, patID string) error {
	ctx, span := tm.tracer.Start(ctx, "clear_pat_all_scope_entry", trace.WithAttributes(
		attribute.String("pat_id", patID),
	))
	defer span.End()
	return tm.svc.ClearPATAllScopeEntry(ctx, session, patID)
}

func (tm *tracingMiddleware) IdentifyPAT(ctx context.Context, paToken string) (pat.PAT, error) {
	ctx, span := tm.tracer.Start(ctx, "identity_pat")
	defer span.End()
	return tm.svc.IdentifyPAT(ctx, paToken)
}

func (tm *tracingMiddleware) AuthorizePAT(ctx context.Context, paToken string, platformEntityType pat.PlatformEntityType, optionalDomainID string, optionalDomainEntityType pat.DomainEntityType, operation pat.OperationType, entityIDs ...string) error {
	ctx, span := tm.tracer.Start(ctx, "authorize_pat", trace.WithAttributes(
		attribute.String("personal_access_token", paToken),
		attribute.String("platform_entity", platformEntityType.String()),
		attribute.String("optional_domain_id", optionalDomainID),
		attribute.String("optional_domain_entity", optionalDomainEntityType.String()),
		attribute.String("operation", operation.String()),
		attribute.StringSlice("entities", entityIDs),
	))
	defer span.End()
	return tm.svc.AuthorizePAT(ctx, paToken, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
}

func (tm *tracingMiddleware) CheckPAT(ctx context.Context, userID, patID string, platformEntityType pat.PlatformEntityType, optionalDomainID string, optionalDomainEntityType pat.DomainEntityType, operation pat.OperationType, entityIDs ...string) error {
	ctx, span := tm.tracer.Start(ctx, "check_pat", trace.WithAttributes(
		attribute.String("user_id", userID),
		attribute.String("patID", patID),
		attribute.String("platform_entity", platformEntityType.String()),
		attribute.String("optional_domain_id", optionalDomainID),
		attribute.String("optional_domain_entity", optionalDomainEntityType.String()),
		attribute.String("operation", operation.String()),
		attribute.StringSlice("entities", entityIDs),
	))
	defer span.End()
	return tm.svc.CheckPAT(ctx, userID, patID, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
}
