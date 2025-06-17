// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"

	"github.com/absmach/supermq/http"
	"go.opentelemetry.io/otel/trace"
)

var _ http.Service = (*tracingMiddleware)(nil)

const (
	subscribeOP   = "subscribe_op"
	unsubscribeOP = "unsubscribe_op"
)

type tracingMiddleware struct {
	tracer trace.Tracer
	svc    http.Service
}

// Tracing returns a new http service with tracing capabilities.
func Tracing(tracer trace.Tracer, svc http.Service) http.Service {
	return &tracingMiddleware{
		tracer: tracer,
		svc:    svc,
	}
}

// Subscribe traces the "Subscribe" operation of the wrapped service.
func (tm *tracingMiddleware) Subscribe(ctx context.Context, sessionID, clientKey, domainID, chanID, subtopic string, client *http.Client) error {
	ctx, span := tm.tracer.Start(ctx, subscribeOP)
	defer span.End()

	return tm.svc.Subscribe(ctx, sessionID, clientKey, domainID, chanID, subtopic, client)
}

// Unsubscribe traces the "Unsubscribe" operation of the wrapped service.
func (tm *tracingMiddleware) Unsubscribe(ctx context.Context, sessionID, domainID, chanID, subtopic string) error {
	ctx, span := tm.tracer.Start(ctx, unsubscribeOP)
	defer span.End()

	return tm.svc.Unsubscribe(ctx, sessionID, domainID, chanID, subtopic)
}
