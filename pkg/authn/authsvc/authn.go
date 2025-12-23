// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package authsvc

import (
	"context"
	"strings"

	grpcAuthV1 "github.com/absmach/supermq/api/grpc/auth/v1"
	"github.com/absmach/supermq/auth/api/grpc/auth"
	"github.com/absmach/supermq/pkg/authn"
	"github.com/absmach/supermq/pkg/errors"
	"github.com/absmach/supermq/pkg/grpcclient"
	grpchealth "google.golang.org/grpc/health/grpc_health_v1"
)

const patPrefix = "pat_"

var errTokenTypeMismatch = errors.NewAuthNError("token type mismatch: token format does not match the service-returned token type")

type authentication struct {
	authSvcClient grpcAuthV1.AuthServiceClient
}

var _ authn.Authentication = (*authentication)(nil)

func NewAuthentication(ctx context.Context, cfg grpcclient.Config) (authn.Authentication, grpcclient.Handler, error) {
	client, err := grpcclient.NewHandler(cfg)
	if err != nil {
		return nil, nil, err
	}

	health := grpchealth.NewHealthClient(client.Connection())
	resp, err := health.Check(ctx, &grpchealth.HealthCheckRequest{
		Service: "auth",
	})
	if err != nil || resp.GetStatus() != grpchealth.HealthCheckResponse_SERVING {
		return nil, nil, grpcclient.ErrSvcNotServing
	}
	authSvcClient := auth.NewAuthClient(client.Connection(), cfg.Timeout)
	return authentication{authSvcClient}, client, nil
}

func (a authentication) Authenticate(ctx context.Context, token string) (authn.Session, error) {
	res, err := a.authSvcClient.Authenticate(ctx, &grpcAuthV1.AuthNReq{Token: token})
	if err != nil {
		return authn.Session{}, errors.Wrap(errors.ErrAuthentication, err)
	}

	isPAT := strings.HasPrefix(token, patPrefix)
	tokenType := authn.TokenType(res.GetTokenType())

	if isPAT && tokenType != authn.PersonalAccessToken {
		return authn.Session{}, errTokenTypeMismatch
	}
	if !isPAT && tokenType == authn.PersonalAccessToken {
		return authn.Session{}, errTokenTypeMismatch
	}

	switch tokenType {
	case authn.PersonalAccessToken:
		return authn.Session{Type: authn.PersonalAccessToken, PatID: res.GetId(), UserID: res.GetUserId(), Role: authn.Role(res.GetUserRole())}, nil
	default:
		return authn.Session{Type: authn.AccessToken, UserID: res.GetUserId(), Role: authn.Role(res.GetUserRole()), Verified: res.GetVerified()}, nil
	}
}
