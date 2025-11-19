// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/absmach/supermq/internal/testsutil"
	"github.com/absmach/supermq/pkg/authn"
	"github.com/absmach/supermq/pkg/errors"
	"github.com/absmach/supermq/users"
	"github.com/absmach/supermq/users/events"
	"github.com/absmach/supermq/users/mocks"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	validSession = authn.Session{
		UserID: testsutil.GenerateUUID(&testing.T{}),
	}
	validUser = users.User{
		ID:        testsutil.GenerateUUID(&testing.T{}),
		FirstName: "Test",
		LastName:  "User",
		Email:     "testuser@example.com",
	}
)

func newEventStoreMiddleware(t *testing.T) (*mocks.Service, users.Service) {
	svc := new(mocks.Service)
	nsvc, err := events.NewEventStoreMiddleware(context.Background(), svc, storeURL)
	require.Nil(t, err, fmt.Sprintf("create events store middleware failed with unexpected error: %s", err))

	return svc, nsvc
}

func TestRegister(t *testing.T) {
	_, nsvc := newEventStoreMiddleware(t)

	validID := testsutil.GenerateUUID(t)
	validCtx := context.WithValue(context.Background(), middleware.RequestIDKey, validID)

	cases := []struct {
		desc         string
		session      authn.Session
		user         users.User
		selfRegister bool
		svcRes       users.User
		svcErr       error
		err          error
	}{
		{
			desc:         "register user successfully",
			session:      validSession,
			user:         validUser,
			selfRegister: true,
			svcRes:       validUser,
			svcErr:       nil,
			err:          nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			_, err := nsvc.Register(validCtx, tc.session, tc.user, tc.selfRegister)
			assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		})
	}
}
