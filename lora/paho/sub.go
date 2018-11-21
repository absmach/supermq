package paho

// LoraSubscribe subscribe to lora server messages
import (
	"encoding/json"
	"fmt"

	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/lora"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	nats "github.com/nats-io/go-nats"
)

const loraServerTopic = "application/+/node/+/rx"

type pubsub struct {
	nc     *nats.Conn
	svc    lora.Service
	logger logger.Logger
}

// Subscribe subscribes to the Lora MQTT message broker
func Subscribe(svc lora.Service, mc mqtt.Client, nc *nats.Conn, log logger.Logger) error {
	ps := pubsub{
		svc:    svc,
		nc:     nc,
		logger: log,
	}

	s := mc.Subscribe(loraServerTopic, 0, ps.handleMsg)
	if err := s.Error(); s.Wait() && err != nil {
		ps.logger.Error(fmt.Sprintf("Failed to subscribe to lora message broker: %s", err.Error()))
		return err
	}

	return nil
}

// handleMsg triggered when new message is received on Lora MQTT broker
func (ps *pubsub) handleMsg(c mqtt.Client, msg mqtt.Message) {
	m := lora.Message{}
	err := json.Unmarshal(msg.Payload(), &m)
	if err != nil {
		ps.logger.Error(fmt.Sprintf("Failed to Unmarshal message: %s", err.Error()))
		return
	}

	// TODO: Decode data to publish on proper channel
	ps.svc.MessageRouter(m, ps.nc)
	return
}
