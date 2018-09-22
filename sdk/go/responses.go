//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package sdk

import (
	"github.com/mainflux/mainflux/things"
)

type tokenRes struct {
	token string `json:"token,omitempty"`
}

type thingRes struct {
	id      uint64 `json:"id,omitempty"`
	created bool   `json:"created,omitempty"`
}

type viewThingRes struct {
	thing things.Thing `json:"thing,omitempty"`
}

type listThingsRes struct {
	things []things.Thing `json:"things,omitempty"`
}

type channelRes struct {
	id      uint64 `json:"id,omitempty"`
	created bool   `json:"created,omitempty"`
}

type viewChannelRes struct {
	channel things.Channel `json:"things,omitempty"`
}

type listChannelsRes struct {
	channels []things.Channel `json:"channels,omitempty"`
}
