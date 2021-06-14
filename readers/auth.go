package readers

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mainflux/mainflux"
	"google.golang.org/grpc"
)

type authService struct {
	ts mainflux.ThingsServiceClient
	us mainflux.AuthServiceClient
}

type Auth interface {
	Identify(ctx context.Context, in *mainflux.Token, opts ...grpc.CallOption) (*mainflux.UserIdentity, error)
	CanAccessByKey(ctx context.Context, in *mainflux.AccessByKeyReq, opts ...grpc.CallOption) (*mainflux.ThingID, error)
	IsChannelOwner(ctx context.Context, in *mainflux.ChannelOwnerReq, opts ...grpc.CallOption) (*empty.Empty, error)
	CanAccessByID(ctx context.Context, in *mainflux.AccessByIDReq, opts ...grpc.CallOption) (*empty.Empty, error)
}

func NewAuthService(ts mainflux.ThingsServiceClient, auth mainflux.AuthServiceClient) Auth {
	return &authService{
		ts: ts,
		us: auth,
	}
}

func (as *authService) Identify(ctx context.Context, in *mainflux.Token, opts ...grpc.CallOption) (*mainflux.UserIdentity, error) {
	return as.us.Identify(ctx, in, opts...)
}

func (as *authService) CanAccessByKey(ctx context.Context, in *mainflux.AccessByKeyReq, opts ...grpc.CallOption) (*mainflux.ThingID, error) {
	return as.ts.CanAccessByKey(ctx, in, opts...)
}

func (as *authService) CanAccessByID(ctx context.Context, in *mainflux.AccessByIDReq, opts ...grpc.CallOption) (*empty.Empty, error) {
	return as.ts.CanAccessByID(ctx, in, opts...)
}

func (as *authService) IsChannelOwner(ctx context.Context, in *mainflux.ChannelOwnerReq, opts ...grpc.CallOption) (*empty.Empty, error) {
	return as.ts.IsChannelOwner(ctx, in, opts...)
}
