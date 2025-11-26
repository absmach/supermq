// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package emailer

import (
	"fmt"

	"github.com/absmach/supermq/internal/email"
	"github.com/absmach/supermq/users"
)

var _ users.Emailer = (*emailer)(nil)

type emailer struct {
	resetURL                string
	verificationURL         string
	resetAgent              *email.Agent
	verifyAgent             *email.Agent
	invitationAgent         *email.Agent
	invitationAcceptedAgent *email.Agent
}

// New creates new emailer utility.
func New(resetURL, verificationURL string, resetConfig, verifyConfig, invitationConfig, invitationAcceptedConfig *email.Config) (users.Emailer, error) {
	resetAgent, err := email.New(resetConfig)
	if err != nil {
		return nil, err
	}

	verifyAgent, err := email.New(verifyConfig)
	if err != nil {
		return nil, err
	}

	invitationAgent, err := email.New(invitationConfig)
	if err != nil {
		return nil, err
	}

	invitationAcceptedAgent, err := email.New(invitationAcceptedConfig)
	if err != nil {
		return nil, err
	}

	return &emailer{
		resetURL:                resetURL,
		verificationURL:         verificationURL,
		resetAgent:              resetAgent,
		verifyAgent:             verifyAgent,
		invitationAgent:         invitationAgent,
		invitationAcceptedAgent: invitationAcceptedAgent,
	}, nil
}

func (e *emailer) SendPasswordReset(to []string, user, token string) error {
	url := fmt.Sprintf("%s?token=%s", e.resetURL, token)
	return e.resetAgent.Send(to, "", "Password Reset Request", "", user, url, "")
}

func (e *emailer) SendVerification(to []string, user, verificationToken string) error {
	url := fmt.Sprintf("%s?token=%s", e.verificationURL, verificationToken)
	return e.verifyAgent.Send(to, "", "Email Verification", "", user, url, "")
}

func (e *emailer) SendInvitation(to []string, inviteeName, inviterName, domainName, roleName string) error {
	subject := fmt.Sprintf("You've been invited to join %s", domainName)
	content := fmt.Sprintf("%s has invited you to join %s as %s.", inviterName, domainName, roleName)
	return e.invitationAgent.Send(to, "", subject, "", inviteeName, content, "SuperMQ Team")
}

func (e *emailer) SendInvitationAccepted(to []string, inviterName, inviteeName, domainName, roleName string) error {
	subject := fmt.Sprintf("%s accepted your invitation to %s", inviteeName, domainName)
	content := fmt.Sprintf("%s has accepted your invitation to join %s as %s.", inviteeName, domainName, roleName)
	return e.invitationAcceptedAgent.Send(to, "", subject, "", inviterName, content, "SuperMQ Team")
}
