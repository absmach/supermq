// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package events

import "github.com/absmach/supermq/pkg/events"

const (
	wsPrefix          = "websocket"
	clientPublish     = wsPrefix + ".client_publish"
	clientSubscribe   = wsPrefix + ".client_subscribe"
	clientUnsubscribe = wsPrefix + ".client_unsubscribe"
)

var _ events.Event = (*wsEvent)(nil)

type wsEvent struct {
	operation string
	channelID string
	clientID  string
	topic     string
	instance  string
}

func (we wsEvent) Encode() (map[string]interface{}, error) {
	return map[string]interface{}{
		"operation":  we.operation,
		"channel_id": we.channelID,
		"client_id":  we.clientID,
		"topic":      we.topic,
		"instance":   we.instance,
	}, nil
}
