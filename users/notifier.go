// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package users

import "context"

// PasswordResetNotification contains data for password reset notifications.
type PasswordResetNotification struct {
	To    []string
	User  string
	Token string
}

// EmailVerificationNotification contains data for email verification notifications.
type EmailVerificationNotification struct {
	To    []string
	User  string
	Token string
}

// InvitationSentNotification contains data for invitation sent notifications.
type InvitationSentNotification struct {
	To          []string
	InviteeName string
	InviterName string
	DomainName  string
	RoleName    string
}

// InvitationAcceptedNotification contains data for invitation accepted notifications.
type InvitationAcceptedNotification struct {
	To          []string
	InviteeName string
	InviterName string
	DomainName  string
	RoleName    string
}

// InvitationRejectedNotification contains data for invitation rejected notifications.
type InvitationRejectedNotification struct {
	To          []string
	InviteeName string
	InviterName string
	DomainName  string
	RoleName    string
}

// Notifier is an interface for sending notifications through various channels.
type Notifier interface {
	// Notify sends a notification based on the notification type.
	Notify(ctx context.Context, notification any) error
}
