// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/absmach/magistrala/activitylog"
)

var _ activitylog.Service = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger  *slog.Logger
	service activitylog.Service
}

// LoggingMiddleware adds logging facilities to the adapter.
func LoggingMiddleware(service activitylog.Service, logger *slog.Logger) activitylog.Service {
	return &loggingMiddleware{
		logger:  logger,
		service: service,
	}
}

func (lm *loggingMiddleware) Save(ctx context.Context, activity activitylog.Activity) (err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.Group("activity",
				slog.String("occurred_at", activity.OccurredAt.Format(time.RFC3339Nano)),
				slog.String("operation", activity.Operation),
			),
		}
		if err != nil {
			args = append(args, slog.Any("error", err))
			lm.logger.Warn("Save activity failed to complete successfully", args...)
			return
		}
		lm.logger.Info("Save activity completed successfully", args...)
	}(time.Now())

	return lm.service.Save(ctx, activity)
}

func (lm *loggingMiddleware) RetrieveAll(ctx context.Context, token string, page activitylog.Page) (activitiesPage activitylog.ActivitiesPage, err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.Group("page",
				slog.String("operation", page.Operation),
				slog.String("entity_type", page.EntityType.String()),
				slog.Uint64("offset", page.Offset),
				slog.Uint64("limit", page.Limit),
				slog.Uint64("total", activitiesPage.Total),
			),
		}
		if err != nil {
			args = append(args, slog.Any("error", err))
			lm.logger.Warn("Retrieve all activities failed to complete successfully", args...)
			return
		}
		lm.logger.Info("Retrieve all activities completed successfully", args...)
	}(time.Now())

	return lm.service.RetrieveAll(ctx, token, page)
}
