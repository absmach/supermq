// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package symmetric_test

import (
	"testing"
	"time"

	"github.com/absmach/supermq/auth"
	"github.com/absmach/supermq/auth/keymanager/symmetric"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewKeyManager(t *testing.T) {
	cases := []struct {
		name        string
		algorithm   string
		secret      []byte
		expectErr   bool
		errContains string
	}{
		{
			name:      "valid HS256 algorithm",
			algorithm: "HS256",
			secret:    []byte("my-secret-key-32-bytes-long!!"),
			expectErr: false,
		},
		{
			name:      "valid HS384 algorithm",
			algorithm: "HS384",
			secret:    []byte("my-secret-key-48-bytes-long-for-hs384-!!!!!!"),
			expectErr: false,
		},
		{
			name:      "valid HS512 algorithm",
			algorithm: "HS512",
			secret:    []byte("my-secret-key-64-bytes-long-for-hs512-algorithm-testing!!!!"),
			expectErr: false,
		},
		{
			name:        "invalid algorithm",
			algorithm:   "INVALID_ALG",
			secret:      []byte("my-secret-key"),
			expectErr:   true,
			errContains: "unsupported key algorithm",
		},
		{
			name:        "empty secret",
			algorithm:   "HS256",
			secret:      []byte{},
			expectErr:   true,
			errContains: "invalid symmetric key",
		},
		{
			name:        "nil secret",
			algorithm:   "HS256",
			secret:      nil,
			expectErr:   true,
			errContains: "invalid symmetric key",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			km, err := symmetric.NewKeyManager(tc.algorithm, tc.secret)

			if tc.expectErr {
				assert.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
				assert.Nil(t, km)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, km)
			}
		})
	}
}

func TestSign(t *testing.T) {
	secret := []byte("my-super-secret-key-for-testing")

	km, err := symmetric.NewKeyManager("HS256", secret)
	require.NoError(t, err)

	cases := []struct {
		name string
		key  auth.Key
	}{
		{
			name: "sign valid key with all fields",
			key: auth.Key{
				ID:        "key-id",
				Type:      auth.AccessKey,
				Issuer:    "supermq.auth",
				Subject:   "user-id",
				Role:      auth.UserRole,
				IssuedAt:  time.Now().UTC(),
				ExpiresAt: time.Now().Add(1 * time.Hour).UTC(),
				Verified:  true,
			},
		},
		{
			name: "sign key without subject",
			key: auth.Key{
				ID:        "key-id",
				Type:      auth.APIKey,
				Issuer:    "supermq.auth",
				Role:      auth.AdminRole,
				IssuedAt:  time.Now().UTC(),
				ExpiresAt: time.Now().Add(24 * time.Hour).UTC(),
				Verified:  false,
			},
		},
		{
			name: "sign key without ID",
			key: auth.Key{
				Type:      auth.AccessKey,
				Subject:   "user-id",
				Role:      auth.UserRole,
				IssuedAt:  time.Now().UTC(),
				ExpiresAt: time.Now().Add(1 * time.Hour).UTC(),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := km.Sign(tc.key)
			assert.NoError(t, err)
			assert.NotEmpty(t, token)

			// Verify the token is valid JWT format (3 parts separated by dots)
			parts := splitJWT(token)
			assert.Equal(t, 3, len(parts), "JWT should have 3 parts")
		})
	}
}

func TestVerify(t *testing.T) {
	secret := []byte("my-super-secret-key-for-testing")

	km, err := symmetric.NewKeyManager("HS256", secret)
	require.NoError(t, err)

	validKey := auth.Key{
		ID:        "key-id",
		Type:      auth.AccessKey,
		Issuer:    "supermq.auth",
		Subject:   "user-id",
		Role:      auth.UserRole,
		IssuedAt:  time.Now().UTC(),
		ExpiresAt: time.Now().Add(1 * time.Hour).UTC(),
		Verified:  true,
	}

	// Sign a valid token
	validToken, err := km.Sign(validKey)
	require.NoError(t, err)

	// Create an expired token
	expiredKey := validKey
	expiredKey.ExpiresAt = time.Now().Add(-1 * time.Hour).UTC()
	expiredToken, err := km.Sign(expiredKey)
	require.NoError(t, err)

	// Create token with wrong issuer using another key manager
	wrongIssuerKey := validKey
	wrongIssuerKey.Issuer = "wrong.issuer"

	builder := jwt.NewBuilder()
	builder.Issuer(wrongIssuerKey.Issuer).
		Subject(wrongIssuerKey.Subject).
		IssuedAt(wrongIssuerKey.IssuedAt).
		Expiration(wrongIssuerKey.ExpiresAt).
		JwtID(wrongIssuerKey.ID).
		Claim("type", wrongIssuerKey.Type).
		Claim("role", wrongIssuerKey.Role).
		Claim("verified", wrongIssuerKey.Verified)

	wrongIssuerJWT, err := builder.Build()
	require.NoError(t, err)

	wrongIssuerTokenBytes, err := jwt.Sign(wrongIssuerJWT, jwt.WithKey(jwa.HS256, secret))
	require.NoError(t, err)
	wrongIssuerToken := string(wrongIssuerTokenBytes)

	// Create token signed with different secret
	wrongSecretKM, err := symmetric.NewKeyManager("HS256", []byte("different-secret-key-here"))
	require.NoError(t, err)
	wrongSecretToken, err := wrongSecretKM.Sign(validKey)
	require.NoError(t, err)

	cases := []struct {
		name        string
		token       string
		expectErr   bool
		errContains string
	}{
		{
			name:      "verify valid token",
			token:     validToken,
			expectErr: false,
		},
		{
			name:        "verify expired token",
			token:       expiredToken,
			expectErr:   true,
			errContains: "exp",
		},
		{
			name:        "verify token with wrong issuer",
			token:       wrongIssuerToken,
			expectErr:   true,
			errContains: "invalid token issuer",
		},
		{
			name:        "verify token with wrong secret",
			token:       wrongSecretToken,
			expectErr:   true,
			errContains: "",
		},
		{
			name:        "verify malformed token",
			token:       "not.a.valid.jwt",
			expectErr:   true,
			errContains: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			key, err := km.Verify(tc.token)

			if tc.expectErr {
				assert.Error(t, err)
				if tc.errContains != "" {
					assert.Contains(t, err.Error(), tc.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, validKey.Subject, key.Subject)
				assert.Equal(t, validKey.Type, key.Type)
				assert.Equal(t, validKey.Role, key.Role)
			}
		})
	}
}

