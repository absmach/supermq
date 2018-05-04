package mocks

import (
	"context"

	"github.com/mainflux/mainflux/users"
	pb "github.com/mainflux/mainflux/users/api/grpc"
	"google.golang.org/grpc"
)

var _ pb.UsersServiceClient = (*usersServiceMock)(nil)

type usersServiceMock struct {
	users map[string]string
}

// NewUsersService creates mock of users service.
func NewUsersService(users map[string]string) pb.UsersServiceClient {
	return &usersServiceMock{users}
}

func (svc usersServiceMock) Identify(ctx context.Context, in *pb.Token, opts ...grpc.CallOption) (*pb.Identity, error) {
	if id, ok := svc.users[in.Value]; ok {
		return &pb.Identity{id}, nil
	}
	return nil, users.ErrUnauthorizedAccess
}
