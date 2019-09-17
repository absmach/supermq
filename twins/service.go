//
// Copyright (c) 2019
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package twins

import (
	"errors"
)

var (
	// ErrMalformedEntity indicates malformed entity specification (e.g.
	// invalid username or password).
	ErrMalformedEntity = errors.New("malformed entity specification")

	// ErrUnauthorizedAccess indicates missing or invalid credentials provided
	// when accessing a protected resource.
	ErrUnauthorizedAccess = errors.New("missing or invalid credentials provided")

	// ErrNotFound indicates a non-existent entity request.
	ErrNotFound = errors.New("non-existent entity")

	// ErrConflict indicates that entity already exists.
	ErrConflict = errors.New("entity already exists")
)

// Service specifies an API that must be fullfiled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	// Ping compares a given string with secret
	Ping(string) (string, error)
}

type twinsService struct {
	secret string
	twins  TwinRepository
}

var _ Service = (*twinsService)(nil)

// New instantiates the twins service implementation.
func New(secret string, twins TwinRepository) Service {
	return &twinsService{
		secret: secret,
		twins:  twins,
	}
}

func (ks *twinsService) Ping(secret string) (string, error) {
	if ks.secret != secret {
		return "", ErrUnauthorizedAccess
	}
	return "Hello World :)", nil
}
