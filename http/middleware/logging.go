// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"log/slog"
	"time"

	smqhttp "github.com/absmach/supermq/http"
	"github.com/absmach/supermq/pkg/messaging"
)

var _ smqhttp.Service = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger *slog.Logger
	svc    smqhttp.Service
}

// NewLogging adds logging facilities to the websocket service.
func NewLogging(svc smqhttp.Service, logger *slog.Logger) smqhttp.Service {
	return &loggingMiddleware{logger, svc}
}

// Subscribe logs the subscribe request. It logs the channel and subtopic(if present) and the time it took to complete the request.
// If the request fails, it logs the error.
func (lm *loggingMiddleware) Subscribe(ctx context.Context, sessionID, username, password, domainID, chanID, subtopic string, topicType messaging.TopicType, c *smqhttp.Client) (err error) {
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
			args = append(args, slog.String("error", err.Error()))
			lm.logger.Warn("Subscribe failed", args...)
			return
		}
		lm.logger.Info("Subscribe completed successfully", args...)
	}(time.Now())

	return lm.svc.Subscribe(ctx, sessionID, username, password, domainID, chanID, subtopic, topicType, c)
}

func (lm *loggingMiddleware) Unsubscribe(ctx context.Context, sessionID, domainID, chanID, subtopic string, topicType messaging.TopicType) (err error) {
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
			args = append(args, slog.String("error", err.Error()))
			lm.logger.Warn("Unsubscribe failed", args...)
			return
		}
		lm.logger.Info("Unsubscribe completed successfully", args...)
	}(time.Now())

	return lm.svc.Unsubscribe(ctx, sessionID, domainID, chanID, subtopic, topicType)
}
