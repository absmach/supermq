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
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	tokenType     = "type"
	roleField     = "role"
	VerifiedField = "verified"
	issuerName    = "supermq.auth"
)

var (
	errInvalidIssuer = errors.New("invalid token issuer value")
	errJWTExpiryKey  = errors.New(`"exp" not satisfied`)
	keyManager       = new(mocks.KeyManager)
)

func TestIssue(t *testing.T) {
	tokenizer := authjwt.New(keyManager)

	validKey := key()
	signedToken, _, err := signToken(issuerName, validKey, false)
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
			key:         validKey,
			managerResp: []byte(signedToken),
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
			managerResp: []byte(signedToken),
			err:         nil,
		},
		{
			desc: "issue token without a domain",
			key: auth.Key{
				ID:       testsutil.GenerateUUID(t),
				Type:     auth.AccessKey,
				Subject:  testsutil.GenerateUUID(t),
				IssuedAt: time.Now().Add(-10 * time.Second).Round(time.Second),
			},
			managerResp: []byte(signedToken),
			err:         nil,
		},
		{
			desc: "issue token without a subject",
			key: auth.Key{
				ID:       testsutil.GenerateUUID(t),
				Type:     auth.AccessKey,
				Subject:  "",
				IssuedAt: time.Now().Add(-10 * time.Second).Round(time.Second),
			},
			managerResp: []byte(signedToken),
			err:         nil,
		},
		{
			desc: "issue token without type",
			key: auth.Key{
				ID:       testsutil.GenerateUUID(t),
				Type:     auth.KeyType(auth.InvitationKey + 1),
				Subject:  testsutil.GenerateUUID(t),
				IssuedAt: time.Now().Add(-10 * time.Second).Round(time.Second),
			},
			managerResp: []byte(signedToken),
			err:         nil,
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
			managerResp: []byte(signedToken),
			err:         nil,
		},
		{
			desc:       "issue token with failed to sign jwt",
			key:        validKey,
			managerErr: svcerr.ErrAuthentication,
			err:        authjwt.ErrSignJWT,
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			tc.managerReq = newToken(issuerName, tc.key)
			kmCall := keyManager.On("SignJWT", tc.managerReq).Return(tc.managerResp, tc.managerErr)
			tkn, err := tokenizer.Issue(tc.key)
			assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s expected %s, got %s", tc.desc, tc.err, err))
			if err == nil {
				assert.NotEmpty(t, tkn, fmt.Sprintf("%s expected token, got empty string", tc.desc))
			}
			kmCall.Unset()
		})
	}
}

func TestParse(t *testing.T) {
	tokenizer := authjwt.New(keyManager)

	signedTkn, parsedTkn, err := signToken(issuerName, key(), true)
	require.Nil(t, err, fmt.Sprintf("issuing key expected to succeed: %s", err))

	apiKey := key()
	apiKey.Type = auth.APIKey
	apiKey.ExpiresAt = time.Now().UTC().Add(-1 * time.Minute).Round(time.Second)
	apiToken, _, err := signToken(issuerName, apiKey, false)
	require.Nil(t, err, fmt.Sprintf("issuing api key expected to succeed: %s", err))

	expKey := key()
	expKey.ExpiresAt = time.Now().UTC().Add(-1 * time.Minute).Round(time.Second)
	expToken, _, err := signToken(issuerName, expKey, false)
	require.Nil(t, err, fmt.Sprintf("issuing expired key expected to succeed: %s", err))

	emptySubjectKey := key()
	emptySubjectKey.Subject = ""
	signedEmptySubjectTkn, parsedEmptySubjectTkn, err := signToken(issuerName, emptySubjectKey, true)
	require.Nil(t, err, fmt.Sprintf("issuing user key expected to succeed: %s", err))

	emptyTypeKey := key()
	emptyTypeKey.Type = auth.KeyType(auth.InvitationKey + 1)
	emptyTypeToken, _, err := signToken(issuerName, emptyTypeKey, false)
	require.Nil(t, err, fmt.Sprintf("issuing user key expected to succeed: %s", err))

	emptyKey := key()
	emptyKey.Subject = ""

	signedInValidTkn, parsedInvalidTkn, err := signToken("invalid.issuer", key(), true)
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
			token:      signedTkn,
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
			token:      expToken,
			managerErr: errJWTExpiryKey,
			err:        auth.ErrExpiry,
		},
		{
			desc:       "parse expired API key",
			key:        apiKey,
			token:      apiToken,
			managerErr: errJWTExpiryKey,
			err:        auth.ErrExpiry,
		},
		{
			desc:       "parse token with invalid issuer",
			key:        auth.Key{},
			token:      signedInValidTkn,
			managerRes: parsedInvalidTkn,
			err:        errInvalidIssuer,
		},
		{
			desc:       "parse token with empty subject",
			key:        emptySubjectKey,
			token:      signedEmptySubjectTkn,
			managerRes: parsedEmptySubjectTkn,
			err:        nil,
		},
		{
			desc:       "parse token with empty type",
			key:        emptyTypeKey,
			token:      emptyTypeToken,
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

func TestRetrieveJWKS(t *testing.T) {
	tokenizer := authjwt.New(keyManager)

	cases := []struct {
		desc string
		keys []jwk.Key
	}{
		{
			desc: "retrieve jwks with keys",
			keys: []jwk.Key{newJWK(t), newJWK(t)},
		},
		{
			desc: "retrieve empty jwks",
			keys: []jwk.Key{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			kmCall := keyManager.On("PublicJWKS").Return(tc.keys)
			jwks := tokenizer.RetrieveJWKS()
			assert.Equal(t, tc.keys, jwks, fmt.Sprintf("%s expected %v, got %v", tc.desc, tc.keys, jwks))
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

func signToken(issuerName string, key auth.Key, parseToken bool) (string, jwt.Token, error) {
	tkn := newToken(issuerName, key)
	pKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return "", nil, err
	}
	pubKey := &pKey.PublicKey
	sTkn, err := jwt.Sign(tkn, jwt.WithKey(jwa.RS256, pKey))
	if err != nil {
		return "", nil, err
	}
	if !parseToken {
		return string(sTkn), nil, nil
	}
	pTkn, err := jwt.Parse(
		sTkn,
		jwt.WithValidate(true),
		jwt.WithKey(jwa.RS256, pubKey),
	)
	if err != nil {
		return "", nil, err
	}
	return string(sTkn), pTkn, nil
}

func newJWK(t *testing.T) jwk.Key {
	pKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.Nil(t, err, fmt.Sprintf("generating rsa key expected to succeed: %s", err))
	jwkKey, err := jwk.FromRaw(&pKey.PublicKey)
	require.Nil(t, err, fmt.Sprintf("creating jwk from rsa public key expected to succeed: %s", err))
	err = jwkKey.Set(jwk.KeyIDKey, "test-key-id")
	require.Nil(t, err, fmt.Sprintf("setting jwk key id expected to succeed: %s", err))
	err = jwkKey.Set(jwk.AlgorithmKey, jwa.RS256.String())
	require.Nil(t, err, fmt.Sprintf("setting jwk algorithm expected to succeed: %s", err))
	return jwkKey
}
