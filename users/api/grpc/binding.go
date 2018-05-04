package grpc

import (
	"context"

	"github.com/mainflux/mainflux/users"
)

type identifierService struct {
	svc users.Service
}

// NewService returns new gRPC users service instance.
func NewService(svc users.Service) UsersServiceServer {
	return identifierService{svc}
}

func (is identifierService) Identify(ctx context.Context, token *Token) (*Identity, error) {
	id, err := is.svc.Identify(token.GetValue())
	if err != nil {
		return nil, err
	}
	return &Identity{id}, nil
}
