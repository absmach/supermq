// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package auth_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/mainflux/mainflux/auth"
	"github.com/mainflux/mainflux/auth/mocks"
	"github.com/mainflux/mainflux/users"
	"github.com/stretchr/testify/assert"
)

const (
	secret = "secret"
	email  = "test@example.com"
)

var (
	user            = users.User{Email: "user@example.com", Password: "password", Metadata: map[string]interface{}{"role": "user"}}
	nonExistingUser = users.User{Email: "non-ex-user@example.com", Password: "password", Metadata: map[string]interface{}{"role": "user"}}
)

func newService() auth.Service {
	repo := mocks.NewKeyRepository()
	idp := mocks.NewIdentityProvider()
	return auth.New(repo, idp, secret)
}

func TestIssue(t *testing.T) {
	svc := newService()
	loginKey, err := svc.Issue(context.Background(), user.Email, auth.Key{Type: auth.LoginKey, IssuedAt: time.Now()})
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))

	cases := map[string]struct {
		key    auth.Key
		issuer string
		err    error
	}{
		"issue login key": {
			key: auth.Key{
				Type:     auth.LoginKey,
				IssuedAt: time.Now(),
			},
			issuer: email,
			err:    nil,
		},
		"issue login key no issue time": {
			key: auth.Key{
				Type: auth.LoginKey,
			},
			issuer: email,
			err:    auth.ErrInvalidKeyIssuedAt,
		},
		"issue user key": {
			key: auth.Key{
				Type:     auth.UserKey,
				IssuedAt: time.Now(),
			},
			issuer: loginKey.Secret,
			err:    nil,
		},
		"issue user key unauthorized": {
			key: auth.Key{
				Type:     auth.UserKey,
				IssuedAt: time.Now(),
			},
			issuer: "",
			err:    auth.ErrUnauthorizedAccess,
		},
		"issue user key no issue time": {
			key: auth.Key{
				Type: auth.UserKey,
			},
			issuer: loginKey.Secret,
			err:    auth.ErrInvalidKeyIssuedAt,
		},
		"issue reset key": {
			key: auth.Key{
				Type:     auth.ResetKey,
				IssuedAt: time.Now(),
			},
			issuer: loginKey.Secret,
			err:    nil,
		},
		"issue reset key no issue time": {
			key: auth.Key{
				Type: auth.ResetKey,
			},
			issuer: loginKey.Secret,
			err:    auth.ErrInvalidKeyIssuedAt,
		},
	}

	for desc, tc := range cases {
		_, err := svc.Issue(context.Background(), tc.issuer, tc.key)
		assert.Equal(t, err, tc.err, fmt.Sprintf("%s expected %s got %s\n", desc, tc.err, err))
	}
}
func TestRevoke(t *testing.T) {
	svc := newService()
	loginKey, err := svc.Issue(context.Background(), user.Email, auth.Key{Type: auth.LoginKey, IssuedAt: time.Now()})
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))
	key := auth.Key{
		Type:     auth.UserKey,
		IssuedAt: time.Now(),
	}
	newKey, err := svc.Issue(context.Background(), loginKey.Secret, key)
	assert.Nil(t, err, fmt.Sprintf("Issuing users key expected to succeed: %s", err))

	cases := map[string]struct {
		id     string
		issuer string
		err    error
	}{
		"revoke user key": {
			id:     newKey.ID,
			issuer: loginKey.Secret,
			err:    nil,
		},
		"revoke non-existing user key": {
			id:     newKey.ID,
			issuer: loginKey.Secret,
			err:    nil,
		},
		"revoke unauthorized": {
			id:     newKey.ID,
			issuer: "",
			err:    auth.ErrUnauthorizedAccess,
		},
	}

	for desc, tc := range cases {
		err := svc.Revoke(context.Background(), tc.issuer, tc.id)
		assert.Equal(t, err, tc.err, fmt.Sprintf("%s expected %s got %s\n", desc, tc.err, err))
	}
}
func TestRetrieve(t *testing.T) {
	svc := newService()
	loginKey, err := svc.Issue(context.Background(), user.Email, auth.Key{Type: auth.LoginKey, IssuedAt: time.Now()})
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))
	key := auth.Key{
		ID:       "id",
		Type:     auth.UserKey,
		IssuedAt: time.Now(),
	}
	newKey, err := svc.Issue(context.Background(), loginKey.Secret, key)
	assert.Nil(t, err, fmt.Sprintf("Issuing users key expected to succeed: %s", err))

	resetKey, err := svc.Issue(context.Background(), loginKey.Secret, auth.Key{Type: auth.ResetKey, IssuedAt: time.Now()})
	assert.Nil(t, err, fmt.Sprintf("Issuing reset key expected to succeed: %s", err))

	userKey, err := svc.Issue(context.Background(), loginKey.Secret, auth.Key{Type: auth.UserKey, IssuedAt: time.Now()})
	assert.Nil(t, err, fmt.Sprintf("Issuing user key expected to succeed: %s", err))

	cases := map[string]struct {
		id     string
		issuer string
		err    error
	}{
		"retrieve user key": {
			id:     newKey.ID,
			issuer: loginKey.Secret,
			err:    nil,
		},
		"retrieve non-existing user key": {
			id:     "invalid",
			issuer: loginKey.Secret,
			err:    auth.ErrNotFound,
		},
		"retrieve unauthorized": {
			id:     newKey.ID,
			issuer: "wrong",
			err:    auth.ErrUnauthorizedAccess,
		},
		"retrieve with user key": {
			id:     newKey.ID,
			issuer: userKey.Secret,
			err:    auth.ErrUnauthorizedAccess,
		},
		"retrieve with reset key": {
			id:     newKey.ID,
			issuer: resetKey.Secret,
			err:    auth.ErrUnauthorizedAccess,
		},
	}

	for desc, tc := range cases {
		_, err := svc.Retrieve(context.Background(), tc.issuer, tc.id)
		assert.Equal(t, err, tc.err, fmt.Sprintf("%s expected %s got %s\n", desc, tc.err, err))
	}
}
func TestIdentify(t *testing.T) {
	svc := newService()
	loginKey, err := svc.Issue(context.Background(), email, auth.Key{Type: auth.LoginKey, IssuedAt: time.Now()})
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))

	resetKey, err := svc.Issue(context.Background(), loginKey.Secret, auth.Key{Type: auth.ResetKey, IssuedAt: time.Now()})
	assert.Nil(t, err, fmt.Sprintf("Issuing reset key expected to succeed: %s", err))

	exp := time.Now().Add(1 * time.Second)
	userKey, err := svc.Issue(context.Background(), loginKey.Secret, auth.Key{Type: auth.UserKey, IssuedAt: time.Now(), ExpiresAt: &exp})
	assert.Nil(t, err, fmt.Sprintf("Issuing user key expected to succeed: %s", err))

	invalidKey, err := svc.Issue(context.Background(), loginKey.Secret, auth.Key{Type: 22, IssuedAt: time.Now()})
	assert.Nil(t, err, fmt.Sprintf("Issuing user key expected to succeed: %s", err))

	cases := []struct {
		desc string
		key  string
		id   string
		err  error
	}{
		{
			desc: "identify login key",
			key:  loginKey.Secret,
			id:   email,
			err:  nil,
		},
		{
			desc: "identify reset key",
			key:  resetKey.Secret,
			id:   "mainflux.auth",
			err:  nil,
		},
		{
			desc: "identify user key",
			key:  userKey.Secret,
			id:   email,
			err:  nil,
		},
		{
			desc: "identify expired user key",
			key:  userKey.Secret,
			id:   "",
			err:  auth.ErrUnauthorizedAccess,
		},
		{
			desc: "identify expired key",
			key:  invalidKey.Secret,
			id:   "",
			err:  auth.ErrUnauthorizedAccess,
		},
		{
			desc: "identify invalid key",
			key:  "invalid",
			id:   "",
			err:  auth.ErrUnauthorizedAccess,
		},
	}

	for i, tc := range cases {
		id, err := svc.Identify(context.Background(), tc.key)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s expected %s got %s\n", tc.desc, tc.err, err))
		assert.Equal(t, tc.id, id, fmt.Sprintf("%s expected %s got %s\n", tc.desc, tc.id, id))
		if i == 2 {
			// Wait for key to expire.
			time.Sleep(2 * time.Second)
		}
	}
}
