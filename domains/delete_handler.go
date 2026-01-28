// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

// The DeleteHandler is a cron job that runs periodically to delete domains that have been marked as deleted
// for a certain period of time together with the domain's policies from the auth service.
// The handler runs in a separate goroutine and checks for domains that have been marked as deleted for a certain period of time.
// If the domain has been marked as deleted for more than the specified period,
// the handler deletes the domain's policies from the auth service and deletes the domain from the database.

package domains

import (
	"context"
	"log/slog"
	"time"

	grpcChannelsV1 "github.com/absmach/supermq/api/grpc/channels/v1"
	grpcClientsV1 "github.com/absmach/supermq/api/grpc/clients/v1"
	grpcCommonV1 "github.com/absmach/supermq/api/grpc/common/v1"
	grpcDomainsV1 "github.com/absmach/supermq/api/grpc/domains/v1"
	grpcGroupsV1 "github.com/absmach/supermq/api/grpc/groups/v1"
	"github.com/absmach/supermq/pkg/policies"
)

const defLimit = uint64(100)

type handler struct {
	domains       Repository
	channels      grpcChannelsV1.ChannelsServiceClient
	clients       grpcClientsV1.ClientsServiceClient
	groups        grpcGroupsV1.GroupsServiceClient
	policies      policies.Service
	checkInterval time.Duration
	deleteAfter   time.Duration
	logger        *slog.Logger
}

func NewDeleteHandler(ctx context.Context, domains Repository, policyService policies.Service, domainsClient grpcDomainsV1.DomainsServiceClient, channelsClient grpcChannelsV1.ChannelsServiceClient,
	clientsClient grpcClientsV1.ClientsServiceClient, groupsClient grpcGroupsV1.GroupsServiceClient, defCheckInterval, deleteAfter time.Duration, logger *slog.Logger) {
	handler := &handler{
		domains:       domains,
		channels:      channelsClient,
		clients:       clientsClient,
		groups:        groupsClient,
		policies:      policyService,
		checkInterval: defCheckInterval,
		deleteAfter:   deleteAfter,
		logger:        logger,
	}

	go func() {
		ticker := time.NewTicker(handler.checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				handler.handle(ctx)
			}
		}
	}()
}

func (h *handler) handle(ctx context.Context) {
	pm := Page{Limit: defLimit, Offset: 0, Status: DeletedStatus}

	for {
		domainsPage, err := h.domains.ListDomains(ctx, pm)
		if err != nil {
			h.logger.Error("failed to list deleted domains", "error", err)
			break
		}
		if domainsPage.Total == 0 {
			break
		}

		for _, domain := range domainsPage.Domains {
			if time.Since(domain.UpdatedAt) < h.deleteAfter {
				continue
			}

			res, err := h.channels.DeleteDomainChannels(ctx, &grpcCommonV1.DeleteDomainEntitiesReq{
				DomainId: domain.ID,
			})
			if err != nil || !res.GetDeleted() {
				h.logger.Error("failed to delete domain channels", "domain_id", domain.ID, "error", err)
				continue
			}

			res, err = h.clients.DeleteDomainClients(ctx, &grpcCommonV1.DeleteDomainEntitiesReq{
				DomainId: domain.ID,
			})
			if err != nil || !res.GetDeleted() {
				h.logger.Error("failed to delete domain clients", "domain_id", domain.ID, "error", err)
				continue
			}

			res, err = h.groups.DeleteDomainGroups(ctx, &grpcCommonV1.DeleteDomainEntitiesReq{
				DomainId: domain.ID,
			})
			if err != nil || !res.GetDeleted() {
				h.logger.Error("failed to delete domain groups", "domain_id", domain.ID, "error", err)
				continue
			}

			h.logger.Info("deleted domain", "domain_id", domain.ID)
		}
	}
}
