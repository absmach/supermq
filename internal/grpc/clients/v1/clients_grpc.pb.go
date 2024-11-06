// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.28.2
// source: clients/v1/clients.proto

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
	ClientsService_Authenticate_FullMethodName               = "/things.v1.ClientsService/Authenticate"
	ClientsService_RetrieveEntity_FullMethodName             = "/things.v1.ClientsService/RetrieveEntity"
	ClientsService_RetrieveEntities_FullMethodName           = "/things.v1.ClientsService/RetrieveEntities"
	ClientsService_AddConnections_FullMethodName             = "/things.v1.ClientsService/AddConnections"
	ClientsService_RemoveConnections_FullMethodName          = "/things.v1.ClientsService/RemoveConnections"
	ClientsService_RemoveChannelConnections_FullMethodName   = "/things.v1.ClientsService/RemoveChannelConnections"
	ClientsService_UnsetParentGroupFromClient_FullMethodName = "/things.v1.ClientsService/UnsetParentGroupFromClient"
)

// ClientsServiceClient is the client API for ClientsService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// ClientsService is a service that provides things authorization functionalities
// for magistrala services.
type ClientsServiceClient interface {
	// Authorize checks if the thing is authorized to perform
	Authenticate(ctx context.Context, in *AuthnReq, opts ...grpc.CallOption) (*AuthnRes, error)
	RetrieveEntity(ctx context.Context, in *v1.RetrieveEntityReq, opts ...grpc.CallOption) (*v1.RetrieveEntityRes, error)
	RetrieveEntities(ctx context.Context, in *v1.RetrieveEntitiesReq, opts ...grpc.CallOption) (*v1.RetrieveEntitiesRes, error)
	AddConnections(ctx context.Context, in *v1.AddConnectionsReq, opts ...grpc.CallOption) (*v1.AddConnectionsRes, error)
	RemoveConnections(ctx context.Context, in *v1.RemoveConnectionsReq, opts ...grpc.CallOption) (*v1.RemoveConnectionsRes, error)
	RemoveChannelConnections(ctx context.Context, in *RemoveChannelConnectionsReq, opts ...grpc.CallOption) (*RemoveChannelConnectionsRes, error)
	UnsetParentGroupFromClient(ctx context.Context, in *UnsetParentGroupFromClientReq, opts ...grpc.CallOption) (*UnsetParentGroupFromClientRes, error)
}

type clientsServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewClientsServiceClient(cc grpc.ClientConnInterface) ClientsServiceClient {
	return &clientsServiceClient{cc}
}

