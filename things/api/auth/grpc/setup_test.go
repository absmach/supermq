// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package grpc_test

import (
	"fmt"
	"log"
	"net"
	"os"
	"testing"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/mainflux/mainflux/things"
	grpcapi "github.com/mainflux/mainflux/things/api/auth/grpc"
	"github.com/mainflux/mainflux/things/mocks"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

const (
	port  = 7000
	token = "token"
	wrong = "wrong"
	email = "john.doe@email.com"
)

var svc things.Service

func TestMain(m *testing.M) {
	serverErr := make(chan error)
	testRes := make(chan int)
	startServer(&testing.T{}, serverErr)

	go func() {
		for {
			select {
			case code := <-testRes:
				os.Exit(code)
			case err := <-serverErr:
				if err != nil {
					log.Fatalf("gPRC Server Terminated")
				}
			}
		}
	}()

	testRes <- m.Run()
}

func startServer(t *testing.T, serverErr chan error) {
	svc = newService(map[string]string{token: email})
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	assert.Nil(t, err, fmt.Sprintf("got unexpected error while creating new listener: %s", err))

	server := grpc.NewServer()
	mainflux.RegisterThingsServiceServer(server, grpcapi.NewServer(mocktracer.New(), svc))

	go func() {
		serverErr <- server.Serve(listener)
	}()
}

func newService(tokens map[string]string) things.Service {
	policies := []mocks.MockSubjectSet{{Object: "users", Relation: "member"}}
	auth := mocks.NewAuthService(tokens, map[string][]mocks.MockSubjectSet{email: policies})
	conns := make(chan mocks.Connection)
	thingsRepo := mocks.NewThingRepository(conns)
	channelsRepo := mocks.NewChannelRepository(thingsRepo, conns)
	chanCache := mocks.NewChannelCache()
	thingCache := mocks.NewThingCache()
	idProvider := uuid.NewMock()

	return things.New(auth, thingsRepo, channelsRepo, chanCache, thingCache, idProvider)
}
