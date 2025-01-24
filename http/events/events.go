// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package events

import "github.com/absmach/supermq/pkg/events"

const (
	httpPrefix    = "http"
	clientPublish = httpPrefix + ".client_publish"
	clientConnect = httpPrefix + ".client_connect"
)

var _ events.Event = (*httpEvent)(nil)

type httpEvent struct {
	operation string
	channelID string
	clientID  string
	topic     string
	instance  string
}

func (he httpEvent) Encode() (map[string]interface{}, error) {
	val := map[string]interface{}{
		"operation":  he.operation,
		"channel_id": he.channelID,
		"client_id":  he.clientID,
		"topic":      he.topic,
		"instance":   he.instance,
	}
	return val, nil
}
