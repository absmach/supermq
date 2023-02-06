// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0
package emailer

import (
	"fmt"

	"github.com/mainflux/mainflux/clients/clients"
	"github.com/mainflux/mainflux/internal/email"
)

var _ clients.Emailer = (*emailer)(nil)

type emailer struct {
	resetURL string
	agent    *email.Agent
}

// New creates new emailer utility
func New(url string, c *email.Config) (clients.Emailer, error) {
	e, err := email.New(c)
	return &emailer{resetURL: url, agent: e}, err
}

func (e *emailer) SendPasswordReset(To []string, host string, token string) error {
	url := fmt.Sprintf("%s%s?token=%s", host, e.resetURL, token)
	return e.agent.Send(To, "", "Password reset", "", url, "")
}
