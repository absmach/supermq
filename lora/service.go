package lora

import (
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/gogo/protobuf/proto"
	"github.com/nats-io/go-nats"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/logger"
)

const protocol = "lora"

// Service specifies an API that must be fullfiled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	// Create Thing route-map
	CreateThing(string, string) error

	// Create Thing route-map
	UpdateThing(string, string) error

	// Remove Thing route-map
	RemoveThing(string) error

	// Create Channel route-map
	CreateChannel(string, string) error

	// Remove Channel route-map
	RemoveChannel(string) error

	// Update Channel route-map
	UpdateChannel(string, string) error

	// Publish messages on Mainflux NATS broker
	MessageRouter(Message) error
}

var _ Service = (*adapterService)(nil)

type adapterService struct {
	natsConn   *nats.Conn
	eventStore EventStore
	logger     logger.Logger
	routeMap   RouteMapRepository
}

// EventStore represents event source for things and channels provisioning.
type EventStore interface {
	// Subscribes to geven subject and receives events.
	Subscribe(string)
}

// New instantiates the HTTP adapter implementation.
func New(nc *nats.Conn, m RouteMapRepository, logger logger.Logger) Service {
	return &adapterService{
		natsConn: nc,
		routeMap: m,
		logger:   logger,
	}
}

// MessageRouter routes messages from Lora MQTT broker to Mainflux NATS broker
func (as *adapterService) MessageRouter(m Message) error {
	// Get route map of lora application
	d, err := as.routeMap.Get(m.DevEUI)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Routing doesn't exist for this LoRa DeviceEUI: %s", m.DevEUI))
	}
	mfxDev, err := strconv.ParseUint(d, 10, 64)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to decode deviceEUI: %s", m.DevEUI))
	}

	// Get route map of lora application
	c, err := as.routeMap.Get(m.ApplicationID)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Routing doesn't exist for this LoRa applicationID: %s", m.ApplicationID))
	}
	mfxChan, err := strconv.ParseUint(c, 10, 64)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to decode LoRa applicationID: %s", err.Error()))
	}

	payload, err := base64.StdEncoding.DecodeString(m.Data)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Failed to decode Lora message: %s", err.Error()))
	}

	// Publish on Mainflux NATS broker
	msg := mainflux.RawMessage{
		Publisher:   mfxDev,
		Protocol:    protocol,
		ContentType: "Content-Type",
		Channel:     mfxChan,
		Payload:     payload,
	}

	data, err := proto.Marshal(&msg)
	if err != nil {
		return err
	}

	subject := fmt.Sprintf("channel.%d", msg.Channel)
	return as.natsConn.Publish(subject, data)
}

func (as *adapterService) CreateThing(mfxDevID string, loraDevEUI string) error {
	return as.routeMap.Save(mfxDevID, loraDevEUI)
}

func (as *adapterService) UpdateThing(mfxDevID string, loraDevEUI string) error {
	return as.routeMap.Save(mfxDevID, loraDevEUI)
}

func (as *adapterService) RemoveThing(mfxDevID string) error {
	return as.routeMap.Remove(mfxDevID)
}

func (as *adapterService) CreateChannel(mfxChanID string, loraAppID string) error {
	return as.routeMap.Save(mfxChanID, loraAppID)
}

func (as *adapterService) UpdateChannel(mfxChanID string, loraAppID string) error {
	return as.routeMap.Save(mfxChanID, loraAppID)
}

func (as *adapterService) RemoveChannel(mfxChanID string) error {
	return as.routeMap.Remove(mfxChanID)
}
