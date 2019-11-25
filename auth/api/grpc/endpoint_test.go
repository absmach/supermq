// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package grpc_test

import (
	"github.com/mainflux/mainflux/auth"
	"github.com/mainflux/mainflux/auth/mocks"
	"github.com/mainflux/mainflux/users"
)

const (
	port   = 8081
	secret = "secret"
)

var (
	user = users.User{
		Email:    "john.doe@email.com",
		Password: "pass",
	}
	svc auth.Service
)

func newService() auth.Service {
	repo := mocks.NewKeyRepository()
	idp := mocks.NewIdentityProvider()

	return auth.New(repo, idp, secret)
}

func startGRPCServer(svc auth.Service, port int) {
	// listener, _ := net.Listen("tcp", fmt.Sprintf(":%d", port))
	// server := grpc.NewServer()
	// mainflux.RegisterAuthServiceServer(server, api.NewServer(mocktracer.New(), svc))
	// go server.Serve(listener)
}

// func TestIdentify(t *testing.T) {
// 	svc.Register(context.Background(), user)

// 	usersAddr := fmt.Sprintf("localhost:%d", port)
// 	conn, _ := grpc.Dial(usersAddr, grpc.WithInsecure())
// 	client := grpcapi.NewClient(mocktracer.New(), conn, time.Second)
// 	j := jwt.New("secret")
// 	token, _ := j.TemporaryKey(user.Email)

// 	cases := map[string]struct {
// 		token string
// 		id    string
// 		err   error
// 	}{
// 		"identify user with valid token":   {token, user.Email, nil},
// 		"identify user that doesn't exist": {"", "", status.Error(codes.InvalidArgument, "received invalid token request")},
// 	}

// 	for desc, tc := range cases {
// 		id, err := client.Identify(context.Background(), &mainflux.Token{Value: tc.token})
// 		assert.Equal(t, tc.id, id.GetValue(), fmt.Sprintf("%s: expected %s got %s", desc, tc.id, id.GetValue()))
// 		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s", desc, tc.err, err))
// 	}
// }

// func TestIssue(t *testing.T) {

// }
