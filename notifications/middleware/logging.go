// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/absmach/supermq/notifications"
)

var _ notifications.Notifier = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger   *slog.Logger
	notifier notifications.Notifier
}

// NewLogging adds logging facilities to the notifier.
func NewLogging(notifier notifications.Notifier, logger *slog.Logger) notifications.Notifier {
	return &loggingMiddleware{
		logger:   logger,
		notifier: notifier,
	}
}

func (lm *loggingMiddleware) SendInvitationNotification(ctx context.Context, inviterID, inviteeID, domainID, domainName, roleID, roleName string) (err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.Group("invitation",
				slog.String("inviter_id", inviterID),
				slog.String("invitee_id", inviteeID),
				slog.String("domain_id", domainID),
				slog.String("domain_name", domainName),
				slog.String("role_id", roleID),
				slog.String("role_name", roleName),
			),
		}
		if err != nil {
			args = append(args, slog.String("error", err.Error()))
			lm.logger.Warn("Send invitation notification failed", args...)
			return
		}
		lm.logger.Info("Send invitation notification completed successfully", args...)
	}(time.Now())

	return lm.notifier.SendInvitationNotification(ctx, inviterID, inviteeID, domainID, domainName, roleID, roleName)
}

func (lm *loggingMiddleware) SendAcceptanceNotification(ctx context.Context, inviterID, inviteeID, domainID, domainName, roleID, roleName string) (err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.Group("acceptance",
				slog.String("inviter_id", inviterID),
				slog.String("invitee_id", inviteeID),
				slog.String("domain_id", domainID),
				slog.String("domain_name", domainName),
				slog.String("role_id", roleID),
				slog.String("role_name", roleName),
			),
		}
		if err != nil {
			args = append(args, slog.String("error", err.Error()))
			lm.logger.Warn("Send acceptance notification failed", args...)
			return
		}
		lm.logger.Info("Send acceptance notification completed successfully", args...)
	}(time.Now())

	return lm.notifier.SendAcceptanceNotification(ctx, inviterID, inviteeID, domainID, domainName, roleID, roleName)
}

func (lm *loggingMiddleware) SendRejectionNotification(ctx context.Context, inviterID, inviteeID, domainID, domainName, roleID, roleName string) (err error) {
	defer func(begin time.Time) {
		args := []any{
			slog.String("duration", time.Since(begin).String()),
			slog.Group("rejection",
				slog.String("inviter_id", inviterID),
				slog.String("invitee_id", inviteeID),
				slog.String("domain_id", domainID),
				slog.String("domain_name", domainName),
				slog.String("role_id", roleID),
				slog.String("role_name", roleName),
			),
		}
		if err != nil {
			args = append(args, slog.String("error", err.Error()))
			lm.logger.Warn("Send rejection notification failed", args...)
			return
		}
		lm.logger.Info("Send rejection notification completed successfully", args...)
	}(time.Now())

	return lm.notifier.SendRejectionNotification(ctx, inviterID, inviteeID, domainID, domainName, roleID, roleName)
}
