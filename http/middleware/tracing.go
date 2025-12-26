// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"

	smqhttp "github.com/absmach/supermq/http"
	"github.com/absmach/supermq/pkg/messaging"
	"go.opentelemetry.io/otel/trace"
)

var _ smqhttp.Service = (*tracingMiddleware)(nil)

const (
	subscribeOP   = "subscribe_op"
	unsubscribeOP = "unsubscribe_op"
)

type tracingMiddleware struct {
	tracer trace.Tracer
	svc    smqhttp.Service
}

// NewTracing returns a new websocket service with tracing capabilities.
func NewTracing(tracer trace.Tracer, svc smqhttp.Service) smqhttp.Service {
	return &tracingMiddleware{
		tracer: tracer,
		svc:    svc,
	}
}

// Subscribe traces the "Subscribe" operation of the wrapped smqhttp.Service.
func (tm *tracingMiddleware) Subscribe(ctx context.Context, sessionID, username, password, domainID, chanID, subtopic string, topicType messaging.TopicType, client *smqhttp.Client) error {
	ctx, span := tm.tracer.Start(ctx, subscribeOP)
	defer span.End()

	return tm.svc.Subscribe(ctx, sessionID, username, password, domainID, chanID, subtopic, topicType, client)
}

func (tm *tracingMiddleware) Unsubscribe(ctx context.Context, sessionID, domainID, chanID, subtopic string, topicType messaging.TopicType) error {
	ctx, span := tm.tracer.Start(ctx, unsubscribeOP)
	defer span.End()

	return tm.svc.Unsubscribe(ctx, sessionID, domainID, chanID, subtopic, topicType)
}
