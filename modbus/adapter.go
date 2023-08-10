package modbus

import (
	"context"
	"net/url"
	"strings"

	mflog "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging"
)

const (
	readTopic  = "modbus-read"
	writeTopic = "modbus-write"
)

type Service interface {
	// Forward subscribes to the Subscriber and
	// publishes messages using provided Publisher.
	Read(ctx context.Context, id string, sub messaging.Subscriber, pub messaging.Publisher) error
	// Forward subscribes to the Subscriber and
	// publishes messages using provided Publisher.
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

func getClient(address string) (ModbusService, error) {
	switch {
	case strings.HasPrefix(address, "/dev/") || strings.Contains(address, "COM"):
		return NewRTUClient(address, RTUHandlerOptions{})
	default:
		if _, err := url.Parse(address); err != nil {
			return nil, err
		}
		return NewTCPClient(address, TCPHandlerOptions{})
	}
}
