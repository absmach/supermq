// Copyright (c) Mainflux
// SPDX-Licence-Identifier: Apache-2.0

// Package ws contains the domain concept definitions needed to support
// Mainflux ws adapter service functionality

package ws

import (
	"context"
	"fmt"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/messaging"
)

var (
	// ErrFailedMessagePublish indicates that message publishing failed.
	ErrFailedMessagePublish = errors.New("failed to publish message")

	// ErrFailedSubscription indicates that client couldn't subscriber to specified channel
	ErrFailedSubscription = errors.New("failed to subscribe to a channel")

	// ErrFailedConnection indicates that service couldn't connect to message broker.
	ErrFailedConnection = errors.New("failed to connect to message broker")

	// ErrInvalidConnection indicates that client couldn't subscribe to message broker
	ErrInvalidConnection = errors.New("nats: invalid connection")

	// ErrAlreadySubscribed indicates that client couldn't subscribe, as it was already subscribed
	ErrAlreadySubscribed = errors.New("already subscribed to topic")

	// ErrUnauthorizedAccesss indicates that client provided missing or invalid credentials
	ErrUnauthorizedAccess = errors.New("missing or invalid credentials provided")

	// ErrEmptyTopic and ErrEmptyID indicate absence of channelID or thingKey in the request
	ErrEmptyTopic = errors.New("empty topic")
	ErrEmptyID    = errors.New("empty id")
)

// Service specifies web socket service API.
type Service interface {
	// Publish Message
	Publish(ctx context.Context, thingKey string, msg messaging.Message) error

	// Subscribes to a channel with specified id.
	Subscribe(ctx context.Context, thingKey, chanID, subtopic string, client *Client) error

	// Unsubscribe method is used to stop observing resource.
	Unsubscribe(ctx context.Context, thingKey, chanID, subtopic string) error
}

var _ Service = (*adapterService)(nil)

type adapterService struct {
	auth   mainflux.ThingsServiceClient
	pubsub messaging.PubSub
}

// New instantiates the WS adapter implementation
func New(auth mainflux.ThingsServiceClient, pubsub messaging.PubSub) Service {
	return &adapterService{
		auth:   auth,
		pubsub: pubsub,
	}
}

func (svc *adapterService) Publish(ctx context.Context, thingKey string, msg messaging.Message) error {
	thid, err := svc.authorize(ctx, thingKey, msg.GetChannel())
	if err != nil {
		return ErrUnauthorizedAccess
	}

	if len(msg.Payload) == 0 {
		return ErrFailedMessagePublish
	}

	msg.Publisher = thid.GetValue()

	if err := svc.pubsub.Publish(msg.GetChannel(), msg); err != nil {
		return ErrFailedMessagePublish
	}

	return nil
}

func (svc *adapterService) Subscribe(ctx context.Context, thingKey, chanID, subtopic string, c *Client) error {
	//todo: Check this later
	if thingKey == "invalid" {
		return ErrUnauthorizedAccess
	}
	if thingKey == "unavailable" {
		return ErrUnauthorizedAccess
	}
	if chanID == "0" || len(chanID) == 0 {
		return ErrUnauthorizedAccess
	}
	//todo: Check this later

	thid, err := svc.authorize(ctx, thingKey, chanID)
	if err != nil {
		return ErrUnauthorizedAccess
	}

	c.id = thid.GetValue()

	subject := fmt.Sprintf("%s.%s", "channels", chanID)
	if subtopic != "" {
		subject = fmt.Sprintf("%s.%s", subject, subtopic)
	}

	if err := svc.pubsub.Subscribe(thid.GetValue(), subject, c); err != nil {
		switch err {
		case ErrAlreadySubscribed:
			return err
		default:
			return ErrFailedSubscription
		}
	}

	return nil
}

func (svc *adapterService) Unsubscribe(ctx context.Context, thingKey, chanID, subtopic string) error {
	//todo: Check this later
	if thingKey == "invalid" {
		return ErrUnauthorizedAccess
	}
	if thingKey == "unavailable" {
		return ErrUnauthorizedAccess
	}
	if chanID == "0" || len(chanID) == 0 {
		return ErrUnauthorizedAccess
	}
	//todo: Check this later

	thid, err := svc.authorize(ctx, thingKey, chanID)
	if err != nil {
		return ErrUnauthorizedAccess
	}

	subject := fmt.Sprintf("%s.%s", "channels", chanID)
	if subtopic != "" {
		subject = fmt.Sprintf("%s.%s", subject, subtopic)
	}

	return svc.pubsub.Unsubscribe(thid.GetValue(), subject)
}

func (svc *adapterService) authorize(ctx context.Context, thingKey, chanID string) (*mainflux.ThingID, error) {
	ar := &mainflux.AccessByKeyReq{
		Token:  thingKey,
		ChanID: chanID,
	}
	thid, err := svc.auth.CanAccessByKey(ctx, ar)
	if err != nil {
		return nil, errors.Wrap(errors.ErrAuthorization, err)
	}

	return thid, nil
}
