// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package users

// Emailer wrapper around the email.
type Emailer interface {
	// SendPasswordReset sends an email to the user with a link to reset the password.
	SendPasswordReset(To []string, user, token string) error

	// SendVerification sends an email to the user with a verification token.
	SendVerification(To []string, user, verificationToken string) error

	// SendInvitation sends an email to the invitee when they are invited to a domain.
	SendInvitation(To []string, inviteeName, inviterName, domainName, roleName string) error

	// SendInvitationAccepted sends an email to the inviter when their invitation is accepted.
	SendInvitationAccepted(To []string, inviterName, inviteeName, domainName, roleName string) error
}
