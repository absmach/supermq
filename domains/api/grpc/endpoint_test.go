// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package grpc_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	grpcCommonV1 "github.com/absmach/supermq/api/grpc/common/v1"
	grpcDomainsV1 "github.com/absmach/supermq/api/grpc/domains/v1"
	apiutil "github.com/absmach/supermq/api/http/util"
	"github.com/absmach/supermq/domains"
	grpcapi "github.com/absmach/supermq/domains/api/grpc"
	"github.com/absmach/supermq/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	port            = 8081
	secret          = "secret"
	email           = "test@example.com"
	id              = "testID"
	clientsType     = "clients"
	usersType       = "users"
	description     = "Description"
	groupName       = "smqx"
	adminpermission = "admin"
	authoritiesObj  = "authorities"
	memberRelation  = "member"
	loginDuration   = 30 * time.Minute
	refreshDuration = 24 * time.Hour
	invalidDuration = 7 * 24 * time.Hour
	validToken      = "valid"
	inValidToken    = "invalid"
	validPolicy     = "valid"
)

var authAddr = fmt.Sprintf("localhost:%d", port)

func startGRPCServer(svc domains.Service, port int) *grpc.Server {
	listener, _ := net.Listen("tcp", fmt.Sprintf(":%d", port))
	server := grpc.NewServer()
	grpcDomainsV1.RegisterDomainsServiceServer(server, grpcapi.NewDomainsServer(svc))
	go func() {
		err := server.Serve(listener)
		assert.Nil(&testing.T{}, err, fmt.Sprintf(`"Unexpected error creating auth server %s"`, err))
	}()

	return server
}

func TestDeleteUserFromDomains(t *testing.T) {
	conn, err := grpc.NewClient(authAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.Nil(t, err, fmt.Sprintf("Unexpected error creating client connection %s", err))
	grpcClient := grpcapi.NewDomainsClient(conn, time.Second)

	cases := []struct {
		desc          string
		token         string
		deleteUserReq *grpcDomainsV1.DeleteUserReq
		deleteUserRes *grpcDomainsV1.DeleteUserRes
		err           error
	}{
		{
			desc:  "delete valid req",
			token: validToken,
			deleteUserReq: &grpcDomainsV1.DeleteUserReq{
				Id: id,
			},
			deleteUserRes: &grpcDomainsV1.DeleteUserRes{Deleted: true},
			err:           nil,
		},
		{
			desc:          "delete invalid req with invalid token",
			token:         inValidToken,
			deleteUserReq: &grpcDomainsV1.DeleteUserReq{},
			deleteUserRes: &grpcDomainsV1.DeleteUserRes{Deleted: false},
			err:           apiutil.ErrMissingID,
		},
		{
			desc:  "delete invalid req with invalid token",
			token: inValidToken,
			deleteUserReq: &grpcDomainsV1.DeleteUserReq{
				Id: id,
			},
			deleteUserRes: &grpcDomainsV1.DeleteUserRes{Deleted: false},
			err:           apiutil.ErrMissingPolicyEntityType,
		},
	}
	for _, tc := range cases {
		repoCall := svc.On("DeleteUserFromDomains", mock.Anything, tc.deleteUserReq.Id).Return(tc.err)
		dpr, err := grpcClient.DeleteUserFromDomains(context.Background(), tc.deleteUserReq)
		assert.Equal(t, tc.deleteUserRes.GetDeleted(), dpr.GetDeleted(), fmt.Sprintf("%s: expected %v got %v", tc.desc, tc.deleteUserRes.GetDeleted(), dpr.GetDeleted()))
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		repoCall.Unset()
	}
}

func TestRetrieveEntity(t *testing.T) {
	conn, err := grpc.NewClient(authAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.Nil(t, err, fmt.Sprintf("Unexpected error creating client connection %s", err))
	grpcClient := grpcapi.NewDomainsClient(conn, time.Second)

	dom := domains.Domain{
		ID:     id,
		Status: domains.EnabledStatus,
	}
	cases := []struct {
		desc        string
		token       string
		retrieveReq *grpcCommonV1.RetrieveEntityReq
		svcRes      domains.Domain
		svcErr      error
		retrieveRes *grpcCommonV1.RetrieveEntityRes
		err         error
	}{
		{
			desc:  "retrieve entity with valid req",
			token: validToken,
			retrieveReq: &grpcCommonV1.RetrieveEntityReq{
				Id: id,
			},
			retrieveRes: &grpcCommonV1.RetrieveEntityRes{
				Entity: &grpcCommonV1.EntityBasic{
					Id:     id,
					Status: uint32(domains.EnabledStatus),
				},
			},
			err: nil,
		},
	}
	for _, tc := range cases {
		svcCall := svc.On("RetrieveEntity", mock.Anything, tc.retrieveReq.Id).Return(tc.svcRes, tc.svcErr)
		dpr, err := grpcClient.RetrieveEntity(context.Background(), tc.retrieveReq)
		assert.Equal(t, tc.retrieveRes.Entity, dpr.Entity, fmt.Sprintf("%s: expected %v got %v", tc.desc, tc.retrieveRes.Entity, dpr.Entity))
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		svcCall.Unset()
	}
}
