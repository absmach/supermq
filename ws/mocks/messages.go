package mocks

import (
	"github.com/mainflux/mainflux"
	broker "github.com/nats-io/go-nats"
)

var _ mainflux.MessagePublisher = (*publisher)(nil)

type publisher struct{}

// NewMessagePublisher returns mock message publisher.
func NewMessagePublisher() mainflux.MessagePublisher {
	return publisher{}
}

func (pub publisher) Publish(msg mainflux.RawMessage) error {
	if len(msg.Payload) == 0 {
		return broker.ErrInvalidMsg
	}
	return nil
}
