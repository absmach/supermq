// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/absmach/supermq/auditlogs"
	"github.com/absmach/supermq/pkg/authn"
	"github.com/go-chi/chi/v5/middleware"
)

var _ auditlogs.Service = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger *slog.Logger
	svc    auditlogs.Service
}

func LoggingMiddleware(svc auditlogs.Service, logger *slog.Logger) auditlogs.Service {
	return &loggingMiddleware{
		logger,
		svc,
	}
}

func (lm *loggingMiddleware) Save(ctx context.Context, log auditlogs.AuditLog) (err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
		}
		if err != nil {
			args = append(args, slog.String("error", err.Error()))
			lm.logger.Warn("Save audit log failed", args...)
			return
		}
		lm.logger.Info("Save audit log completed successfully", args...)
	}(time.Now())
	return lm.svc.Save(ctx, log)
}

func (lm *loggingMiddleware) RetrieveByID(ctx context.Context, session authn.Session, id string) (log auditlogs.AuditLog, err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.String("domain_id", session.DomainID),
			slog.String("request_id", middleware.GetReqID(ctx)),
		}
		if err != nil {
			args = append(args, slog.String("error", err.Error()))
			lm.logger.Warn("Retrieve audit log by ID failed", args...)
			return
		}
		lm.logger.Info("Retrieve audit log by ID completed successfully", args...)
	}(time.Now())
	return lm.svc.RetrieveByID(ctx, session, id)
}

func (lm *loggingMiddleware) RetrieveAll(ctx context.Context, session authn.Session, pm auditlogs.Page) (page auditlogs.AuditLogPage, err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.String("domain_id", session.DomainID),
			slog.String("request_id", middleware.GetReqID(ctx)),
		}
		if err != nil {
			args = append(args, slog.String("error", err.Error()))
			lm.logger.Warn("Retrieve all audit logs failed", args...)
			return
		}
		lm.logger.Info("Retrieve all audit logs completed successfully", args...)
	}(time.Now())
	return lm.svc.RetrieveAll(ctx, session, pm)
}
