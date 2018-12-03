package lora

import (
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/logger"
)

const (
	protocol     = "lora"
	thingSufix   = "thing"
	channelSufix = "channel"
)

// Service specifies an API that must be fullfiled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	// CreateThing creates thing  mfx:lora & lora:mfx route-map
	CreateThing(string, string) error

	// UpdateThing updates thing mfx:lora & lora:mfx route-map
	UpdateThing(string, string) error

	// RemoveThing removes thing mfx:lora & lora:mfx route-map
	RemoveThing(string) error

	// CreateChannel creates channel mfx:lora & lora:mfx route-map
	CreateChannel(string, string) error

	// UpdateChannel updates mfx:lora & lora:mfx route-map
	UpdateChannel(string, string) error

	// RemoveChannel removes channel mfx:lora & lora:mfx route-map
	RemoveChannel(string) error

	// MessageRouter forward Lora messages to Mainflux NATS broker
	MessageRouter(Message) error
}

var _ Service = (*adapterService)(nil)

type adapterService struct {
	publisher mainflux.MessagePublisher
	routeMap  RouteMapRepository
	logger    logger.Logger
}

// New instantiates the LoRa adapter implementation.
func New(pub mainflux.MessagePublisher, m RouteMapRepository, logger logger.Logger) Service {
	return &adapterService{
		publisher: pub,
		routeMap:  m,
		logger:    logger,
	}
}

// MessageRouter routes messages from Lora MQTT broker to Mainflux NATS broker
func (as *adapterService) MessageRouter(m Message) error {
	// Get route map of lora application
	d, err := as.routeMap.Get(m.DevEUI, thingSufix)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Route map not foud for device EUI %s", m.DevEUI))
	}
	mfxDev, err := strconv.ParseUint(d, 10, 64)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to decode %s as device EUI", m.DevEUI))
	}

	// Get route map of lora application
	c, err := as.routeMap.Get(m.ApplicationID, channelSufix)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Route map not found for application ID %s", m.ApplicationID))
	}
	mfxChan, err := strconv.ParseUint(c, 10, 64)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to decode %s as application ID", m.ApplicationID))
	}

	payload, err := base64.StdEncoding.DecodeString(m.Data)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to decode message %s", err.Error()))
	}

	// Publish on Mainflux NATS broker
	msg := mainflux.RawMessage{
		Publisher:   mfxDev,
		Protocol:    protocol,
		ContentType: "Content-Type",
		Channel:     mfxChan,
		Payload:     payload,
	}

	return as.publisher.Publish(msg)
}

func (as *adapterService) CreateThing(mfxDevID string, loraDevEUI string) error {
	return as.routeMap.Save(mfxDevID, loraDevEUI, thingSufix)
}

func (as *adapterService) UpdateThing(mfxDevID string, loraDevEUI string) error {
	return as.routeMap.Save(mfxDevID, loraDevEUI, thingSufix)
}

func (as *adapterService) RemoveThing(mfxDevID string) error {
	return as.routeMap.Remove(mfxDevID, thingSufix)
}

func (as *adapterService) CreateChannel(mfxChanID string, loraAppID string) error {
	return as.routeMap.Save(mfxChanID, loraAppID, channelSufix)
}

func (as *adapterService) UpdateChannel(mfxChanID string, loraAppID string) error {
	return as.routeMap.Save(mfxChanID, loraAppID, channelSufix)
}

func (as *adapterService) RemoveChannel(mfxChanID string) error {
	return as.routeMap.Remove(mfxChanID, channelSufix)
}
