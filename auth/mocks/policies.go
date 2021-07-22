// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"

	"github.com/mainflux/mainflux/auth"
)

type ketoMock struct{}

// NewKetoMock returns a mock service for Keto.
// This mock is not implemented yet.
func NewKetoMock() auth.PolicyCommunicator {
	return ketoMock{}
}

func (k ketoMock) CheckPolicy(ctx context.Context, subject, object, relation string) (auth.PolicyResult, error) {
	// Not implemented yet.
	return auth.PolicyResult{Authorized: true}, nil
}

func (k ketoMock) AddPolicy(ctx context.Context, subject, object, relation string) error {
	// Not implemented yet.
	return nil
}
