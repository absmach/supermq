// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package grpc_test

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/auth"
	grpcapi "github.com/mainflux/mainflux/auth/api/grpc"
	"github.com/mainflux/mainflux/auth/mocks"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	port   = 8081
	secret = "secret"
	email  = "test@example.com"
)

var svc auth.Service

func newService() auth.Service {
	repo := mocks.NewKeyRepository()
	idp := mocks.NewIdentityProvider()

	return auth.New(repo, idp, secret)
}

func startGRPCServer(svc auth.Service, port int) {
	listener, _ := net.Listen("tcp", fmt.Sprintf(":%d", port))
	server := grpc.NewServer()
	mainflux.RegisterAuthServiceServer(server, grpcapi.NewServer(mocktracer.New(), svc))
	go server.Serve(listener)
}

func TestIssue(t *testing.T) {
	loginKey, err := svc.Issue(context.Background(), email, auth.Key{Type: auth.LoginKey, IssuedAt: time.Now()})
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))

	authAddr := fmt.Sprintf("localhost:%d", port)
	conn, _ := grpc.Dial(authAddr, grpc.WithInsecure())
	client := grpcapi.NewClient(mocktracer.New(), conn, time.Second)

	cases := map[string]struct {
		token string
		id    string
		kind  uint32
		err   error
	}{
		"issue for user with valid token":   {"", email, auth.LoginKey, nil},
		"issue for user that doesn't exist": {"", loginKey.Secret, 32, status.Error(codes.InvalidArgument, "received invalid token request")},
	}

	for desc, tc := range cases {
		_, err := client.Issue(context.Background(), &mainflux.IssueReq{Issuer: tc.id, Type: tc.kind})
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s", desc, tc.err, err))
	}
}

func TestIdentify(t *testing.T) {
	loginKey, err := svc.Issue(context.Background(), email, auth.Key{Type: auth.LoginKey, IssuedAt: time.Now()})
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))

	authAddr := fmt.Sprintf("localhost:%d", port)
	conn, _ := grpc.Dial(authAddr, grpc.WithInsecure())
	client := grpcapi.NewClient(mocktracer.New(), conn, time.Second)

	cases := map[string]struct {
		token string
		id    string
		err   error
	}{
		"identify user with valid token":   {loginKey.Secret, email, nil},
		"identify user that doesn't exist": {"", "", status.Error(codes.InvalidArgument, "received invalid token request")},
	}

	for desc, tc := range cases {
		id, err := client.Identify(context.Background(), &mainflux.Token{Value: tc.token})
		assert.Equal(t, tc.id, id.GetValue(), fmt.Sprintf("%s: expected %s got %s", desc, tc.id, id.GetValue()))
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s", desc, tc.err, err))
	}
}
