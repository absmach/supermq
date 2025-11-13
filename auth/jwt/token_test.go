// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package jwt_test

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"testing"
	"time"

	"github.com/absmach/supermq/auth"
	authjwt "github.com/absmach/supermq/auth/jwt"
	"github.com/absmach/supermq/auth/mocks"
	"github.com/absmach/supermq/internal/testsutil"
	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	tokenType     = "type"
	roleField     = "role"
	VerifiedField = "verified"
	issuerName    = "supermq.auth"
	secret        = "test"
)

var (
	errInvalidIssuer = errors.New("invalid token issuer value")
	errJWTExpiryKey  = errors.New(`"exp" not satisfied`)
	keyManager       = new(mocks.KeyManager)
)

func newToken(issuerName string, key auth.Key) jwt.Token {
	builder := jwt.NewBuilder()
	builder.
		Issuer(issuerName).
		IssuedAt(key.IssuedAt).
		Claim(tokenType, key.Type).
		Expiration(key.ExpiresAt)
	builder.Claim(roleField, key.Role)
	builder.Claim(VerifiedField, key.Verified)
	if key.Subject != "" {
		builder.Subject(key.Subject)
	}
	if key.ID != "" {
		builder.JwtID(key.ID)
	}
	tkn, _ := builder.Build()
	return tkn
}

func TestIssue(t *testing.T) {
	tokenizer := authjwt.New(keyManager)

	signedToken, _, err := signToken(issuerName, key())
	require.Nil(t, err, fmt.Sprintf("issuing key expected to succeed: %s", err))

	cases := []struct {
		desc        string
		key         auth.Key
		managerReq  jwt.Token
		managerResp []byte
		managerErr  error
		err         error
	}{
		{
			desc:        "issue new token",
			key:         key(),
			managerResp: signedToken,
			err:         nil,
		},
		{
			desc: "issue token with OAuth token",
			key: auth.Key{
				ID:        testsutil.GenerateUUID(t),
				Type:      auth.AccessKey,
				Subject:   testsutil.GenerateUUID(t),
				IssuedAt:  time.Now().Add(-10 * time.Second).Round(time.Second),
				ExpiresAt: time.Now().Add(10 * time.Minute).Round(time.Second),
			},
			err: nil,
		},
		{
			desc: "issue token without a domain",
			key: auth.Key{
				ID:       testsutil.GenerateUUID(t),
				Type:     auth.AccessKey,
				Subject:  testsutil.GenerateUUID(t),
				IssuedAt: time.Now().Add(-10 * time.Second).Round(time.Second),
			},
			err: nil,
		},
		{
			desc: "issue token without a subject",
			key: auth.Key{
				ID:       testsutil.GenerateUUID(t),
				Type:     auth.AccessKey,
				Subject:  "",
				IssuedAt: time.Now().Add(-10 * time.Second).Round(time.Second),
			},
			err: nil,
		},
		{
			desc: "issue token without type",
			key: auth.Key{
				ID:       testsutil.GenerateUUID(t),
				Type:     auth.KeyType(auth.InvitationKey + 1),
				Subject:  testsutil.GenerateUUID(t),
				IssuedAt: time.Now().Add(-10 * time.Second).Round(time.Second),
			},
			err: nil,
		},
		{
			desc: "issue token without a domain and subject",
			key: auth.Key{
				ID:        testsutil.GenerateUUID(t),
				Type:      auth.AccessKey,
				Subject:   "",
				IssuedAt:  time.Now().Add(-10 * time.Second).Round(time.Second),
				ExpiresAt: time.Now().Add(10 * time.Minute).Round(time.Second),
			},
			err: nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			tc.managerReq = newToken(issuerName, tc.key)
			kmCall := keyManager.On("SignJWT", tc.managerReq).Return(tc.managerResp, tc.managerErr)
			tkn, err := tokenizer.Issue(tc.key)
			assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s expected %s, got %s", tc.desc, tc.err, err))
			if err != nil {
				assert.NotEmpty(t, tkn, fmt.Sprintf("%s expected token, got empty string", tc.desc))
			}
			kmCall.Unset()
		})
	}
}

