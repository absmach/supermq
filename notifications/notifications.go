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
type Service interface {
	// SendNotification sends a notification.
	SendNotification(ctx context.Context, notification Notification) error
}

// Emailer sends notification emails.
type Emailer interface {
	// SendInvitationSentEmail sends an email to the invitee when they receive an invitation.
	SendInvitationSentEmail(to []string, inviteeName, domainName, inviterName string) error

	// SendInvitationAcceptedEmail sends an email to the inviter when their invitation is accepted.
	SendInvitationAcceptedEmail(to []string, inviterName, inviteeName, domainName string) error
}
