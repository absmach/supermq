// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.0
// source: auth/v1/auth.proto

package v1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	AuthService_Authorize_FullMethodName       = "/auth.v1.AuthService/Authorize"
	AuthService_AuthorizePAT_FullMethodName    = "/auth.v1.AuthService/AuthorizePAT"
	AuthService_Authenticate_FullMethodName    = "/auth.v1.AuthService/Authenticate"
	AuthService_AuthenticatePAT_FullMethodName = "/auth.v1.AuthService/AuthenticatePAT"
)

// AuthServiceClient is the client API for AuthService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// AuthService is a service that provides authentication
// and authorization functionalities for SuperMQ services.
type AuthServiceClient interface {
	Authorize(ctx context.Context, in *AuthZReq, opts ...grpc.CallOption) (*AuthZRes, error)
	AuthorizePAT(ctx context.Context, in *AuthZPatReq, opts ...grpc.CallOption) (*AuthZRes, error)
	Authenticate(ctx context.Context, in *AuthNReq, opts ...grpc.CallOption) (*AuthNRes, error)
	AuthenticatePAT(ctx context.Context, in *AuthNReq, opts ...grpc.CallOption) (*AuthNRes, error)
}

type authServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAuthServiceClient(cc grpc.ClientConnInterface) AuthServiceClient {
	return &authServiceClient{cc}
}

func (c *authServiceClient) Authorize(ctx context.Context, in *AuthZReq, opts ...grpc.CallOption) (*AuthZRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AuthZRes)
	err := c.cc.Invoke(ctx, AuthService_Authorize_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authServiceClient) AuthorizePAT(ctx context.Context, in *AuthZPatReq, opts ...grpc.CallOption) (*AuthZRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AuthZRes)
	err := c.cc.Invoke(ctx, AuthService_AuthorizePAT_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authServiceClient) Authenticate(ctx context.Context, in *AuthNReq, opts ...grpc.CallOption) (*AuthNRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AuthNRes)
	err := c.cc.Invoke(ctx, AuthService_Authenticate_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authServiceClient) AuthenticatePAT(ctx context.Context, in *AuthNReq, opts ...grpc.CallOption) (*AuthNRes, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(AuthNRes)
	err := c.cc.Invoke(ctx, AuthService_AuthenticatePAT_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AuthServiceServer is the server API for AuthService service.
// All implementations must embed UnimplementedAuthServiceServer
// for forward compatibility.
//
// AuthService is a service that provides authentication
// and authorization functionalities for SuperMQ services.
type AuthServiceServer interface {
	Authorize(context.Context, *AuthZReq) (*AuthZRes, error)
	AuthorizePAT(context.Context, *AuthZPatReq) (*AuthZRes, error)
	Authenticate(context.Context, *AuthNReq) (*AuthNRes, error)
	AuthenticatePAT(context.Context, *AuthNReq) (*AuthNRes, error)
	mustEmbedUnimplementedAuthServiceServer()
}

// UnimplementedAuthServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedAuthServiceServer struct{}

func (UnimplementedAuthServiceServer) Authorize(context.Context, *AuthZReq) (*AuthZRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Authorize not implemented")
}
func (UnimplementedAuthServiceServer) AuthorizePAT(context.Context, *AuthZPatReq) (*AuthZRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AuthorizePAT not implemented")
}
func (UnimplementedAuthServiceServer) Authenticate(context.Context, *AuthNReq) (*AuthNRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Authenticate not implemented")
}
func (UnimplementedAuthServiceServer) AuthenticatePAT(context.Context, *AuthNReq) (*AuthNRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AuthenticatePAT not implemented")
}
func (UnimplementedAuthServiceServer) mustEmbedUnimplementedAuthServiceServer() {}
func (UnimplementedAuthServiceServer) testEmbeddedByValue()                     {}

// UnsafeAuthServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AuthServiceServer will
// result in compilation errors.
type UnsafeAuthServiceServer interface {
	mustEmbedUnimplementedAuthServiceServer()
}

func RegisterAuthServiceServer(s grpc.ServiceRegistrar, srv AuthServiceServer) {
	// If the following call pancis, it indicates UnimplementedAuthServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&AuthService_ServiceDesc, srv)
}

func _AuthService_Authorize_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AuthZReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServiceServer).Authorize(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AuthService_Authorize_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServiceServer).Authorize(ctx, req.(*AuthZReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthService_AuthorizePAT_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AuthZPatReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServiceServer).AuthorizePAT(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AuthService_AuthorizePAT_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServiceServer).AuthorizePAT(ctx, req.(*AuthZPatReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthService_Authenticate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AuthNReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServiceServer).Authenticate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AuthService_Authenticate_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServiceServer).Authenticate(ctx, req.(*AuthNReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthService_AuthenticatePAT_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AuthNReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServiceServer).AuthenticatePAT(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AuthService_AuthenticatePAT_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServiceServer).AuthenticatePAT(ctx, req.(*AuthNReq))
	}
	return interceptor(ctx, in, info, handler)
}

// AuthService_ServiceDesc is the grpc.ServiceDesc for AuthService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AuthService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "auth.v1.AuthService",
	HandlerType: (*AuthServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Authorize",
			Handler:    _AuthService_Authorize_Handler,
		},
		{
			MethodName: "Authenticate",
			Handler:    _AuthService_Authenticate_Handler,
		},
		{
			MethodName: "AuthenticatePAT",
			Handler:    _AuthService_AuthenticatePAT_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "auth/v1/auth.proto",
}
