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
	"github.com/mainflux/mainflux/auth"
	grpcapi "github.com/mainflux/mainflux/auth/api/grpc"
	"github.com/mainflux/mainflux/auth/jwt"
	"github.com/mainflux/mainflux/auth/mocks"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestMain(m *testing.M) {
	t := &testing.T{}
	serverErr := make(chan error)
	testRes := make(chan int)

	svc = newService(t)
	startGRPCServer(t, serverErr, svc, port)

	for {
		select {
		case testRes <- m.Run():
			code := <-testRes
			os.Exit(code)
		case err := <-serverErr:
			if err != nil {
				log.Fatalf("gPRC Server Terminated")
			}
		}
	}
}

func startGRPCServer(t *testing.T, serverErr chan error, svc auth.Service, port int) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	assert.Nil(t, err, fmt.Sprintf("got unexpected error while creating new listerner: %s", err))

	server := grpc.NewServer()
	mainflux.RegisterAuthServiceServer(server, grpcapi.NewServer(mocktracer.New(), svc))

	go func() {
		serverErr <- server.Serve(listener)
	}()
}

func newService(t *testing.T) auth.Service {
	repo := mocks.NewKeyRepository()
	groupRepo := mocks.NewGroupRepository()
	idProvider := uuid.NewMock()

	mockAuthzDB := map[string][]mocks.MockSubjectSet{}
	mockAuthzDB[id] = append(mockAuthzDB[id], mocks.MockSubjectSet{Object: authoritiesObj, Relation: memberRelation})
	ketoMock := mocks.NewKetoMock(mockAuthzDB)

	tokenizer := jwt.New(secret)

	return auth.New(repo, groupRepo, idProvider, tokenizer, ketoMock, loginDuration)
}
