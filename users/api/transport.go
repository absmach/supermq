// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"log/slog"
	"net/http"
	"regexp"

	"github.com/absmach/supermq"
	grpcTokenV1 "github.com/absmach/supermq/api/grpc/token/v1"
	smqauthn "github.com/absmach/supermq/pkg/authn"
	"github.com/absmach/supermq/pkg/oauth2"
	oauthhttp "github.com/absmach/supermq/pkg/oauth2/http"
	"github.com/absmach/supermq/pkg/oauth2/store"
	"github.com/absmach/supermq/users"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

// MakeHandler returns a HTTP handler for Users and Groups API endpoints.
// It accepts separate providers for device flow and user flow.
// For backward compatibility, if only one provider is passed, it's used for both flows.
func MakeHandler(svc users.Service, authn smqauthn.AuthNMiddleware, tokensvc grpcTokenV1.TokenServiceClient, selfRegister bool, mux *chi.Mux, logger *slog.Logger, instanceID string, pr *regexp.Regexp, idp supermq.IDProvider, cacheClient *redis.Client, userProviders, deviceProviders []oauth2.Provider) http.Handler {
	ctx := context.Background()

	mux = usersHandler(svc, authn, tokensvc, selfRegister, mux, logger, pr, idp)

	deviceStore := store.NewRedisDeviceCodeStore(ctx, cacheClient)
	oauthSvc := oauth2.NewOAuthService(deviceStore, svc, tokensvc)

	mux = oauthhttp.Handler(mux, tokensvc, oauthSvc, userProviders...)
	mux = oauthhttp.DeviceHandler(mux, tokensvc, oauthSvc, deviceProviders...)

	mux.Get("/health", supermq.Health("users", instanceID))
	mux.Handle("/metrics", promhttp.Handler())

	return mux
}
