// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package users

// Emailer wrapper around the email.
type Emailer interface {
	// SendPasswordReset sends an email to the user with a link to reset the password.
	SendPasswordReset(To []string, user, token string) error

	// SendVerification sends an email to the user with a verification token.
	SendVerification(To []string, user, verificationToken string) error
}
