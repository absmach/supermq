// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package notifications

import (
	"context"
)

// Notifier represents a service for sending notifications.
type Notifier interface {
	// SendInvitationNotification sends a notification when a user is invited to a domain.
	SendInvitationNotification(ctx context.Context, inviterID, inviteeID, domainID, domainName, roleID, roleName string) error

	// SendAcceptanceNotification sends a notification when a user accepts a domain invitation.
	SendAcceptanceNotification(ctx context.Context, inviterID, inviteeID, domainID, domainName, roleID, roleName string) error

	// SendRejectionNotification sends a notification when a user rejects a domain invitation.
	SendRejectionNotification(ctx context.Context, inviterID, inviteeID, domainID, domainName, roleID, roleName string) error
}
