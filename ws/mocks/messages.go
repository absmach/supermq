// Copyright (c) Mainflux
// SPDX-Licence-Identifier: Apache-2.0

package mocks

// import (
// 	"github.com/mainflux/mainflux/pkg/messaging"
// 	"github.com/mainflux/mainflux/ws"
// )

// var _ messaging.PubSub = (*mockPubSub)(nil)

// type mockPubSub struct {
// 	subscriptions map[string]*ws.Connclient
// }

// // New returns mock message publisher
// func New(sub map[string]*ws.Connclient) messaging.PubSub {
// 	return &mockPubSub{
// 		subscriptions: sub,
// 	}
// }

// func (mb mockPubSub) Publish(_ string, msg messaging.Message) error {
// 	if len(msg.Payload) == 0 {
// 		return ws.ErrFailedMessagePublish
// 	}

// 	return nil
// }

// func (mb mockPubSub) Subscribe(_, _ string, c messaging.MessageHandler) error {
// 	// Try to mock this :- svc.pubsub.Subscribe(thid.GetValue(), subject, c)
// 	return nil
// }

// func (mb mockPubSub) Unsubscribe(_, _ string) error {
// 	// Try to mock this :- svc.pubsub.Subscribe(thid.GetValue(), subject, c)
// 	return nil
// }

// func (mb mockPubSub) Close() error {
// 	return nil
// }
