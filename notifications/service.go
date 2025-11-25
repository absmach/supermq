// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package notifications

import (
	"context"
	"log/slog"
)

var _ Service = (*service)(nil)

// Config holds configuration for the notifications service.
type Config struct {
	// InvitationSentTemplate is the template path for invitation sent emails
	InvitationSentTemplate string
	// InvitationAcceptedTemplate is the template path for invitation accepted emails
	InvitationAcceptedTemplate string
	// DefaultSubjectPrefix is the prefix for email subjects (e.g., "SuperMQ" or "PRISM")
	DefaultSubjectPrefix string
	// DefaultFooter is the footer text for emails
	DefaultFooter string
}

type service struct {
	notifiers []Notifier
	config    Config
	logger    *slog.Logger
}

// NewService returns a new notifications service.
func NewService(notifiers []Notifier, config Config, logger *slog.Logger) Service {
	return &service{
		notifiers: notifiers,
		config:    config,
		logger:    logger,
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

	// Normalize domain name for logging
	domainName := notification.DomainName
	if domainName == "" {
		domainName = notification.DomainID
	}

	s.logger.Info("sending invitation sent notification",
		slog.String("to", notification.InviteeEmail),
		slog.String("domain", domainName),
		slog.Int("notifiers", len(s.notifiers)),
	)

	// Route notification to all registered notifiers
	for _, notifier := range s.notifiers {
		if err := notifier.Notify(ctx, notification); err != nil {
			s.logger.Error("failed to send notification via notifier",
				slog.String("error", err.Error()),
				slog.String("to", notification.InviteeEmail),
			)
			// Continue to other notifiers if one fails
		}
	}

	return nil
}

func (s *service) sendInvitationAccepted(ctx context.Context, notification Notification) error {
	if notification.InviterEmail == "" {
		s.logger.Warn("inviter email is empty, skipping notification")
		return nil
	}

	// Normalize domain name for logging
	domainName := notification.DomainName
	if domainName == "" {
		domainName = notification.DomainID
	}

	s.logger.Info("sending invitation accepted notification",
		slog.String("to", notification.InviterEmail),
		slog.String("domain", domainName),
		slog.Int("notifiers", len(s.notifiers)),
	)

	// Route notification to all registered notifiers
	for _, notifier := range s.notifiers {
		if err := notifier.Notify(ctx, notification); err != nil {
			s.logger.Error("failed to send notification via notifier",
				slog.String("error", err.Error()),
				slog.String("to", notification.InviterEmail),
			)
			// Continue to other notifiers even if one fails
		}
	}

	return nil
}
