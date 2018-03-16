package mocks

import (
	"errors"

	"github.com/mainflux/mainflux"
)

var _ mainflux.MessagePublisher = (*publisher)(nil)

var (
	errFailedMessagePublish = errors.New("failed to publish message")
)

type publisher struct{}

// NewMessagePublisher returns mock message publisher.
func NewMessagePublisher() mainflux.MessagePublisher {
	return publisher{}
}

func (pub publisher) Publish(msg mainflux.RawMessage) error {
	if len(msg.Payload) == 0 {
		return errFailedMessagePublish
	}
	return nil
}
