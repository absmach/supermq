package modbus

import (
	"context"

	mflog "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging"
)

const (
	readTopic  = "channels.*.modbus.read.*"
	writeTopic = "channels.*.modbus.write.*"
)

type Service interface {
	// Read subscribes to the Subscriber and
	// reads modbus sensor values while publishing them to publisher.
	Read(ctx context.Context, id string, sub messaging.Subscriber, pub messaging.Publisher) error
	// Write subscribes to the Subscriber and
	// writes to modbus sensor.
	Write(ctx context.Context, id string, sub messaging.Subscriber, pub messaging.Publisher) error
}

type service struct {
	logger mflog.Logger
}

// NewForwarder returns new Forwarder implementation.
func New(logger mflog.Logger) Service {
	return service{
		logger: logger,
	}
}

func (f service) Read(ctx context.Context, id string, sub messaging.Subscriber, pub messaging.Publisher) error {
	return sub.Subscribe(ctx, id, readTopic, handleRead(ctx, pub, f.logger))
}

func handleRead(ctx context.Context, pub messaging.Publisher, logger mflog.Logger) handleFunc {
	return func(msg *messaging.Message) error {
		return nil
	}
}

func (f service) Write(ctx context.Context, id string, sub messaging.Subscriber, pub messaging.Publisher) error {
	return sub.Subscribe(ctx, id, writeTopic, handleWrite(ctx, pub, f.logger))
}

func handleWrite(ctx context.Context, pub messaging.Publisher, logger mflog.Logger) handleFunc {
	return func(msg *messaging.Message) error {
		return nil
	}
}

type handleFunc func(msg *messaging.Message) error

func (h handleFunc) Handle(msg *messaging.Message) error {
	return h(msg)

}

func (h handleFunc) Cancel() error {
	return nil
}
