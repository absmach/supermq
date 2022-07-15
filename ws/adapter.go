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
)

// Service specifies web socket service API.
type Service interface {
	// Publish Message
	Publish(ctx context.Context, token string, msg messaging.Message) error

	// Authorize message is used to authenticate the thingKey and chanID
	Authorize(ctx context.Context, thingKey, chanID string) (*mainflux.ThingID, error)

	// Subscribes to a channel with specified id.
	Subscribe(ctx context.Context, thingKey, chanID, subtopic string, channel Client) error

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

func (svc *adapterService) Authorize(ctx context.Context, thingKey, chanID string) (*mainflux.ThingID, error) {
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

func (svc *adapterService) Publish(ctx context.Context, thingKey string, msg messaging.Message) error {
	// ar := &mainflux.AccessByKeyReq{
	// 	Token:  thingKey,
	// 	ChanID: msg.GetChannel(),
	// }
	// thid, err := svc.auth.CanAccessByKey(ctx, ar)
	// if err != nil {
	// 	return errors.Wrap(errors.ErrAuthorization, err)
	// }

	thid, err := svc.Authorize(ctx, thingKey, msg.GetChannel())
	if err != nil {
		return err
	}

	msg.Publisher = thid.GetValue()

	return svc.pubsub.Publish(msg.GetChannel(), msg)
}

func (svc *adapterService) Subscribe(ctx context.Context, thingKey, chanID, subtopic string, c Client) error {
	// ar := &mainflux.AccessByKeyReq{
	// 	Token:  thingKey,
	// 	ChanID: chanID,
	// }
	// thid, err := svc.auth.CanAccessByKey(ctx, ar)
	// if err != nil {
	// 	return errors.Wrap(errors.ErrAuthorization, err)
	// }

	thid, err := svc.Authorize(ctx, thingKey, chanID)
	if err != nil {
		return err
	}

	subject := fmt.Sprintf("%s.%s", "channels", chanID)
	if subtopic != "" {
		subject = fmt.Sprintf("%s.%s", subject, subtopic)
	}

	return svc.pubsub.Subscribe(thid.GetValue(), subject, c)
}

func (svc *adapterService) Unsubscribe(ctx context.Context, thingKey, chanID, subtopic string) error {
	// ar := &mainflux.AccessByKeyReq{
	// 	Token:  thingKey,
	// 	ChanID: chanID,
	// }
	// if _, err := svc.auth.CanAccessByKey(ctx, ar); err != nil {
	// 	return errors.Wrap(errors.ErrAuthorization, err)
	// }

	_, err := svc.Authorize(ctx, thingKey, chanID)
	if err != nil {
		return err
	}

	subject := fmt.Sprintf("%s.%s", "channels", chanID)
	if subtopic != "" {
		subject = fmt.Sprintf("%s.%s", subject, subtopic)
	}

	return svc.pubsub.Unsubscribe(thingKey, subject)
}
