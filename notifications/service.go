// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package notifications

import (
	"context"
	"log/slog"
)

var _ Service = (*service)(nil)

type service struct {
	emailer Emailer
	logger  *slog.Logger
}

// NewService returns a new notifications service.
func NewService(emailer Emailer, logger *slog.Logger) Service {
	return &service{
		emailer: emailer,
		logger:  logger,
	}
}

func (s *service) SendNotification(ctx context.Context, notification Notification) error {
	switch notification.Type {
	case InvitationSent:
		return s.sendInvitationSent(ctx, notification)
	case InvitationAccepted:
		return s.sendInvitationAccepted(ctx, notification)
	default:
		s.logger.Warn("unknown notification type", slog.String("type", notification.Type.String()))
		return nil
	}
}

func (s *service) sendInvitationSent(ctx context.Context, notification Notification) error {
	if notification.InviteeEmail == "" {
		s.logger.Warn("invitee email is empty, skipping notification")
		return nil
	}

	inviteeName := notification.InviteeUsername
	if inviteeName == "" {
		inviteeName = notification.InviteeEmail
	}

	inviterName := notification.InviterUsername
	if inviterName == "" {
		inviterName = notification.InviterEmail
	}
	if inviterName == "" {
		inviterName = "A user"
	}

	domainName := notification.DomainName
	if domainName == "" {
		domainName = notification.DomainID
	}

	s.logger.Info("sending invitation sent notification",
		slog.String("to", notification.InviteeEmail),
		slog.String("domain", domainName),
	)

	return s.emailer.SendInvitationSentEmail(
		[]string{notification.InviteeEmail},
		inviteeName,
		domainName,
		inviterName,
	)
}

func (s *service) sendInvitationAccepted(ctx context.Context, notification Notification) error {
	if notification.InviterEmail == "" {
		s.logger.Warn("inviter email is empty, skipping notification")
		return nil
	}

	inviteeName := notification.InviteeUsername
	if inviteeName == "" {
		inviteeName = notification.InviteeEmail
	}
	if inviteeName == "" {
		inviteeName = "A user"
	}

	inviterName := notification.InviterUsername
	if inviterName == "" {
		inviterName = notification.InviterEmail
	}

	domainName := notification.DomainName
	if domainName == "" {
		domainName = notification.DomainID
	}

	s.logger.Info("sending invitation accepted notification",
		slog.String("to", notification.InviterEmail),
		slog.String("domain", domainName),
	)

	return s.emailer.SendInvitationAcceptedEmail(
		[]string{notification.InviterEmail},
		inviterName,
		inviteeName,
		domainName,
	)
}
