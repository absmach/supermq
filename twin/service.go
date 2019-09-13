//
// Copyright (c) 2019
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package twin

import (
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
)

var (
	// ErrMalformedEntity indicates malformed entity specification (e.g.
	// invalid username or password).
	ErrMalformedEntity = errors.New("malformed entity specification")

	// ErrUnauthorizedAccess indicates missing or invalid credentials provided
	// when accessing a protected resource.
	ErrUnauthorizedAccess = errors.New("missing or invalid credentials provided")
)

// Service specifies an API that must be fullfiled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	// Ping compares a given string with secret
	Ping(string) (string, error)
}

type twinService struct {
	secret string
	db     *mongo.Database
}

var _ Service = (*twinService)(nil)

// New instantiates the twin service implementation.
func New(secret string, db *mongo.Database) Service {
	return &twinService{
		secret: secret,
		db:     db,
	}
}

func (ks *twinService) Ping(secret string) (string, error) {
	if ks.secret != secret {
		return "", ErrUnauthorizedAccess
	}
	return "Hello World :)", nil
}