func (c *clientsServiceClient) Authenticate(ctx context.Context, in *AuthnReq, opts ...grpc.CallOption) (*AuthnRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AuthnRes)
	err := c.cc.Invoke(ctx, ClientsService_Authenticate_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientsServiceClient) RetrieveEntity(ctx context.Context, in *v1.RetrieveEntityReq, opts ...grpc.CallOption) (*v1.RetrieveEntityRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(v1.RetrieveEntityRes)
	err := c.cc.Invoke(ctx, ClientsService_RetrieveEntity_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientsServiceClient) RetrieveEntities(ctx context.Context, in *v1.RetrieveEntitiesReq, opts ...grpc.CallOption) (*v1.RetrieveEntitiesRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(v1.RetrieveEntitiesRes)
	err := c.cc.Invoke(ctx, ClientsService_RetrieveEntities_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientsServiceClient) AddConnections(ctx context.Context, in *v1.AddConnectionsReq, opts ...grpc.CallOption) (*v1.AddConnectionsRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(v1.AddConnectionsRes)
	err := c.cc.Invoke(ctx, ClientsService_AddConnections_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientsServiceClient) RemoveConnections(ctx context.Context, in *v1.RemoveConnectionsReq, opts ...grpc.CallOption) (*v1.RemoveConnectionsRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(v1.RemoveConnectionsRes)
	err := c.cc.Invoke(ctx, ClientsService_RemoveConnections_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientsServiceClient) RemoveChannelConnections(ctx context.Context, in *RemoveChannelConnectionsReq, opts ...grpc.CallOption) (*RemoveChannelConnectionsRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(RemoveChannelConnectionsRes)
	err := c.cc.Invoke(ctx, ClientsService_RemoveChannelConnections_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clientsServiceClient) UnsetParentGroupFromClient(ctx context.Context, in *UnsetParentGroupFromClientReq, opts ...grpc.CallOption) (*UnsetParentGroupFromClientRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(UnsetParentGroupFromClientRes)
	err := c.cc.Invoke(ctx, ClientsService_UnsetParentGroupFromClient_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ClientsServiceServer is the server API for ClientsService service.
// All implementations must embed UnimplementedClientsServiceServer
// for forward compatibility.
//
// ClientsService is a service that provides things authorization functionalities
// for magistrala services.
type ClientsServiceServer interface {
	// Authorize checks if the thing is authorized to perform
	Authenticate(context.Context, *AuthnReq) (*AuthnRes, error)
	RetrieveEntity(context.Context, *v1.RetrieveEntityReq) (*v1.RetrieveEntityRes, error)
	RetrieveEntities(context.Context, *v1.RetrieveEntitiesReq) (*v1.RetrieveEntitiesRes, error)
	AddConnections(context.Context, *v1.AddConnectionsReq) (*v1.AddConnectionsRes, error)
	RemoveConnections(context.Context, *v1.RemoveConnectionsReq) (*v1.RemoveConnectionsRes, error)
	RemoveChannelConnections(context.Context, *RemoveChannelConnectionsReq) (*RemoveChannelConnectionsRes, error)
	UnsetParentGroupFromClient(context.Context, *UnsetParentGroupFromClientReq) (*UnsetParentGroupFromClientRes, error)
	mustEmbedUnimplementedClientsServiceServer()
}

// UnimplementedClientsServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedClientsServiceServer struct{}

func (UnimplementedClientsServiceServer) Authenticate(context.Context, *AuthnReq) (*AuthnRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Authenticate not implemented")
}
func (UnimplementedClientsServiceServer) RetrieveEntity(context.Context, *v1.RetrieveEntityReq) (*v1.RetrieveEntityRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RetrieveEntity not implemented")
}
func (UnimplementedClientsServiceServer) RetrieveEntities(context.Context, *v1.RetrieveEntitiesReq) (*v1.RetrieveEntitiesRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RetrieveEntities not implemented")
}
func (UnimplementedClientsServiceServer) AddConnections(context.Context, *v1.AddConnectionsReq) (*v1.AddConnectionsRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddConnections not implemented")
}
func (UnimplementedClientsServiceServer) RemoveConnections(context.Context, *v1.RemoveConnectionsReq) (*v1.RemoveConnectionsRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveConnections not implemented")
}
func (UnimplementedClientsServiceServer) RemoveChannelConnections(context.Context, *RemoveChannelConnectionsReq) (*RemoveChannelConnectionsRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RemoveChannelConnections not implemented")
}
func (UnimplementedClientsServiceServer) UnsetParentGroupFromClient(context.Context, *UnsetParentGroupFromClientReq) (*UnsetParentGroupFromClientRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UnsetParentGroupFromClient not implemented")
}
func (UnimplementedClientsServiceServer) mustEmbedUnimplementedClientsServiceServer() {}
func (UnimplementedClientsServiceServer) testEmbeddedByValue()                        {}

// UnsafeClientsServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ClientsServiceServer will
// result in compilation errors.
type UnsafeClientsServiceServer interface {
	mustEmbedUnimplementedClientsServiceServer()
}

func RegisterClientsServiceServer(s grpc.ServiceRegistrar, srv ClientsServiceServer) {
	// If the following call pancis, it indicates UnimplementedClientsServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&ClientsService_ServiceDesc, srv)
}

func _ClientsService_Authenticate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AuthnReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientsServiceServer).Authenticate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientsService_Authenticate_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientsServiceServer).Authenticate(ctx, req.(*AuthnReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientsService_RetrieveEntity_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v1.RetrieveEntityReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientsServiceServer).RetrieveEntity(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientsService_RetrieveEntity_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientsServiceServer).RetrieveEntity(ctx, req.(*v1.RetrieveEntityReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientsService_RetrieveEntities_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v1.RetrieveEntitiesReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientsServiceServer).RetrieveEntities(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientsService_RetrieveEntities_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientsServiceServer).RetrieveEntities(ctx, req.(*v1.RetrieveEntitiesReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientsService_AddConnections_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v1.AddConnectionsReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientsServiceServer).AddConnections(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientsService_AddConnections_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientsServiceServer).AddConnections(ctx, req.(*v1.AddConnectionsReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientsService_RemoveConnections_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v1.RemoveConnectionsReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientsServiceServer).RemoveConnections(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientsService_RemoveConnections_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientsServiceServer).RemoveConnections(ctx, req.(*v1.RemoveConnectionsReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientsService_RemoveChannelConnections_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RemoveChannelConnectionsReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientsServiceServer).RemoveChannelConnections(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientsService_RemoveChannelConnections_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientsServiceServer).RemoveChannelConnections(ctx, req.(*RemoveChannelConnectionsReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ClientsService_UnsetParentGroupFromClient_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UnsetParentGroupFromClientReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ClientsServiceServer).UnsetParentGroupFromClient(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ClientsService_UnsetParentGroupFromClient_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ClientsServiceServer).UnsetParentGroupFromClient(ctx, req.(*UnsetParentGroupFromClientReq))
	}
	return interceptor(ctx, in, info, handler)
}

// ClientsService_ServiceDesc is the grpc.ServiceDesc for ClientsService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ClientsService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "things.v1.ClientsService",
	HandlerType: (*ClientsServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Authenticate",
			Handler:    _ClientsService_Authenticate_Handler,
		},
		{
			MethodName: "RetrieveEntity",
			Handler:    _ClientsService_RetrieveEntity_Handler,
		},
		{
			MethodName: "RetrieveEntities",
			Handler:    _ClientsService_RetrieveEntities_Handler,
		},
		{
			MethodName: "AddConnections",
			Handler:    _ClientsService_AddConnections_Handler,
		},
		{
			MethodName: "RemoveConnections",
			Handler:    _ClientsService_RemoveConnections_Handler,
		},
		{
			MethodName: "RemoveChannelConnections",
			Handler:    _ClientsService_RemoveChannelConnections_Handler,
		},
		{
			MethodName: "UnsetParentGroupFromClient",
			Handler:    _ClientsService_UnsetParentGroupFromClient_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "clients/v1/clients.proto",
}
