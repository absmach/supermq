// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// Package http contains the domain concept definitions needed to support
// Mainflux http adapter service functionality.
package http

import (
	"context"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/opentracing/opentracing-go"
)

// Service specifies coap service API.
type Service interface {
	// Publish Messssage
	Publish(ctx context.Context, token string, msg *messaging.Message) error
}

var _ Service = (*adapterService)(nil)

type adapterService struct {
	publisher messaging.Publisher
	things    mainflux.ThingsServiceClient
	tracer    opentracing.Tracer
}

// New instantiates the HTTP adapter implementation.
func New(publisher messaging.Publisher, things mainflux.ThingsServiceClient, tracer opentracing.Tracer) Service {
	return &adapterService{
		publisher: publisher,
		things:    things,
		tracer:    tracer,
	}
}

func (as *adapterService) Publish(ctx context.Context, token string, msg *messaging.Message) error {
	span := as.tracer.StartSpan("http publish")
	defer span.Finish()
	ar := &mainflux.AccessByKeyReq{
		Token:  token,
		ChanID: msg.Channel,
	}
	thid, err := as.things.CanAccessByKey(ctx, ar)
	if err != nil {
		return err
	}
	msg.Publisher = thid.GetValue()

	return as.publisher.Publish(msg.Channel, msg, span.Context())
}
