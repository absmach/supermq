// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// Package tracing contains middlewares that will add spans
// to existing traces.
package tracing

import (
	"context"

	notifiers "github.com/mainflux/mainflux/consumers/notifiers"
	"go.opentelemetry.io/otel/trace"
)

const (
	saveOp        = "save_op"
	retrieveOp    = "retrieve_op"
	retrieveAllOp = "retrieve_all_op"
	removeOp      = "remove_op"
)

var _ notifiers.SubscriptionsRepository = (*subRepositoryMiddleware)(nil)

type subRepositoryMiddleware struct {
	tracer trace.Tracer
	repo   notifiers.SubscriptionsRepository
}

// New instantiates a new Subscriptions repository that
// tracks request and their latency, and adds spans to context.
func New(repo notifiers.SubscriptionsRepository, tracer trace.Tracer) notifiers.SubscriptionsRepository {
	return subRepositoryMiddleware{
		tracer: tracer,
		repo:   repo,
	}
}

func (urm subRepositoryMiddleware) Save(ctx context.Context, sub notifiers.Subscription) (string, error) {
	ctx, span := createSpan(ctx, urm.tracer, saveOp)
	defer span.End()

	return urm.repo.Save(ctx, sub)
}

func (urm subRepositoryMiddleware) Retrieve(ctx context.Context, id string) (notifiers.Subscription, error) {
	ctx, span := createSpan(ctx, urm.tracer, retrieveOp)
	defer span.End()

	return urm.repo.Retrieve(ctx, id)
}

func (urm subRepositoryMiddleware) RetrieveAll(ctx context.Context, pm notifiers.PageMetadata) (notifiers.Page, error) {
	ctx, span := createSpan(ctx, urm.tracer, retrieveAllOp)
	defer span.End()

	return urm.repo.RetrieveAll(ctx, pm)
}

func (urm subRepositoryMiddleware) Remove(ctx context.Context, id string) error {
	ctx, span := createSpan(ctx, urm.tracer, removeOp)
	defer span.End()

	return urm.repo.Remove(ctx, id)
}

func createSpan(ctx context.Context, tracer trace.Tracer, opName string) (context.Context, trace.Span) {

	return tracer.Start(ctx, opName)
}
