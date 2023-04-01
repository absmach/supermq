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
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestMain(m *testing.M) {
	t := &testing.T{}
	serverErr := make(chan error)
	testRes := make(chan int)

	svc = newService(t)
	startGRPCServer(t, serverErr, svc, port)

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

func startGRPCServer(t *testing.T, serverErr chan error, svc auth.Service, port int) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	require.Nil(t, err, fmt.Sprintf("got unexpected error while creating new listerner: %s", err))

	server := grpc.NewServer()
	mainflux.RegisterAuthServiceServer(server, grpcapi.NewServer(mocktracer.New(), svc))

	done := make(chan bool)

	t.Cleanup(func() {
		fmt.Println()
		fmt.Println("Test complete called t.cleanup")
		fmt.Println()
		close(done)
	})

	go func(done <-chan bool) {
		for {
			select {
			case serverErr <- server.Serve(listener):
				close(serverErr)
				return
			case <-done:
				close(serverErr)
				// serverErr <- nil
				return
			}

		}
	}(done)
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
