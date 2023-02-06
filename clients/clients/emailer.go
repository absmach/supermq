// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package clients

// Emailer wrapper around the email
type Emailer interface {
	SendPasswordReset(To []string, host, token string) error
}
