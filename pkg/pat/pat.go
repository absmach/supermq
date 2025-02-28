// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package pat

import (
	"context"

	grpcAuthV1 "github.com/absmach/supermq/api/grpc/auth/v1"
	smqauth "github.com/absmach/supermq/auth"
	"github.com/absmach/supermq/auth/api/grpc/auth"
	"github.com/absmach/supermq/pkg/errors"
	"github.com/absmach/supermq/pkg/grpcclient"
	grpchealth "google.golang.org/grpc/health/grpc_health_v1"
)

type PatReq struct {
	UserID           string             `json:"user_id,omitempty"`           // UserID
	PatID            string             `json:"pat_id,omitempty"`            // UserID
	EntityType       smqauth.EntityType `json:"entity_type,omitempty"`       // Entity type
	OptionalDomainID string             `json:"optional_domainID,omitempty"` // Optional domain id
	Operation        smqauth.Operation  `json:"operation,omitempty"`         // Operation
	EntityID         string             `json:"entityID,omitempty"`          // EntityID
}

type Authorization interface {
	AuthorizePAT(ctx context.Context, pr PatReq) error
}

type authorization struct {
	authSvcClient grpcAuthV1.AuthServiceClient
}

func NewAuthorization(ctx context.Context, cfg grpcclient.Config) (Authorization, grpcclient.Handler, error) {
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
	return authorization{
		authSvcClient: authSvcClient,
	}, client, nil
}

func (a authorization) AuthorizePAT(ctx context.Context, pr PatReq) error {
	req := grpcAuthV1.AuthZPatReq{
		UserId:           pr.UserID,
		PatId:            pr.PatID,
		EntityType:       uint32(pr.EntityType),
		OptionalDomainId: pr.OptionalDomainID,
		Operation:        uint32(pr.Operation),
		EntityId:         pr.EntityID,
	}
	res, err := a.authSvcClient.AuthorizePAT(ctx, &req)
	if err != nil {
		return errors.Wrap(errors.ErrAuthorization, err)
	}
	if !res.Authorized {
		return errors.ErrAuthorization
	}
	return nil
}
