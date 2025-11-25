// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package emailer

import (
	"context"
	"fmt"

	"github.com/absmach/supermq/internal/email"
	"github.com/absmach/supermq/notifications"
)

var _ notifications.Notifier = (*emailer)(nil)

type emailer struct {
	agent  *email.Agent
	config Config
}

// Config holds email-specific configuration.
type Config struct {
	// Footer is the footer text for emails
	Footer string
	// SubjectPrefix is the prefix for email subjects (e.g., "SuperMQ" or "PRISM")
	SubjectPrefix string
}

// New creates a new email notifier that implements the Notifier interface.
func New(emailConfig *email.Config, config Config) (notifications.Notifier, error) {
	agent, err := email.New(emailConfig)
	if err != nil {
		return nil, err
	}

	return &emailer{
		agent:  agent,
		config: config,
	}, nil
}

// Notify sends an email notification based on the notification type.
// It extracts the necessary information from the Notification struct and
// formats it appropriately for email delivery.
func (e *emailer) Notify(ctx context.Context, notification notifications.Notification) error {
	switch notification.Type {
	case notifications.InvitationSent:
		return e.sendInvitationSentEmail(notification)
	case notifications.InvitationAccepted:
		return e.sendInvitationAcceptedEmail(notification)
	default:
		return fmt.Errorf("unsupported notification type: %s", notification.Type.String())
	}
}

func (e *emailer) sendInvitationSentEmail(notification notifications.Notification) error {
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

	subject := fmt.Sprintf("You've been invited to join %s", domainName)
	if e.config.SubjectPrefix != "" {
		subject = fmt.Sprintf("[%s] %s", e.config.SubjectPrefix, subject)
	}

	header := "Domain Invitation"
	content := fmt.Sprintf("%s has invited you to join the domain %s. Please log in to accept or reject this invitation.", inviterName, domainName)
	footer := e.getFooter()

	return e.agent.Send([]string{notification.InviteeEmail}, "", subject, header, inviteeName, content, footer)
}

func (e *emailer) sendInvitationAcceptedEmail(notification notifications.Notification) error {
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

	subject := fmt.Sprintf("%s accepted your invitation", inviteeName)
	if e.config.SubjectPrefix != "" {
		subject = fmt.Sprintf("[%s] %s", e.config.SubjectPrefix, subject)
	}

	header := "Invitation Accepted"
	content := fmt.Sprintf("%s has accepted your invitation to join the domain %s.", inviteeName, domainName)
	footer := e.getFooter()

	return e.agent.Send([]string{notification.InviterEmail}, "", subject, header, inviterName, content, footer)
}

func (e *emailer) getFooter() string {
	if e.config.Footer != "" {
		return e.config.Footer
	}
	return "Notification Service"
}
