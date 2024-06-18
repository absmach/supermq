// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package emailer

import (
	"fmt"

	"github.com/absmach/magistrala/internal/email"
	"github.com/absmach/magistrala/pkg/errors"
	"github.com/absmach/magistrala/users"
)

var _ users.Emailer = (*emailer)(nil)

type emailer struct {
	resetURL string
	agent    *email.Agent
}

// New creates new emailer utility.
func New(url string, c *email.Config) (users.Emailer, errors.Error) {
	e, err := email.New(c)
	return &emailer{resetURL: url, agent: e}, errors.Cast(err)
}

func (e *emailer) SendPasswordReset(to []string, host, user, token string) errors.Error {
	url := fmt.Sprintf("%s%s?token=%s", host, e.resetURL, token)
	return errors.Cast(e.agent.Send(to, "", "Password Reset Request", "", user, url, ""))
}
