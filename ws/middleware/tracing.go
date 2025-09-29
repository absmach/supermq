// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"

	"github.com/absmach/supermq/pkg/messaging"
	"github.com/absmach/supermq/ws"
	"go.opentelemetry.io/otel/trace"
)

var _ ws.Service = (*tracingMiddleware)(nil)

const (
	subscribeOP   = "subscribe_op"
	unsubscribeOP = "unsubscribe_op"
)

type tracingMiddleware struct {
	tracer trace.Tracer
	svc    ws.Service
}

// NewTracing returns a new websocket service with tracing capabilities.
func NewTracing(tracer trace.Tracer, svc ws.Service) ws.Service {
	return &tracingMiddleware{
		tracer: tracer,
		svc:    svc,
	}
}

// Subscribe traces the "Subscribe" operation of the wrapped ws.Service.
func (tm *tracingMiddleware) Subscribe(ctx context.Context, sessionID, username, password, domainID, chanID, subtopic string, topicType messaging.TopicType, client *ws.Client) error {
	ctx, span := tm.tracer.Start(ctx, subscribeOP)
	defer span.End()

	return tm.svc.Subscribe(ctx, sessionID, username, password, domainID, chanID, subtopic, topicType, client)
}

func (tm *tracingMiddleware) Unsubscribe(ctx context.Context, sessionID, domainID, chanID, subtopic string, topicType messaging.TopicType) error {
	ctx, span := tm.tracer.Start(ctx, unsubscribeOP)
	defer span.End()

	return tm.svc.Unsubscribe(ctx, sessionID, domainID, chanID, subtopic, topicType)
}
