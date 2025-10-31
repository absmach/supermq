// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package users

// Emailer wrapper around the email.
type Emailer interface {
	// SendPasswordReset sends an email to the user with a link to reset the password.
	SendPasswordReset(To []string, user, token string) error

	// SendVerification sends an email to the user with a verification token.
	SendVerification(To []string, user, verificationToken string) error

	// Send sends an email with custom parameters.
	Send(To []string, from, subject, header, user, content, footer string) error

	// SendCustom sends an email with custom parameters using a custom email agent.
	SendCustom(To []string, from, subject, header, user, content, footer string) error
}
