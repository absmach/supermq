// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package notifications

import (
	"context"
)

// NotificationType represents the type of notification.
type NotificationType uint8

const (
	// InvitationSent represents a notification sent when an invitation is sent.
	InvitationSent NotificationType = iota
	// InvitationAccepted represents a notification sent when an invitation is accepted.
	InvitationAccepted
)

// String representation of the notification type.
const (
	invitationSentStr     = "invitation_sent"
	invitationAcceptedStr = "invitation_accepted"
)

// String converts notification type to string literal.
func (n NotificationType) String() string {
	switch n {
	case InvitationSent:
		return invitationSentStr
	case InvitationAccepted:
		return invitationAcceptedStr
	default:
		return ""
	}
}

// Notification represents a notification to be sent.
type Notification struct {
	Type            NotificationType
	InviteeUserID   string
	InvitedBy       string
	DomainID        string
	DomainName      string
	RoleID          string
	RoleName        string
	InviteeEmail    string
	InviterEmail    string
	InviteeUsername string
	InviterUsername string
}

// Service provides access to the notifications service.
// The service acts as a consumer that subscribes to events, converts them to Notifications,
// and routes them to all registered notifiers.
type Service interface {
	// SendNotification sends a notification.
	SendNotification(ctx context.Context, notification Notification) error
}

// Notifier is a generic interface for sending notifications through various channels.
// Implementations can be for email, SMS, Slack, push notifications, etc.
// Multiple notifiers can be registered with the notifications service to enable
// multi-channel notification delivery.
type Notifier interface {
	// Notify sends a notification through the specific channel.
	// The implementation determines how to extract and format the necessary
	// information from the Notification struct.
	Notify(ctx context.Context, notification Notification) error
}
