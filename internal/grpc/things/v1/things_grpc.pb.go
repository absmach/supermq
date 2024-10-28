// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.28.2
// source: things/v1/things.proto

package v1

import (
	context "context"
	v1 "github.com/absmach/magistrala/internal/grpc/common/v1"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	ThingsService_Authorize_FullMethodName                  = "/things.v1.ThingsService/Authorize"
	ThingsService_RetrieveEntity_FullMethodName             = "/things.v1.ThingsService/RetrieveEntity"
	ThingsService_RetrieveEntities_FullMethodName           = "/things.v1.ThingsService/RetrieveEntities"
	ThingsService_AddConnections_FullMethodName             = "/things.v1.ThingsService/AddConnections"
	ThingsService_RemoveConnections_FullMethodName          = "/things.v1.ThingsService/RemoveConnections"
	ThingsService_RemoveChannelConnections_FullMethodName   = "/things.v1.ThingsService/RemoveChannelConnections"
	ThingsService_UnsetParentGroupFromThings_FullMethodName = "/things.v1.ThingsService/UnsetParentGroupFromThings"
)

// ThingsServiceClient is the client API for ThingsService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// ThingsService is a service that provides things authorization functionalities
// for magistrala services.
type ThingsServiceClient interface {
	// Authorize checks if the thing is authorized to perform
	Authorize(ctx context.Context, in *AuthzReq, opts ...grpc.CallOption) (*AuthzRes, error)
	RetrieveEntity(ctx context.Context, in *v1.RetrieveEntityReq, opts ...grpc.CallOption) (*v1.RetrieveEntityRes, error)
	RetrieveEntities(ctx context.Context, in *v1.RetrieveEntitiesReq, opts ...grpc.CallOption) (*v1.RetrieveEntitiesRes, error)
	AddConnections(ctx context.Context, in *v1.AddConnectionsReq, opts ...grpc.CallOption) (*v1.AddConnectionsRes, error)
	RemoveConnections(ctx context.Context, in *v1.RemoveConnectionsReq, opts ...grpc.CallOption) (*v1.RemoveConnectionsRes, error)
	RemoveChannelConnections(ctx context.Context, in *RemoveChannelConnectionsReq, opts ...grpc.CallOption) (*RemoveChannelConnectionsRes, error)
	UnsetParentGroupFromThings(ctx context.Context, in *UnsetParentGroupFromThingsReq, opts ...grpc.CallOption) (*UnsetParentGroupFromThingsRes, error)
}

type thingsServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewThingsServiceClient(cc grpc.ClientConnInterface) ThingsServiceClient {
	return &thingsServiceClient{cc}
}

