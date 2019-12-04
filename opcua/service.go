// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package opcua

import (
	"context"
	"errors"
	"fmt"

	"github.com/mainflux/mainflux"
)

const protocol = "opcua"

var (
	errNotFoundServerURI = errors.New("route map not found for this Server URI")
	errNotFoundNodeID    = errors.New("route map not found for this Node ID")
	errNotFoundConn      = errors.New("connection not found")
)

// Service specifies an API that must be fullfiled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	// CreateThing creates thing  mfx:opc & opc:mfx route-map
	CreateThing(string, string) error

	// UpdateThing updates thing mfx:opc & opc:mfx route-map
	UpdateThing(string, string) error

	// RemoveThing removes thing mfx:opc & opc:mfx route-map
	RemoveThing(string) error

	// CreateChannel creates channel route-map
	CreateChannel(string, string) error

	// UpdateChannel updates chroute-map
	UpdateChannel(string, string) error

	// RemoveChannel removes channel route-map
	RemoveChannel(string) error

	// ConnectThing creates thing and channel connection route-map
	ConnectThing(string, string) error

	// DisconnectThing removes thing and channel connection route-map
	DisconnectThing(string, string) error

	// Publish forwards messages from the OPC-UA MQTT broker to Mainflux NATS broker
	Publish(context.Context, string, Message) error
}

// Config OPC-UA Server
type Config struct {
	ServerURI string
	NodeID    string
	Policy    string
	Mode      string
	CertFile  string
	KeyFile   string
}

var _ Service = (*adapterService)(nil)

type adapterService struct {
	publisher  mainflux.MessagePublisher
	thingsRM   RouteMapRepository
	channelsRM RouteMapRepository
	connectRM  RouteMapRepository
}

// New instantiates the OPC-UA adapter implementation.
func New(pub mainflux.MessagePublisher, thingsRM, channelsRM, connectRM RouteMapRepository) Service {
	return &adapterService{
		publisher:  pub,
		thingsRM:   thingsRM,
		channelsRM: channelsRM,
		connectRM:  connectRM,
	}
}

// Publish forwards messages from OPC-UA MQTT broker to Mainflux NATS broker
func (as *adapterService) Publish(ctx context.Context, token string, m Message) error {
	// Get route-map of the OPC-UA ServerURI
	channelID, err := as.channelsRM.Get(m.ServerURI)
	if err != nil {
		return errNotFoundServerURI
	}

	// Get route-map of the OPC-UA NodeID
	thingID, err := as.thingsRM.Get(m.NodeID)
	if err != nil {
		return errNotFoundNodeID
	}

	// Check connection between ServerURI and NodeID
	c := fmt.Sprintf("%s:%s", channelID, thingID)
	if _, err := as.connectRM.Get(c); err != nil {
		return errNotFoundConn
	}

	// Publish on Mainflux NATS broker
	SenML := fmt.Sprintf(`[{"n":"%s","v":%v}]`, m.Type, m.Data)
	payload := []byte(SenML)
	msg := mainflux.Message{
		Publisher:   thingID,
		Protocol:    protocol,
		ContentType: "Content-Type",
		Channel:     channelID,
		Payload:     payload,
	}

	return as.publisher.Publish(ctx, token, msg)
}

func (as *adapterService) CreateThing(mfxDevID, opcuaNodeID string) error {
	return as.thingsRM.Save(mfxDevID, opcuaNodeID)
}

func (as *adapterService) UpdateThing(mfxDevID, opcuaNodeID string) error {
	return as.thingsRM.Save(mfxDevID, opcuaNodeID)
}

func (as *adapterService) RemoveThing(mfxDevID string) error {
	return as.thingsRM.Remove(mfxDevID)
}

func (as *adapterService) CreateChannel(mfxChanID, opcuaServerURI string) error {
	return as.channelsRM.Save(mfxChanID, opcuaServerURI)
}

func (as *adapterService) UpdateChannel(mfxChanID, opcuaServerURI string) error {
	return as.channelsRM.Save(mfxChanID, opcuaServerURI)
}

func (as *adapterService) RemoveChannel(mfxChanID string) error {
	return as.channelsRM.Remove(mfxChanID)
}

func (as *adapterService) ConnectThing(mfxChanID, mfxThingID string) error {
	c := fmt.Sprintf("%s:%s", mfxChanID, mfxThingID)
	return as.connectRM.Save(c, c)
}

func (as *adapterService) DisconnectThing(mfxChanID, mfxThingID string) error {
	c := fmt.Sprintf("%s:%s", mfxChanID, mfxThingID)
	return as.connectRM.Remove(c)
}
