// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package emailer

import (
	"context"
	"fmt"

	"github.com/absmach/supermq/users"
)

var _ users.Notifier = (*emailNotifier)(nil)

type emailNotifier struct {
	emailer users.Emailer
}

// NewNotifier creates a new email notifier that uses the provided emailer.
func NewNotifier(emailer users.Emailer) users.Notifier {
	return &emailNotifier{
		emailer: emailer,
	}
}

// Notify sends a notification via email based on the notification type.
func (n *emailNotifier) Notify(ctx context.Context, data users.NotificationData) error {
	switch data.Type {
	case users.NotificationInvitationSent:
		return n.notifyInvitationSent(data)
	case users.NotificationInvitationAccepted:
		return n.notifyInvitationAccepted(data)
	default:
		return fmt.Errorf("unknown notification type: %s", data.Type)
	}
}

func (n *emailNotifier) notifyInvitationSent(data users.NotificationData) error {
	inviteeName := data.Metadata["invitee_name"]
	inviterName := data.Metadata["inviter_name"]
	domainName := data.Metadata["domain_name"]
	roleName := data.Metadata["role_name"]

	return n.emailer.SendInvitation(data.Recipients, inviteeName, inviterName, domainName, roleName)
}

func (n *emailNotifier) notifyInvitationAccepted(data users.NotificationData) error {
	inviteeName := data.Metadata["invitee_name"]
	inviterName := data.Metadata["inviter_name"]
	domainName := data.Metadata["domain_name"]
	roleName := data.Metadata["role_name"]

	return n.emailer.SendInvitationAccepted(data.Recipients, inviterName, inviteeName, domainName, roleName)
}
