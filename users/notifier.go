// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package users

import "context"

// NotificationType represents the type of notification to send.
type NotificationType string

const (
	// NotificationPasswordReset is sent when a user requests a password reset.
	NotificationPasswordReset NotificationType = "password_reset"
	// NotificationEmailVerification is sent when a user needs to verify their email.
	NotificationEmailVerification NotificationType = "email_verification"
	// NotificationInvitationSent is sent when a user is invited to a domain.
	NotificationInvitationSent NotificationType = "invitation_sent"
	// NotificationInvitationAccepted is sent when a user accepts an invitation.
	NotificationInvitationAccepted NotificationType = "invitation_accepted"
)

// NotificationData contains the data needed to send a notification.
type NotificationData struct {
	// Type is the type of notification to send.
	Type NotificationType

	// Recipients is the list of recipients for the notification.
	Recipients []string

	// Metadata contains additional data specific to the notification type.
	Metadata map[string]string
}

// Notifier is an interface for sending notifications through various channels.
type Notifier interface {
	// Notify sends a notification of the given type to the specified recipients.
	Notify(ctx context.Context, data NotificationData) error
}
