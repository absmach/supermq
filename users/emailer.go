// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package users

import "github.com/absmach/magistrala/pkg/errors"

// Emailer wrapper around the email.
//
//go:generate mockery --name Emailer --output=./mocks --filename emailer.go --quiet --note "Copyright (c) Abstract Machines"
type Emailer interface {
	// SendPasswordReset sends an email to the user with a link to reset the password.
	SendPasswordReset(To []string, host, user, token string) errors.Error
}
