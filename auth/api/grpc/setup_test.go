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
	"google.golang.org/grpc"
)

func TestMain(m *testing.M) {
	serverErr := make(chan error)
	done := make(chan interface{}, 1)
	svc = newService()
	server := startGRPCServer(serverErr, done)

	go func() {
		for {
			select {
			case <-done:
				return
			case err := <-serverErr:
				if err != nil {
					log.Fatalln("gPRC Server Terminated : ", err)
				}
			}
		}
	}()

	code := m.Run()
	done <- true
	done <- true

	server.Stop()
	os.Exit(code)
}

func startGRPCServer(serverErr chan error, done chan interface{}) *grpc.Server {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("got unexpected error while creating new listerner: %s", err)
	}

	server := grpc.NewServer()
	mainflux.RegisterAuthServiceServer(server, grpcapi.NewServer(mocktracer.New(), svc))

	go func(done chan interface{}, server *grpc.Server) {
		for {
			select {
			case serverErr <- server.Serve(listener):
				return
			case <-done:
				return
			}

		}
	}(done, server)

	return server
}

func newService() auth.Service {
	repo := mocks.NewKeyRepository()
	groupRepo := mocks.NewGroupRepository()
	idProvider := uuid.NewMock()

	mockAuthzDB := map[string][]mocks.MockSubjectSet{}
	mockAuthzDB[id] = append(mockAuthzDB[id], mocks.MockSubjectSet{Object: authoritiesObj, Relation: memberRelation})
	ketoMock := mocks.NewKetoMock(mockAuthzDB)

	tokenizer := jwt.New(secret)

	return auth.New(repo, groupRepo, idProvider, tokenizer, ketoMock, loginDuration)
}
