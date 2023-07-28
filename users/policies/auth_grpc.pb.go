// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.23.3
// source: users/policies/auth.proto

package policies

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	AuthService_Identify_FullMethodName  = "/mainflux.users.policies.AuthService/Identify"
	AuthService_Authorize_FullMethodName = "/mainflux.users.policies.AuthService/Authorize"
)

// AuthServiceClient is the client API for AuthService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type AuthServiceClient interface {
	Identify(ctx context.Context, in *IdentifyReq, opts ...grpc.CallOption) (*IdentifyRes, error)
	// Authorize authorizes the given subject to perform the given action on the
	// given object.
	Authorize(ctx context.Context, in *AuthorizeReq, opts ...grpc.CallOption) (*AuthorizeRes, error)
}

type authServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewAuthServiceClient(cc grpc.ClientConnInterface) AuthServiceClient {
	return &authServiceClient{cc}
}

func (c *authServiceClient) Identify(ctx context.Context, in *IdentifyReq, opts ...grpc.CallOption) (*IdentifyRes, error) {
	out := new(IdentifyRes)
	err := c.cc.Invoke(ctx, AuthService_Identify_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *authServiceClient) Authorize(ctx context.Context, in *AuthorizeReq, opts ...grpc.CallOption) (*AuthorizeRes, error) {
	out := new(AuthorizeRes)
	err := c.cc.Invoke(ctx, AuthService_Authorize_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// AuthServiceServer is the server API for AuthService service.
// All implementations must embed UnimplementedAuthServiceServer
// for forward compatibility
type AuthServiceServer interface {
	Identify(context.Context, *IdentifyReq) (*IdentifyRes, error)
	// Authorize authorizes the given subject to perform the given action on the
	// given object.
	Authorize(context.Context, *AuthorizeReq) (*AuthorizeRes, error)
	mustEmbedUnimplementedAuthServiceServer()
}

// UnimplementedAuthServiceServer must be embedded to have forward compatible implementations.
type UnimplementedAuthServiceServer struct {
}

func (UnimplementedAuthServiceServer) Identify(context.Context, *IdentifyReq) (*IdentifyRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Identify not implemented")
}
func (UnimplementedAuthServiceServer) Authorize(context.Context, *AuthorizeReq) (*AuthorizeRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Authorize not implemented")
}
func (UnimplementedAuthServiceServer) mustEmbedUnimplementedAuthServiceServer() {}

// UnsafeAuthServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to AuthServiceServer will
// result in compilation errors.
type UnsafeAuthServiceServer interface {
	mustEmbedUnimplementedAuthServiceServer()
}

func RegisterAuthServiceServer(s grpc.ServiceRegistrar, srv AuthServiceServer) {
	s.RegisterService(&AuthService_ServiceDesc, srv)
}

func _AuthService_Identify_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IdentifyReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthServiceServer).Identify(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: AuthService_Identify_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AuthServiceServer).Identify(ctx, req.(*IdentifyReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _AuthService_Authorize_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AuthorizeReq)
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
		return srv.(AuthServiceServer).Authorize(ctx, req.(*AuthorizeReq))
	}
	return interceptor(ctx, in, info, handler)
}

// AuthService_ServiceDesc is the grpc.ServiceDesc for AuthService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var AuthService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "mainflux.users.policies.AuthService",
	HandlerType: (*AuthServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Identify",
			Handler:    _AuthService_Identify_Handler,
		},
		{
			MethodName: "Authorize",
			Handler:    _AuthService_Authorize_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "users/policies/auth.proto",
}
