package http

import "github.com/mainflux/mainflux"

var _ Service = (*adapterService)(nil)

type adapterService struct {
	pub mainflux.MessagePublisher
}

// NewService instantiates the domain service implementation.
func NewService(pub mainflux.MessagePublisher) Service {
	return &adapterService{pub}
}

func (as *adapterService) Publish(msg mainflux.RawMessage) error {
	return as.pub.Publish(msg)
}