func (c *thingsServiceClient) Authorize(ctx context.Context, in *AuthzReq, opts ...grpc.CallOption) (*AuthzRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AuthzRes)
	err := c.cc.Invoke(ctx, ThingsService_Authorize_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *thingsServiceClient) RetrieveEntity(ctx context.Context, in *v1.RetrieveEntityReq, opts ...grpc.CallOption) (*v1.RetrieveEntityRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(v1.RetrieveEntityRes)
	err := c.cc.Invoke(ctx, ThingsService_RetrieveEntity_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *thingsServiceClient) RetrieveEntities(ctx context.Context, in *v1.RetrieveEntitiesReq, opts ...grpc.CallOption) (*v1.RetrieveEntitiesRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(v1.RetrieveEntitiesRes)
	err := c.cc.Invoke(ctx, ThingsService_RetrieveEntities_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *thingsServiceClient) AddConnections(ctx context.Context, in *v1.AddConnectionsReq, opts ...grpc.CallOption) (*v1.AddConnectionsRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(v1.AddConnectionsRes)
	err := c.cc.Invoke(ctx, ThingsService_AddConnections_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *thingsServiceClient) RemoveConnections(ctx context.Context, in *v1.RemoveConnectionsReq, opts ...grpc.CallOption) (*v1.RemoveConnectionsRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(v1.RemoveConnectionsRes)
	err := c.cc.Invoke(ctx, ThingsService_RemoveConnections_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *thingsServiceClient) RemoveChannelConnections(ctx context.Context, in *RemoveChannelConnectionsReq, opts ...grpc.CallOption) (*RemoveChannelConnectionsRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RemoveChannelConnectionsRes)
	err := c.cc.Invoke(ctx, ThingsService_RemoveChannelConnections_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *thingsServiceClient) UnsetParentGroupFromThings(ctx context.Context, in *UnsetParentGroupFromThingsReq, opts ...grpc.CallOption) (*UnsetParentGroupFromThingsRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UnsetParentGroupFromThingsRes)
	err := c.cc.Invoke(ctx, ThingsService_UnsetParentGroupFromThings_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ThingsServiceServer is the server API for ThingsService service.
// All implementations must embed UnimplementedThingsServiceServer
// for forward compatibility.
//
// ThingsService is a service that provides things authorization functionalities
// for magistrala services.
type ThingsServiceServer interface {
	// Authorize checks if the thing is authorized to perform
	Authorize(context.Context, *AuthzReq) (*AuthzRes, error)
	RetrieveEntity(context.Context, *v1.RetrieveEntityReq) (*v1.RetrieveEntityRes, error)
	RetrieveEntities(context.Context, *v1.RetrieveEntitiesReq) (*v1.RetrieveEntitiesRes, error)
	AddConnections(context.Context, *v1.AddConnectionsReq) (*v1.AddConnectionsRes, error)
	RemoveConnections(context.Context, *v1.RemoveConnectionsReq) (*v1.RemoveConnectionsRes, error)
	RemoveChannelConnections(context.Context, *RemoveChannelConnectionsReq) (*RemoveChannelConnectionsRes, error)
	UnsetParentGroupFromThings(context.Context, *UnsetParentGroupFromThingsReq) (*UnsetParentGroupFromThingsRes, error)
	mustEmbedUnimplementedThingsServiceServer()
}

// UnimplementedThingsServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedThingsServiceServer struct{}

func (UnimplementedThingsServiceServer) Authorize(context.Context, *AuthzReq) (*AuthzRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Authorize not implemented")
}
func (UnimplementedThingsServiceServer) RetrieveEntity(context.Context, *v1.RetrieveEntityReq) (*v1.RetrieveEntityRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RetrieveEntity not implemented")
}
func (UnimplementedThingsServiceServer) RetrieveEntities(context.Context, *v1.RetrieveEntitiesReq) (*v1.RetrieveEntitiesRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RetrieveEntities not implemented")
}
func (UnimplementedThingsServiceServer) AddConnections(context.Context, *v1.AddConnectionsReq) (*v1.AddConnectionsRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddConnections not implemented")
}
func (UnimplementedThingsServiceServer) RemoveConnections(context.Context, *v1.RemoveConnectionsReq) (*v1.RemoveConnectionsRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveConnections not implemented")
}
func (UnimplementedThingsServiceServer) RemoveChannelConnections(context.Context, *RemoveChannelConnectionsReq) (*RemoveChannelConnectionsRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveChannelConnections not implemented")
}
func (UnimplementedThingsServiceServer) UnsetParentGroupFromThings(context.Context, *UnsetParentGroupFromThingsReq) (*UnsetParentGroupFromThingsRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UnsetParentGroupFromThings not implemented")
}
func (UnimplementedThingsServiceServer) mustEmbedUnimplementedThingsServiceServer() {}
func (UnimplementedThingsServiceServer) testEmbeddedByValue()                       {}

// UnsafeThingsServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ThingsServiceServer will
// result in compilation errors.
type UnsafeThingsServiceServer interface {
	mustEmbedUnimplementedThingsServiceServer()
}

func RegisterThingsServiceServer(s grpc.ServiceRegistrar, srv ThingsServiceServer) {
	// If the following call pancis, it indicates UnimplementedThingsServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&ThingsService_ServiceDesc, srv)
}

func _ThingsService_Authorize_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AuthzReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ThingsServiceServer).Authorize(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ThingsService_Authorize_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ThingsServiceServer).Authorize(ctx, req.(*AuthzReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ThingsService_RetrieveEntity_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v1.RetrieveEntityReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ThingsServiceServer).RetrieveEntity(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ThingsService_RetrieveEntity_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ThingsServiceServer).RetrieveEntity(ctx, req.(*v1.RetrieveEntityReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ThingsService_RetrieveEntities_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v1.RetrieveEntitiesReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ThingsServiceServer).RetrieveEntities(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ThingsService_RetrieveEntities_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ThingsServiceServer).RetrieveEntities(ctx, req.(*v1.RetrieveEntitiesReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ThingsService_AddConnections_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v1.AddConnectionsReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ThingsServiceServer).AddConnections(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ThingsService_AddConnections_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ThingsServiceServer).AddConnections(ctx, req.(*v1.AddConnectionsReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ThingsService_RemoveConnections_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v1.RemoveConnectionsReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ThingsServiceServer).RemoveConnections(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ThingsService_RemoveConnections_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ThingsServiceServer).RemoveConnections(ctx, req.(*v1.RemoveConnectionsReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ThingsService_RemoveChannelConnections_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoveChannelConnectionsReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ThingsServiceServer).RemoveChannelConnections(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ThingsService_RemoveChannelConnections_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ThingsServiceServer).RemoveChannelConnections(ctx, req.(*RemoveChannelConnectionsReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ThingsService_UnsetParentGroupFromThings_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UnsetParentGroupFromThingsReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ThingsServiceServer).UnsetParentGroupFromThings(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ThingsService_UnsetParentGroupFromThings_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ThingsServiceServer).UnsetParentGroupFromThings(ctx, req.(*UnsetParentGroupFromThingsReq))
	}
	return interceptor(ctx, in, info, handler)
}

// ThingsService_ServiceDesc is the grpc.ServiceDesc for ThingsService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ThingsService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "things.v1.ThingsService",
	HandlerType: (*ThingsServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Authorize",
			Handler:    _ThingsService_Authorize_Handler,
		},
		{
			MethodName: "RetrieveEntity",
			Handler:    _ThingsService_RetrieveEntity_Handler,
		},
		{
			MethodName: "RetrieveEntities",
			Handler:    _ThingsService_RetrieveEntities_Handler,
		},
		{
			MethodName: "AddConnections",
			Handler:    _ThingsService_AddConnections_Handler,
		},
		{
			MethodName: "RemoveConnections",
			Handler:    _ThingsService_RemoveConnections_Handler,
		},
		{
			MethodName: "RemoveChannelConnections",
			Handler:    _ThingsService_RemoveChannelConnections_Handler,
		},
		{
			MethodName: "UnsetParentGroupFromThings",
			Handler:    _ThingsService_UnsetParentGroupFromThings_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "things/v1/things.proto",
}
