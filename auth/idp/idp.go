// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// Package idp provides a JWT identity provider.
package idp

import (
	"github.com/gofrs/uuid"
	"github.com/mainflux/mainflux/auth"
)

var _ auth.IdentityProvider = (*authIdentityProvider)(nil)

type authIdentityProvider struct {
}

// New instantiates a Identity Provider.
func New() auth.IdentityProvider {
	return &authIdentityProvider{}
}

func (idp *authIdentityProvider) ID() (string, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	return id.String(), nil
}
