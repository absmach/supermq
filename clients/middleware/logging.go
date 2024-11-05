// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/absmach/magistrala/clients"
	"github.com/absmach/magistrala/pkg/authn"
	rmMW "github.com/absmach/magistrala/pkg/roles/rolemanager/middleware"
)

var _ clients.Service = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger *slog.Logger
	svc    clients.Service
	rmMW.RoleManagerLoggingMiddleware
}

func LoggingMiddleware(svc clients.Service, logger *slog.Logger) clients.Service {
	return &loggingMiddleware{
		logger:                       logger,
		svc:                          svc,
		RoleManagerLoggingMiddleware: rmMW.NewRoleManagerLoggingMiddleware("things", svc, logger),
	}
}

func (lm *loggingMiddleware) CreateClients(ctx context.Context, session authn.Session, clients ...clients.Client) (cs []clients.Client, err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
		}
		if err != nil {
			args = append(args, slog.String("error", err.Error()))
			lm.logger.Warn(fmt.Sprintf("Create %d things failed", len(clients)), args...)
			return
		}
		lm.logger.Info(fmt.Sprintf("Create %d things completed successfully", len(clients)), args...)
	}(time.Now())
	return lm.svc.CreateClients(ctx, session, clients...)
}

func (lm *loggingMiddleware) View(ctx context.Context, session authn.Session, id string) (c clients.Client, err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.Group("thing",
				slog.String("id", c.ID),
				slog.String("name", c.Name),
			),
		}
		if err != nil {
			args = append(args, slog.String("error", err.Error()))
			lm.logger.Warn("View thing failed", args...)
			return
		}
		lm.logger.Info("View thing completed successfully", args...)
	}(time.Now())
	return lm.svc.View(ctx, session, id)
}

func (lm *loggingMiddleware) ListClients(ctx context.Context, session authn.Session, reqUserID string, pm clients.Page) (cp clients.ClientsPage, err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.String("user_id", reqUserID),
			slog.Group("page",
				slog.Uint64("limit", pm.Limit),
				slog.Uint64("offset", pm.Offset),
				slog.Uint64("total", cp.Total),
			),
		}
		if err != nil {
			args = append(args, slog.String("error", err.Error()))
			lm.logger.Warn("List things failed", args...)
			return
		}
		lm.logger.Info("List things completed successfully", args...)
	}(time.Now())
	return lm.svc.ListClients(ctx, session, reqUserID, pm)
}

func (lm *loggingMiddleware) Update(ctx context.Context, session authn.Session, client clients.Client) (c clients.Client, err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.Group("thing",
				slog.String("id", client.ID),
				slog.String("name", client.Name),
				slog.Any("metadata", client.Metadata),
			),
		}
		if err != nil {
			args = append(args, slog.String("error", err.Error()))
			lm.logger.Warn("Update thing failed", args...)
			return
		}
		lm.logger.Info("Update thing completed successfully", args...)
	}(time.Now())
	return lm.svc.Update(ctx, session, client)
}

func (lm *loggingMiddleware) UpdateTags(ctx context.Context, session authn.Session, client clients.Client) (c clients.Client, err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.Group("thing",
				slog.String("id", c.ID),
				slog.String("name", c.Name),
				slog.Any("tags", c.Tags),
			),
		}
		if err != nil {
			args := append(args, slog.String("error", err.Error()))
			lm.logger.Warn("Update thing tags failed", args...)
			return
		}
		lm.logger.Info("Update thing tags completed successfully", args...)
	}(time.Now())
	return lm.svc.UpdateTags(ctx, session, client)
}

func (lm *loggingMiddleware) UpdateSecret(ctx context.Context, session authn.Session, oldSecret, newSecret string) (c clients.Client, err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.Group("thing",
				slog.String("id", c.ID),
				slog.String("name", c.Name),
			),
		}
		if err != nil {
			args = append(args, slog.String("error", err.Error()))
			lm.logger.Warn("Update thing secret failed", args...)
			return
		}
		lm.logger.Info("Update thing secret completed successfully", args...)
	}(time.Now())
	return lm.svc.UpdateSecret(ctx, session, oldSecret, newSecret)
}

func (lm *loggingMiddleware) Enable(ctx context.Context, session authn.Session, id string) (c clients.Client, err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.Group("thing",
				slog.String("id", id),
				slog.String("name", c.Name),
			),
		}
		if err != nil {
			args = append(args, slog.String("error", err.Error()))
			lm.logger.Warn("Enable thing failed", args...)
			return
		}
		lm.logger.Info("Enable thing completed successfully", args...)
	}(time.Now())
	return lm.svc.Enable(ctx, session, id)
}

func (lm *loggingMiddleware) Disable(ctx context.Context, session authn.Session, id string) (c clients.Client, err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.Group("thing",
				slog.String("id", id),
				slog.String("name", c.Name),
			),
		}
		if err != nil {
			args = append(args, slog.String("error", err.Error()))
			lm.logger.Warn("Disable thing failed", args...)
			return
		}
		lm.logger.Info("Disable thing completed successfully", args...)
	}(time.Now())
	return lm.svc.Disable(ctx, session, id)
}

func (lm *loggingMiddleware) Delete(ctx context.Context, session authn.Session, id string) (err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.String("thing_id", id),
		}
		if err != nil {
			args = append(args, slog.String("error", err.Error()))
			lm.logger.Warn("Delete thing failed", args...)
			return
		}
		lm.logger.Info("Delete thing completed successfully", args...)
	}(time.Now())
	return lm.svc.Delete(ctx, session, id)
}

func (lm *loggingMiddleware) SetParentGroup(ctx context.Context, session authn.Session, parentGroupID string, id string) (err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.String("parent_group_id", parentGroupID),
			slog.String("thing_id", id),
		}
		if err != nil {
			args = append(args, slog.String("error", err.Error()))
			lm.logger.Warn("Set parent group to thing failed", args...)
			return
		}
		lm.logger.Info("Set parent group to thing completed successfully", args...)
	}(time.Now())
	return lm.svc.SetParentGroup(ctx, session, parentGroupID, id)
}

func (lm *loggingMiddleware) RemoveParentGroup(ctx context.Context, session authn.Session, id string) (err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.String("client_id", id),
		}
		if err != nil {
			args = append(args, slog.String("error", err.Error()))
			lm.logger.Warn("Remove parent group from thing failed", args...)
			return
		}
		lm.logger.Info("Remove parent group from thing completed successfully", args...)
	}(time.Now())
	return lm.svc.RemoveParentGroup(ctx, session, id)
}
