// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package grpcclient

import (
	"context"

	tokengrpc "github.com/absmach/magistrala/auth/api/grpc/token"
	channelsgrpc "github.com/absmach/magistrala/channels/api/grpc"
	domainsgrpc "github.com/absmach/magistrala/domains/api/grpc"
	grpcChannelsV1 "github.com/absmach/magistrala/internal/grpc/channels/v1"
	grpcDomainsV1 "github.com/absmach/magistrala/internal/grpc/domains/v1"
	grpcThingsV1 "github.com/absmach/magistrala/internal/grpc/things/v1"
	grpcTokenV1 "github.com/absmach/magistrala/internal/grpc/token/v1"
	thingsauth "github.com/absmach/magistrala/things/api/grpc"
	grpchealth "google.golang.org/grpc/health/grpc_health_v1"
)

// SetupTokenClient loads auth services token gRPC configuration and creates new Token services gRPC client.
//
// For example:
//
// tokenClient, tokenHandler, err := grpcclient.SetupTokenClient(ctx, grpcclient.Config{}).
func SetupTokenClient(ctx context.Context, cfg Config) (grpcTokenV1.TokenServiceClient, Handler, error) {
	client, err := NewHandler(cfg)
	if err != nil {
		return nil, nil, err
	}

	health := grpchealth.NewHealthClient(client.Connection())
	resp, err := health.Check(ctx, &grpchealth.HealthCheckRequest{
		Service: "token.v1.TokenService",
	})
	if err != nil || resp.GetStatus() != grpchealth.HealthCheckResponse_SERVING {
		return nil, nil, ErrSvcNotServing
	}

	return tokengrpc.NewTokenClient(client.Connection(), cfg.Timeout), client, nil
}

// SetupDomiansClient loads domains gRPC configuration and creates a new domains gRPC client.
//
// For example:
//
// domainsClient, domainsHandler, err := grpcclient.SetupDomainsClient(ctx, grpcclient.Config{}).
func SetupDomainsClient(ctx context.Context, cfg Config) (grpcDomainsV1.DomainsServiceClient, Handler, error) {
	client, err := NewHandler(cfg)
	if err != nil {
		return nil, nil, err
	}

	health := grpchealth.NewHealthClient(client.Connection())
	resp, err := health.Check(ctx, &grpchealth.HealthCheckRequest{
		Service: "domains.v1.DomainsService",
	})
	if err != nil || resp.GetStatus() != grpchealth.HealthCheckResponse_SERVING {
		return nil, nil, ErrSvcNotServing
	}

	return domainsgrpc.NewDomainsClient(client.Connection(), cfg.Timeout), client, nil
}

// SetupThingsClient loads things gRPC configuration and creates new things gRPC client.
//
// For example:
//
// thingClient, thingHandler, err := grpcclient.SetupThings(ctx, grpcclient.Config{}).
func SetupThingsClient(ctx context.Context, cfg Config) (grpcThingsV1.ThingsServiceClient, Handler, error) {
	client, err := NewHandler(cfg)
	if err != nil {
		return nil, nil, err
	}

	if !cfg.BypassHealthCheck {
		health := grpchealth.NewHealthClient(client.Connection())
		resp, err := health.Check(ctx, &grpchealth.HealthCheckRequest{
			Service: "things.v1.ThingsService",
		})
		if err != nil || resp.GetStatus() != grpchealth.HealthCheckResponse_SERVING {
			return nil, nil, ErrSvcNotServing
		}
	}

	return thingsauth.NewClient(client.Connection(), cfg.Timeout), client, nil
}

// SetupChannelsClient loads channels gRPC configuration and creates new channels gRPC client.
//
// For example:
//
// channelClient, channelHandler, err := grpcclient.SetupChannelsClient(ctx, grpcclient.Config{}).
func SetupChannelsClient(ctx context.Context, cfg Config) (grpcChannelsV1.ChannelsServiceClient, Handler, error) {
	client, err := NewHandler(cfg)
	if err != nil {
		return nil, nil, err
	}

	if !cfg.BypassHealthCheck {
		health := grpchealth.NewHealthClient(client.Connection())
		resp, err := health.Check(ctx, &grpchealth.HealthCheckRequest{
			Service: "channels.v1.ChannelsService",
		})
		if err != nil || resp.GetStatus() != grpchealth.HealthCheckResponse_SERVING {
			return nil, nil, ErrSvcNotServing
		}
	}

	return channelsgrpc.NewClient(client.Connection(), cfg.Timeout), client, nil
}
