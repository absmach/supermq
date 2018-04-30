package grpc

import (
	"context"

	"github.com/mainflux/mainflux/users"
)

type identifierService struct {
	svc users.Service
}

// NewServer returns new gRPC users service instance.
func NewServer(svc users.Service) UsersServiceServer {
	return identifierService{svc}
}

func (is identifierService) Identify(ctx context.Context, req *IdentifyRequest) (*IdentifyResponse, error) {
	id, err := is.svc.Identify(req.GetToken())
	if err != nil {
		return nil, err
	}
	return &IdentifyResponse{id}, nil
}
