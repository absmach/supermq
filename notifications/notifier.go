// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package notifications

import (
	"context"
)

// NotificationType represents the type of notification to send.
type NotificationType uint8

const (
	// Invitation represents an invitation notification.
	Invitation NotificationType = iota
	// Acceptance represents an acceptance notification.
	Acceptance
	// Rejection represents a rejection notification.
	Rejection
)

// Notification contains the data needed to send a notification.
type Notification struct {
	Type       NotificationType
	InviterID  string
	InviteeID  string
	DomainID   string
	DomainName string
	RoleID     string
	RoleName   string
}

// Notifier represents a service for sending notifications.
type Notifier interface {
	// Notify sends a notification based on the provided notification data.
	Notify(ctx context.Context, n Notification) error
}
