package mocks

// var _ mainflux.ThingsServiceClient = (*thingsClient)(nil)

// // ServiceErrToken is used to simulate interal server error
// const ServiceErrToken = "unavailable"

// type thingsClient struct {
// 	things map[string]string
// }

// // NewThingsClient returns mock implementation of things service client.
// func NewThingsClient(data map[string]string) mainflux.ThingsServiceClient {
// 	return &thingsClient{data}
// }

// func (tc thingsClient) CanAccessByKey(ctx context.Context, req *mainflux.AccessByKeyReq, opts ...grpc.CallOption) (*mainflux.ThingID, error) {
// 	key := req.GetToken()

// 	// Since there is no appropriate way to simulate internal server error,
// 	// we had to use this obscure approach. ErrorToken simulates gRPC call
// 	// which returns internal server error.
// 	if key == ServiceErrToken {
// 		return nil, status.Error(codes.Internal, "internal server error")
// 	}
// 	if key == "" {
// 		return nil, errors.New("unauthorized access")
// 	}

// 	thid, ok := tc.things[key]
// 	if !ok {
// 		return nil, status.Error(codes.PermissionDenied, "invalid credentials provided")
// 	}

// 	return &mainflux.ThingID{Value: thid}, nil
// }

// func (tc thingsClient) CanAccessByID(context.Context, *mainflux.AccessByIDReq, ...grpc.CallOption) (*emptypb.Empty, error) {
// 	panic("not implemented")
// }

// func (tc thingsClient) Identify(context.Context, *mainflux.Token, ...grpc.CallOption) (*mainflux.ThingID, error) {
// 	panic("not implemented")
// }

// func (tc thingsClient) IsChannelOwner(ctx context.Context, in *mainflux.ChannelOwnerReq, opts ...grpc.CallOption) (*emptypb.Empty, error) {
// 	panic("not implemented")
// }
