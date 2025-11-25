// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package emailer

import (
	"fmt"

	"github.com/absmach/supermq/internal/email"
	"github.com/absmach/supermq/notifications"
)

var _ notifications.Emailer = (*emailer)(nil)

type emailer struct {
	invitationSentAgent     *email.Agent
	invitationAcceptedAgent *email.Agent
}

// New creates a new notifications emailer.
func New(invitationSentConfig, invitationAcceptedConfig *email.Config) (notifications.Emailer, error) {
	invitationSentAgent, err := email.New(invitationSentConfig)
	if err != nil {
		return nil, err
	}

	invitationAcceptedAgent, err := email.New(invitationAcceptedConfig)
	if err != nil {
		return nil, err
	}

	return &emailer{
		invitationSentAgent:     invitationSentAgent,
		invitationAcceptedAgent: invitationAcceptedAgent,
	}, nil
}

func (e *emailer) SendInvitationSentEmail(to []string, inviteeName, domainName, inviterName string) error {
	subject := fmt.Sprintf("You've been invited to join %s", domainName)
	header := "Domain Invitation"
	content := fmt.Sprintf("%s has invited you to join the domain %s. Please log in to accept or reject this invitation.", inviterName, domainName)
	footer := "SuperMQ Team"

	return e.invitationSentAgent.Send(to, "", subject, header, inviteeName, content, footer)
}

func (e *emailer) SendInvitationAcceptedEmail(to []string, inviterName, inviteeName, domainName string) error {
	subject := fmt.Sprintf("%s accepted your invitation", inviteeName)
	header := "Invitation Accepted"
	content := fmt.Sprintf("%s has accepted your invitation to join the domain %s.", inviteeName, domainName)
	footer := "SuperMQ Team"

	return e.invitationAcceptedAgent.Send(to, "", subject, header, inviterName, content, footer)
}
