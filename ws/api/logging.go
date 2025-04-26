// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"log/slog"
	"time"

	"github.com/absmach/supermq/ws"
)

var _ ws.Service = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger *slog.Logger
	svc    ws.Service
}

// LoggingMiddleware adds logging facilities to the websocket service.
func LoggingMiddleware(svc ws.Service, logger *slog.Logger) ws.Service {
	return &loggingMiddleware{logger, svc}
}

// Subscribe logs the subscribe request. It logs the channel and subtopic(if present) and the time it took to complete the request.
// If the request fails, it logs the error.
func (lm *loggingMiddleware) Subscribe(ctx context.Context, sessionID, clientKey, domainID, chanID, subtopic string, c *ws.Client) (err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.String("session_id", sessionID),
			slog.String("channel_id", chanID),
			slog.String("domain_id", domainID),
		}
		if subtopic != "" {
			args = append(args, "subtopic", subtopic)
		}
		if err != nil {
			args = append(args, slog.Any("error", err))
			lm.logger.Warn("Subscribe failed", args...)
			return
		}
		lm.logger.Info("Subscribe completed successfully", args...)
	}(time.Now())

	return lm.svc.Subscribe(ctx, sessionID, clientKey, domainID, chanID, subtopic, c)
}

func (lm *loggingMiddleware) Unsubscribe(ctx context.Context, sessionID, domainID, chanID, subtopic string) (err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.String("session_id", sessionID),
			slog.String("channel_id", chanID),
			slog.String("domain_id", domainID),
		}
		if subtopic != "" {
			args = append(args, "subtopic", subtopic)
		}
		if err != nil {
			args = append(args, slog.Any("error", err))
			lm.logger.Warn("Unsubscribe failed", args...)
			return
		}
		lm.logger.Info("Unsubscribe completed successfully", args...)
	}(time.Now())

	return lm.svc.Unsubscribe(ctx, sessionID, domainID, chanID, subtopic)
}
