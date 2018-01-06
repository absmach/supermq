package http

import (
	manager "github.com/mainflux/mainflux/manager/client"
	"github.com/mainflux/mainflux/writer"
)

var _ Service = (*adapterService)(nil)

type adapterService struct {
	mr writer.MessageRepository
	mc manager.ManagerClient
}

// NewService instantiates the domain service implementation.
func NewService(mr writer.MessageRepository, mc manager.ManagerClient) Service {
	return &adapterService{mr, mc}
}

func (as *adapterService) Publish(msg writer.RawMessage) error {
	return as.mr.Save(msg)
}

func (as *adapterService) Manager() manager.ManagerClient {
	return as.mc
}
