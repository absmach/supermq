// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"

	"github.com/absmach/magistrala/activitylog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var _ activitylog.Service = (*tracing)(nil)

type tracing struct {
	tracer trace.Tracer
	svc    activitylog.Service
}

func Tracing(svc activitylog.Service, tracer trace.Tracer) activitylog.Service {
	return &tracing{tracer, svc}
}

func (tm *tracing) Save(ctx context.Context, activity activitylog.Activity) error {
	ctx, span := tm.tracer.Start(ctx, "save", trace.WithAttributes(
		attribute.String("id", activity.ID),
		attribute.String("operation", activity.Operation),
	))
	defer span.End()

	return tm.svc.Save(ctx, activity)
}

func (tm *tracing) ReadAll(ctx context.Context, token string, page activitylog.Page) (activitylog.ActivitiesPage, error) {
	ctx, span := tm.tracer.Start(ctx, "read_all", trace.WithAttributes(
		attribute.String("id", page.ID),
		attribute.String("entity_type", page.EntityType),
	))
	defer span.End()

	return tm.svc.ReadAll(ctx, token, page)
}
