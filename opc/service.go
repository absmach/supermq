package opc

import (
	"context"
	"errors"

	"github.com/mainflux/mainflux"
)

const (
	protocol      = "opc"
	thingSuffix   = "thing"
	channelSuffix = "channel"
)

var (
	// ErrMalformedIdentity indicates malformed identity received (e.g.
	// invalid namespace or ID).
	ErrMalformedIdentity = errors.New("malformed identity received")

	// ErrMalformedMessage indicates malformed OPC-UA message.
	ErrMalformedMessage = errors.New("malformed message received")

	// ErrNotFoundDev indicates a non-existent route map for a device EUI.
	ErrNotFoundDev = errors.New("route map not found for this device EUI")

	// ErrNotFoundApp indicates a non-existent route map for an application ID.
	ErrNotFoundApp = errors.New("route map not found for this application ID")
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

	// CreateChannel creates channel mfx:opc & opc:mfx route-map
	CreateChannel(string, string) error

	// UpdateChannel updates mfx:opc & opc:mfx route-map
	UpdateChannel(string, string) error

	// RemoveChannel removes channel mfx:opc & opc:mfx route-map
	RemoveChannel(string) error

	// Publish forwards messages from the OPC-UA MQTT broker to Mainflux NATS broker
	Publish(context.Context, string, Message) error
}

var _ Service = (*adapterService)(nil)

type adapterService struct {
	publisher  mainflux.MessagePublisher
	thingsRM   RouteMapRepository
	channelsRM RouteMapRepository
}

// New instantiates the OPC-UA adapter implementation.
func New(pub mainflux.MessagePublisher, thingsRM, channelsRM RouteMapRepository) Service {
	return &adapterService{
		publisher:  pub,
		thingsRM:   thingsRM,
		channelsRM: channelsRM,
	}
}

// Publish forwards messages from OPC-UA MQTT broker to Mainflux NATS broker
func (as *adapterService) Publish(ctx context.Context, token string, m Message) error {
	// Get route map of opc application
	// thing, err := as.thingsRM.Get(m.ID)
	// if err != nil {
	//	return ErrNotFoundDev
	// }

	// Get route map of opc application
	// channel, err := as.channelsRM.Get(string(m.Namespace))
	// if err != nil {
	//	return ErrNotFoundApp
	// }

	// Publish on Mainflux NATS broker
	msg := mainflux.RawMessage{
		Publisher:   m.ID,
		Protocol:    protocol,
		ContentType: "Content-Type",
		Channel:     m.Namespace,
		Payload:     m.Data,
	}

	return as.publisher.Publish(ctx, token, msg)
}

func (as *adapterService) CreateThing(mfxDevID string, opcID string) error {
	return as.thingsRM.Save(mfxDevID, opcID)
}

func (as *adapterService) UpdateThing(mfxDevID string, opcID string) error {
	return as.thingsRM.Save(mfxDevID, opcID)
}

func (as *adapterService) RemoveThing(mfxDevID string) error {
	return as.thingsRM.Remove(mfxDevID)
}

func (as *adapterService) CreateChannel(mfxChanID string, opcNamespace string) error {
	return as.channelsRM.Save(mfxChanID, opcNamespace)
}

func (as *adapterService) UpdateChannel(mfxChanID string, opcNamespace string) error {
	return as.channelsRM.Save(mfxChanID, opcNamespace)
}

func (as *adapterService) RemoveChannel(mfxChanID string) error {
	return as.channelsRM.Remove(mfxChanID)
}
