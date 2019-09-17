//
// Copyright (c) 2019
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package twins

import (
	"context"
	"errors"

	"github.com/mainflux/mainflux"
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

	// AddTwin adds new twin to the user identified by the provided key.
	AddTwin(context.Context, string, Twin) (Twin, error)

	// UpdateTwin updates twin identified by the provided Twin that
	// belongs to the user identified by the provided key.
	UpdateTwin(context.Context, string, Twin) error

	// UpdateKey updates key value of the existing twin.
	UpdateKey(context.Context, string, string, string) error

	// ViewTwin retrieves data about twin with the provided
	// ID belonging to the user identified by the provided key.
	ViewTwin(context.Context, string, string) (Twin, error)

	// ListTwins retrieves data about subset of twins that belongs to the
	// user identified by the provided key.
	ListTwins(context.Context, string, uint64, string, Metadata) (TwinsSet, error)

	// ListTwinsByChannel retrieves data about subset of twins that are
	// connected to specified channel and belong to the user identified by
	// the provided key.
	ListTwinsByChannel(context.Context, string, string, uint64) (TwinsSet, error)

	// RemoveTwin removes the twin identified with the provided ID, that
	// belongs to the user identified by the provided key.
	RemoveTwin(context.Context, string, string) error
}

type twinsService struct {
	users  mainflux.UsersServiceClient
	secret string
	twins  TwinRepository
	idp    IdentityProvider
}

var _ Service = (*twinsService)(nil)

// New instantiates the twins service implementation.
func New(secret string, users mainflux.UsersServiceClient, twins TwinRepository, idp IdentityProvider) Service {
	return &twinsService{
		users:  users,
		secret: secret,
		twins:  twins,
		idp:    idp,
	}
}

func (ts *twinsService) Ping(secret string) (string, error) {
	if ts.secret != secret {
		return "", ErrUnauthorizedAccess
	}
	return "Hello World :)", nil
}

func (ts *twinsService) AddThing(ctx context.Context, token string, twin Twin) (Twin, error) {
	res, err := ts.users.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Twin{}, ErrUnauthorizedAccess
	}

	twin.ID, err = ts.idp.ID()
	if err != nil {
		return Twin{}, err
	}

	twin.Owner = res.GetValue()

	if twin.Key == "" {
		twin.Key, err = ts.idp.ID()
		if err != nil {
			return Twin{}, err
		}
	}

	if err := ts.twins.Save(ctx, twin); err != nil {
		return Twin{}, err
	}

	return twin, nil
}
