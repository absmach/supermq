// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.12
// source: clients/policies/clientauth.proto

package policies

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ThingsServiceClient is the client API for ThingsService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ThingsServiceClient interface {
	CanAccessByKey(ctx context.Context, in *AccessByKeyReq, opts ...grpc.CallOption) (*ThingID, error)
	IsChannelOwner(ctx context.Context, in *ChannelOwnerReq, opts ...grpc.CallOption) (*emptypb.Empty, error)
	CanAccessByID(ctx context.Context, in *AccessByIDReq, opts ...grpc.CallOption) (*emptypb.Empty, error)
	Identify(ctx context.Context, in *Token, opts ...grpc.CallOption) (*ThingID, error)
}

type thingsServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewThingsServiceClient(cc grpc.ClientConnInterface) ThingsServiceClient {
	return &thingsServiceClient{cc}
}

func (c *thingsServiceClient) CanAccessByKey(ctx context.Context, in *AccessByKeyReq, opts ...grpc.CallOption) (*ThingID, error) {
	out := new(ThingID)
	err := c.cc.Invoke(ctx, "/policies.ThingsService/CanAccessByKey", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *thingsServiceClient) IsChannelOwner(ctx context.Context, in *ChannelOwnerReq, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/policies.ThingsService/IsChannelOwner", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *thingsServiceClient) CanAccessByID(ctx context.Context, in *AccessByIDReq, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/policies.ThingsService/CanAccessByID", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *thingsServiceClient) Identify(ctx context.Context, in *Token, opts ...grpc.CallOption) (*ThingID, error) {
	out := new(ThingID)
	err := c.cc.Invoke(ctx, "/policies.ThingsService/Identify", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ThingsServiceServer is the server API for ThingsService service.
// All implementations must embed UnimplementedThingsServiceServer
// for forward compatibility
type ThingsServiceServer interface {
	CanAccessByKey(context.Context, *AccessByKeyReq) (*ThingID, error)
	IsChannelOwner(context.Context, *ChannelOwnerReq) (*emptypb.Empty, error)
	CanAccessByID(context.Context, *AccessByIDReq) (*emptypb.Empty, error)
	Identify(context.Context, *Token) (*ThingID, error)
	mustEmbedUnimplementedThingsServiceServer()
}

// UnimplementedThingsServiceServer must be embedded to have forward compatible implementations.
type UnimplementedThingsServiceServer struct {
}

func (UnimplementedThingsServiceServer) CanAccessByKey(context.Context, *AccessByKeyReq) (*ThingID, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CanAccessByKey not implemented")
}
func (UnimplementedThingsServiceServer) IsChannelOwner(context.Context, *ChannelOwnerReq) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IsChannelOwner not implemented")
}
func (UnimplementedThingsServiceServer) CanAccessByID(context.Context, *AccessByIDReq) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CanAccessByID not implemented")
}
func (UnimplementedThingsServiceServer) Identify(context.Context, *Token) (*ThingID, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Identify not implemented")
}
func (UnimplementedThingsServiceServer) mustEmbedUnimplementedThingsServiceServer() {}

// UnsafeThingsServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ThingsServiceServer will
// result in compilation errors.
type UnsafeThingsServiceServer interface {
	mustEmbedUnimplementedThingsServiceServer()
}

func RegisterThingsServiceServer(s grpc.ServiceRegistrar, srv ThingsServiceServer) {
	s.RegisterService(&ThingsService_ServiceDesc, srv)
}

func _ThingsService_CanAccessByKey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AccessByKeyReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ThingsServiceServer).CanAccessByKey(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/policies.ThingsService/CanAccessByKey",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ThingsServiceServer).CanAccessByKey(ctx, req.(*AccessByKeyReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ThingsService_IsChannelOwner_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ChannelOwnerReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ThingsServiceServer).IsChannelOwner(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/policies.ThingsService/IsChannelOwner",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ThingsServiceServer).IsChannelOwner(ctx, req.(*ChannelOwnerReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ThingsService_CanAccessByID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AccessByIDReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ThingsServiceServer).CanAccessByID(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/policies.ThingsService/CanAccessByID",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ThingsServiceServer).CanAccessByID(ctx, req.(*AccessByIDReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ThingsService_Identify_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Token)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ThingsServiceServer).Identify(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/policies.ThingsService/Identify",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ThingsServiceServer).Identify(ctx, req.(*Token))
	}
	return interceptor(ctx, in, info, handler)
}

// ThingsService_ServiceDesc is the grpc.ServiceDesc for ThingsService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ThingsService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "policies.ThingsService",
	HandlerType: (*ThingsServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CanAccessByKey",
			Handler:    _ThingsService_CanAccessByKey_Handler,
		},
		{
			MethodName: "IsChannelOwner",
			Handler:    _ThingsService_IsChannelOwner_Handler,
		},
		{
			MethodName: "CanAccessByID",
			Handler:    _ThingsService_CanAccessByID_Handler,
		},
		{
			MethodName: "Identify",
			Handler:    _ThingsService_Identify_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "clients/policies/clientauth.proto",
}
