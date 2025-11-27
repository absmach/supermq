// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package emailer

import (
	"context"
	"errors"
	"fmt"

	"github.com/absmach/supermq/internal/email"
	"github.com/absmach/supermq/users"
)

var (
	errInvalidNotificationType = errors.New("invalid notification type")
	errMissingRecipients       = errors.New("missing recipients")
	errMissingUser             = errors.New("missing user")
	errMissingToken            = errors.New("missing token")
	errMissingInviteeName      = errors.New("missing invitee name")
	errMissingInviterName      = errors.New("missing inviter name")
	errMissingDomainName       = errors.New("missing domain name")
	errMissingRoleName         = errors.New("missing role name")
)

var _ users.Notifier = (*emailer)(nil)

type emailer struct {
	resetURL                string
	verificationURL         string
	resetAgent              *email.Agent
	verifyAgent             *email.Agent
	invitationAgent         *email.Agent
	invitationAcceptedAgent *email.Agent
	invitationRejectedAgent *email.Agent
}

// New creates new email notifier.
func New(resetURL, verificationURL string, resetConfig, verifyConfig, invitationConfig, invitationAcceptedConfig, invitationRejectedConfig *email.Config) (users.Notifier, error) {
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

	invitationRejectedAgent, err := email.New(invitationRejectedConfig)
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
		invitationRejectedAgent: invitationRejectedAgent,
	}, nil
}

// Notify sends a notification via email based on the notification type.
func (e *emailer) Notify(ctx context.Context, notification any) error {
	switch notif := notification.(type) {
	case *users.PasswordResetNotification:
		return e.sendPasswordReset(ctx, notif)
	case *users.EmailVerificationNotification:
		return e.sendEmailVerification(ctx, notif)
	case *users.InvitationNotification:
		return e.sendInvitation(ctx, notif)
	default:
		return fmt.Errorf("%w: %T", errInvalidNotificationType, notification)
	}
}

func (e *emailer) sendPasswordReset(_ context.Context, notif *users.PasswordResetNotification) error {
	if len(notif.To) == 0 {
		return errMissingRecipients
	}
	if notif.User == "" {
		return errMissingUser
	}
	if notif.Token == "" {
		return errMissingToken
	}

	url := fmt.Sprintf("%s?token=%s", e.resetURL, notif.Token)
	return e.resetAgent.Send(notif.To, "", "Password Reset Request", "", notif.User, url, "")
}

func (e *emailer) sendEmailVerification(_ context.Context, notif *users.EmailVerificationNotification) error {
	if len(notif.To) == 0 {
		return errMissingRecipients
	}
	if notif.User == "" {
		return errMissingUser
	}
	if notif.Token == "" {
		return errMissingToken
	}

	url := fmt.Sprintf("%s?token=%s", e.verificationURL, notif.Token)

	return e.verifyAgent.Send(notif.To, "", "Email Verification", "", notif.User, url, "")
}

func (e *emailer) sendInvitation(_ context.Context, notif *users.InvitationNotification) error {
	if len(notif.To) == 0 {
		return errMissingRecipients
	}
	if notif.InviteeName == "" {
		return errMissingInviteeName
	}
	if notif.InviterName == "" {
		return errMissingInviterName
	}
	if notif.DomainName == "" {
		return errMissingDomainName
	}
	if notif.RoleName == "" {
		return errMissingRoleName
	}

	switch notif.Type {
	case users.InvitationSent:
		subject := fmt.Sprintf("You've been invited to join %s", notif.DomainName)
		content := fmt.Sprintf("%s has invited you to join %s as %s.", notif.InviterName, notif.DomainName, notif.RoleName)
		return e.invitationAgent.Send(notif.To, "", subject, "", notif.InviteeName, content, "SuperMQ Team")
	case users.InvitationAccepted:
		subject := fmt.Sprintf("%s accepted your invitation to %s", notif.InviteeName, notif.DomainName)
		content := fmt.Sprintf("%s has accepted your invitation to join %s as %s.", notif.InviteeName, notif.DomainName, notif.RoleName)
		return e.invitationAcceptedAgent.Send(notif.To, "", subject, "", notif.InviterName, content, "SuperMQ Team")
	case users.InvitationRejected:
		subject := fmt.Sprintf("%s declined your invitation to %s", notif.InviteeName, notif.DomainName)
		content := fmt.Sprintf("%s has declined your invitation to join %s as %s.", notif.InviteeName, notif.DomainName, notif.RoleName)
		return e.invitationRejectedAgent.Send(notif.To, "", subject, "", notif.InviterName, content, "SuperMQ Team")
	default:
		return fmt.Errorf("invalid invitation type: %s", notif.Type)
	}
}
