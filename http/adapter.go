// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// Package http contains the domain concept definitions needed to support
// Mainflux http adapter service functionality.
package http

import (
	"context"

	"github.com/mainflux/mainflux"
	broker "github.com/mainflux/mainflux/broker/nats"
)

// Service specifies coap service API.
type Service interface {
	// Publish Messssage
	Publish(context.Context, string, mainflux.Message) error
}

var _ Service = (*adapterService)(nil)

type adapterService struct {
	pub    broker.Publisher
	things mainflux.ThingsServiceClient
}

// New instantiates the HTTP adapter implementation.
func New(pub broker.Publisher, things mainflux.ThingsServiceClient) Service {
	return &adapterService{
		pub:    pub,
		things: things,
	}
}

func (as *adapterService) Publish(ctx context.Context, token string, msg mainflux.Message) error {
	ar := &mainflux.AccessByKeyReq{
		Token:  token,
		ChanID: msg.GetChannel(),
	}
	thid, err := as.things.CanAccessByKey(ctx, ar)
	if err != nil {
		return err
	}
	msg.Publisher = thid.GetValue()

	return as.pub.Publish(ctx, token, msg)
}