func TestPublicKeys(t *testing.T) {
	secret := []byte("my-super-secret-key-for-testing")

	km, err := symmetric.NewKeyManager("HS256", secret)
	require.NoError(t, err)

	keys, err := km.PublicKeys()
	assert.Error(t, err)
	assert.Equal(t, auth.ErrPublicKeysNotSupported, err)
	assert.Nil(t, keys)
}

func TestSignAndVerifyRoundTrip(t *testing.T) {
	algorithms := []string{"HS256", "HS384", "HS512"}

	for _, alg := range algorithms {
		t.Run(alg, func(t *testing.T) {
			secret := []byte("my-super-secret-key-for-testing-" + alg)

			km, err := symmetric.NewKeyManager(alg, secret)
			require.NoError(t, err)

			originalKey := auth.Key{
				ID:        "key-123",
				Type:      auth.AccessKey,
				Issuer:    "supermq.auth",
				Subject:   "user-456",
				Role:      auth.UserRole,
				IssuedAt:  time.Now().UTC().Truncate(time.Second),
				ExpiresAt: time.Now().Add(1 * time.Hour).UTC().Truncate(time.Second),
				Verified:  true,
			}

			// Sign
			token, err := km.Sign(originalKey)
			require.NoError(t, err)

			// Verify
			verifiedKey, err := km.Verify(token)
			require.NoError(t, err)

			// Compare
			assert.Equal(t, originalKey.ID, verifiedKey.ID)
			assert.Equal(t, originalKey.Type, verifiedKey.Type)
			assert.Equal(t, originalKey.Subject, verifiedKey.Subject)
			assert.Equal(t, originalKey.Role, verifiedKey.Role)
			assert.Equal(t, originalKey.Verified, verifiedKey.Verified)
			assert.WithinDuration(t, originalKey.IssuedAt, verifiedKey.IssuedAt, time.Second)
			assert.WithinDuration(t, originalKey.ExpiresAt, verifiedKey.ExpiresAt, time.Second)
		})
	}
}

func TestDifferentAlgorithms(t *testing.T) {
	secret := []byte("my-super-secret-key-for-testing-algorithms")

	key := auth.Key{
		ID:        "key-id",
		Type:      auth.AccessKey,
		Issuer:    "supermq.auth",
		Subject:   "user-id",
		Role:      auth.UserRole,
		IssuedAt:  time.Now().UTC(),
		ExpiresAt: time.Now().Add(1 * time.Hour).UTC(),
		Verified:  true,
	}

	// Sign with HS256
	km256, err := symmetric.NewKeyManager("HS256", secret)
	require.NoError(t, err)
	token256, err := km256.Sign(key)
	require.NoError(t, err)

	// Sign with HS384
	km384, err := symmetric.NewKeyManager("HS384", secret)
	require.NoError(t, err)
	token384, err := km384.Sign(key)
	require.NoError(t, err)

	// Sign with HS512
	km512, err := symmetric.NewKeyManager("HS512", secret)
	require.NoError(t, err)
	token512, err := km512.Sign(key)
	require.NoError(t, err)

	// Each algorithm should produce different tokens
	assert.NotEqual(t, token256, token384)
	assert.NotEqual(t, token256, token512)
	assert.NotEqual(t, token384, token512)

	// Verify with correct key manager
	_, err = km256.Verify(token256)
	assert.NoError(t, err)

	_, err = km384.Verify(token384)
	assert.NoError(t, err)

	_, err = km512.Verify(token512)
	assert.NoError(t, err)

	// Cross-verification should fail (token signed with HS256, verified with HS384)
	_, err = km384.Verify(token256)
	assert.Error(t, err)

	_, err = km512.Verify(token256)
	assert.Error(t, err)
}

// Helper function to split JWT into parts
func splitJWT(token string) []string {
	parts := []string{}
	start := 0
	for i := 0; i < len(token); i++ {
		if token[i] == '.' {
			parts = append(parts, token[start:i])
			start = i + 1
		}
	}
	parts = append(parts, token[start:])
	return parts
}
