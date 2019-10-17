// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0
package emailer

import (
	"fmt"

	"github.com/mainflux/mainflux/internal/email"
	"github.com/mainflux/mainflux/users"
)

const (
	message = `
You have initiated password reset.
Follow the link below to reset password.`
)

var _ users.Emailer = (*emailer)(nil)

type emailer struct {
	resetURL string
	agent    *email.Agent
}

// New creates new emailer utility
func New(url string, c *email.Config) (users.Emailer, error) {
	e, err := email.New(c, nil)
	if err != nil {
		return nil, err
	}
	return &emailer{resetURL: url, agent: e}, nil
}

func (e *emailer) SendPasswordReset(To []string, host string, token string) error {
	url := fmt.Sprintf("%s%s?token=%s", host, e.resetURL, token)
	content := fmt.Sprintf("%s\r\n%s\r\n", message, url)
	return e.agent.Send(To, "", "Password reset", "", content, "")
}
