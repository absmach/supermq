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

// InvitationType represents the type of invitation notification.
type InvitationType string

const (
	// InvitationSent indicates an invitation was sent.
	InvitationSent InvitationType = "sent"
	// InvitationAccepted indicates an invitation was accepted.
	InvitationAccepted InvitationType = "accepted"
	// InvitationRejected indicates an invitation was rejected.
	InvitationRejected InvitationType = "rejected"
)

// InvitationNotification contains data for invitation notifications.
type InvitationNotification struct {
	Type        InvitationType
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
