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

func (is identifierService) Identity(ctx context.Context, req *IdentityRequest) (*IdentityResponse, error) {
	id, err := is.svc.Identity(req.GetToken())
	if err != nil {
		return nil, err
	}
	return &IdentityResponse{id}, nil
}