func TestParse(t *testing.T) {
	tokenizer := authjwt.New(keyManager)

	signedTkn, parsedTkn, err := signToken(issuerName, key())
	require.Nil(t, err, fmt.Sprintf("issuing key expected to succeed: %s", err))

	apiKey := key()
	apiKey.Type = auth.APIKey
	apiKey.ExpiresAt = time.Now().UTC().Add(-1 * time.Minute).Round(time.Second)
	apiToken, _, _ := signToken(issuerName, apiKey)

	expKey := key()
	expKey.ExpiresAt = time.Now().UTC().Add(-1 * time.Minute).Round(time.Second)
	expToken, _, _ := signToken(issuerName, expKey)

	emptySubjectKey := key()
	emptySubjectKey.Subject = ""
	signedEmptySubjectTkn, parsedEmptySubjectTkn, err := signToken(issuerName, emptySubjectKey)
	require.Nil(t, err, fmt.Sprintf("issuing user key expected to succeed: %s", err))

	emptyTypeKey := key()
	emptyTypeKey.Type = auth.KeyType(auth.InvitationKey + 1)
	emptyTypeToken, _, err := signToken(issuerName, emptyTypeKey)
	require.Nil(t, err, fmt.Sprintf("issuing user key expected to succeed: %s", err))

	emptyKey := key()
	emptyKey.Subject = ""

	signedInValidTkn, parsedInvalidTkn, err := signToken("invalid.issuer", key())
	require.Nil(t, err, fmt.Sprintf("issuing key expected to succeed: %s", err))

	cases := []struct {
		desc       string
		key        auth.Key
		token      string
		managerRes jwt.Token
		managerErr error
		err        error
	}{
		{
			desc:       "parse valid key",
			key:        key(),
			token:      string(signedTkn),
			managerRes: parsedTkn,
			err:        nil,
		},
		{
			desc:       "parse invalid key",
			key:        auth.Key{},
			token:      "invalid",
			managerErr: svcerr.ErrAuthentication,
			err:        svcerr.ErrAuthentication,
		},
		{
			desc:       "parse expired key",
			key:        auth.Key{},
			token:      string(expToken),
			managerErr: errJWTExpiryKey,
			err:        auth.ErrExpiry,
		},
		{
			desc:       "parse expired API key",
			key:        apiKey,
			token:      string(apiToken),
			managerErr: errJWTExpiryKey,
			err:        auth.ErrExpiry,
		},
		{
			desc:       "parse token with invalid issuer",
			key:        auth.Key{},
			token:      string(signedInValidTkn),
			managerRes: parsedInvalidTkn,
			err:        errInvalidIssuer,
		},
		{
			desc:       "parse token with empty subject",
			key:        emptySubjectKey,
			token:      string(signedEmptySubjectTkn),
			managerRes: parsedEmptySubjectTkn,
			err:        nil,
		},
		{
			desc:       "parse token with empty type",
			key:        emptyTypeKey,
			token:      string(emptyTypeToken),
			managerRes: newToken(issuerName, emptyKey),
			err:        errors.ErrAuthentication,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			kmCall := keyManager.On("ParseJWT", tc.token).Return(tc.managerRes, tc.managerErr)
			key, err := tokenizer.Parse(tc.token)
			assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s expected %s, got %s", tc.desc, tc.err, err))
			if err == nil {
				assert.Equal(t, tc.key, key, fmt.Sprintf("%s expected %v, got %v", tc.desc, tc.key, key))
			}
			kmCall.Unset()
		})
	}
}

func key() auth.Key {
	exp := time.Now().UTC().Add(10 * time.Minute).Round(time.Second)
	return auth.Key{
		ID:        "66af4a67-3823-438a-abd7-efdb613eaef6",
		Type:      auth.AccessKey,
		Issuer:    "supermq.auth",
		Role:      auth.UserRole,
		Subject:   "66af4a67-3823-438a-abd7-efdb613eaef6",
		IssuedAt:  time.Now().UTC().Add(-10 * time.Second).Round(time.Second),
		ExpiresAt: exp,
	}
}

func signToken(issuerName string, key auth.Key) ([]byte, jwt.Token, error) {
	tkn := newToken(issuerName, key)
	pKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, nil, err
	}
	pubKey := &pKey.PublicKey
	sToken, err := jwt.Sign(tkn, jwt.WithKey(jwa.RS256, pKey))
	if err != nil {
		return nil, nil, err
	}
	pToken, err := jwt.Parse(
		sToken,
		jwt.WithValidate(true),
		jwt.WithKey(jwa.RS256, pubKey),
	)
	if err != nil {
		return nil, nil, err
	}
	return sToken, pToken, nil
}
