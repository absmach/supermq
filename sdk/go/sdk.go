//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package sdk

import (
	"fmt"
	"net/http"

	"github.com/mainflux/mainflux/things"
)

var _ SDK = (*MfxSDK)(nil)

type SDK interface {
	// Users
	CreateUser(user, pwd string) error
	CreateToken(user, pwd string) (string, error)

	// Things
	CreateThing(data, token string) (string, error)
	Things(token string) ([]things.Thing, error)
	Thing(id, token string) (things.Thing, error)
	UpdateThing(id, data, token string) error
	DeleteThing(id, token string) error
	ConnectThing(thingID, chanID, token string) error
	DisconnectThing(thingID, chanID, token string) error

	// Channels
	CreateChannel(data, token string) (string, error)
	Channels(token string) ([]things.Channel, error)
	Channel(id, token string) (things.Channel, error)
	UpdateChannel(id, data, token string) error
	DeleteChannel(id, token string) error

	// Messages
	SendMessage(id, msg, token string) error
	SetContentType(ct string) error
}

type sdkConfig struct {
	host       string
	port       string
	url        string
	httpClient *http.Client
	tls        bool
}

type MfxSDK struct {
	config sdkConfig
}

func NewMfxSDK(host, port string, tls bool) *MfxSDK {
	cfg := sdkConfig{
		host: host,
		port: port,
		tls:  tls,
	}

	if tls == true {
		cfg.url = fmt.Sprintf("https://%s:%s", host, port)
		cfg.httpClient = setCerts()
	} else {
		cfg.url = fmt.Sprintf("http://%s:%s", host, port)
		cfg.httpClient = &http.Client{}
	}

	return &MfxSDK{config: cfg}
}

func (sdk *MfxSDK) sendRequest(req *http.Request, token, contentType string) (*http.Response, error) {
	req.Header.Set("Authorization", token)
	req.Header.Add("Content-Type", contentType)

	return sdk.config.httpClient.Do(req)
}
