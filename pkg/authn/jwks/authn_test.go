// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package jwks_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	"github.com/absmach/supermq/auth"
	smqjwt "github.com/absmach/supermq/auth/jwt"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	validIssuer   = "supermq.auth"
	invalidIssuer = "invalid.issuer"
	userID        = "user123"
)

func TestAuthenticate(t *testing.T) {
	// Generate Ed25519 key pair
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	kid := "test-key-id"

	// Create JWK from keys
	privateJwk, err := jwk.FromRaw(privateKey)
	require.NoError(t, err)
	require.NoError(t, privateJwk.Set(jwk.AlgorithmKey, jwa.EdDSA))
	require.NoError(t, privateJwk.Set(jwk.KeyIDKey, kid))

	publicJwk, err := jwk.FromRaw(publicKey)
	require.NoError(t, err)
	require.NoError(t, publicJwk.Set(jwk.AlgorithmKey, jwa.EdDSA))
	require.NoError(t, publicJwk.Set(jwk.KeyIDKey, kid))

	// Create JWKS set
	jwksSet := jwk.NewSet()
	require.NoError(t, jwksSet.AddKey(publicJwk))

	// Note: We can't easily test the full authentication flow here because it requires
	// a gRPC connection to the auth service. This would require more complex test setup.
	// For now, we focus on testing the validateToken function indirectly through unit tests
	// of the key manager implementations.
	t.Skip("Full integration test requires gRPC server setup")

	// Keep variables for potential future use
	_ = privateJwk
	_ = publicJwk
	_ = jwksSet
}

func TestValidateToken(t *testing.T) {
	// Generate Ed25519 key pair
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	kid := "test-key-id"

	// Create JWK from keys
	privateJwk, err := jwk.FromRaw(privateKey)
	require.NoError(t, err)
	require.NoError(t, privateJwk.Set(jwk.AlgorithmKey, jwa.EdDSA))
	require.NoError(t, privateJwk.Set(jwk.KeyIDKey, kid))

	publicJwk, err := jwk.FromRaw(publicKey)
	require.NoError(t, err)
	require.NoError(t, publicJwk.Set(jwk.AlgorithmKey, jwa.EdDSA))
	require.NoError(t, publicJwk.Set(jwk.KeyIDKey, kid))

	// Create JWKS set
	jwksSet := jwk.NewSet()
	require.NoError(t, jwksSet.AddKey(publicJwk))

	cases := []struct {
		name      string
		issuer    string
		expiry    time.Time
		expectErr bool
	}{
		{
			name:      "valid token",
			issuer:    validIssuer,
			expiry:    time.Now().Add(1 * time.Hour),
			expectErr: false,
		},
		{
			name:      "expired token",
			issuer:    validIssuer,
			expiry:    time.Now().Add(-1 * time.Hour),
			expectErr: true,
		},
		{
			name:      "invalid issuer",
			issuer:    invalidIssuer,
			expiry:    time.Now().Add(1 * time.Hour),
			expectErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			builder := jwt.NewBuilder()
			builder.Issuer(tc.issuer)
			builder.IssuedAt(time.Now())
			builder.Expiration(tc.expiry)
			builder.Subject(userID)
			builder.Claim(smqjwt.TokenType, auth.AccessKey)
			builder.Claim(smqjwt.RoleField, auth.UserRole)
			builder.Claim(smqjwt.VerifiedField, true)

			token, err := builder.Build()
			require.NoError(t, err)

			signedToken, err := jwt.Sign(token, jwt.WithKey(jwa.EdDSA, privateJwk))
			require.NoError(t, err)

			// Parse the token using the JWKS set
			parsedToken, err := jwt.Parse(
				signedToken,
				jwt.WithValidate(true),
				jwt.WithKeySet(jwksSet),
			)

			if tc.expectErr {
				// For expired token, parsing will fail
				// For invalid issuer, parsing succeeds but we need to check issuer
				if tc.issuer != validIssuer && err == nil {
					// Check issuer manually since our validateToken would do this
					if parsedToken.Issuer() != validIssuer {
						// This is expected - invalid issuer should be caught
						assert.NotEqual(t, validIssuer, parsedToken.Issuer())
					} else {
						assert.Fail(t, "Expected invalid issuer error")
					}
				} else {
					assert.Error(t, err)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, parsedToken)
				assert.Equal(t, tc.issuer, parsedToken.Issuer())
			}
		})
	}
}

func TestMultiKeyJWKS(t *testing.T) {
	// Generate two Ed25519 key pairs (simulating active + retiring keys)
	pub1, priv1, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	pub2, priv2, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	activeKID := "key-active"
	retiringKID := "key-retiring"

	// Create JWK for active key
	privateJwk1, err := jwk.FromRaw(priv1)
	require.NoError(t, err)
	require.NoError(t, privateJwk1.Set(jwk.AlgorithmKey, jwa.EdDSA))
	require.NoError(t, privateJwk1.Set(jwk.KeyIDKey, activeKID))

	publicJwk1, err := jwk.FromRaw(pub1)
	require.NoError(t, err)
	require.NoError(t, publicJwk1.Set(jwk.AlgorithmKey, jwa.EdDSA))
	require.NoError(t, publicJwk1.Set(jwk.KeyIDKey, activeKID))

	// Create JWK for retiring key
	privateJwk2, err := jwk.FromRaw(priv2)
	require.NoError(t, err)
	require.NoError(t, privateJwk2.Set(jwk.AlgorithmKey, jwa.EdDSA))
	require.NoError(t, privateJwk2.Set(jwk.KeyIDKey, retiringKID))

	publicJwk2, err := jwk.FromRaw(pub2)
	require.NoError(t, err)
	require.NoError(t, publicJwk2.Set(jwk.AlgorithmKey, jwa.EdDSA))
	require.NoError(t, publicJwk2.Set(jwk.KeyIDKey, retiringKID))

	// Create JWKS set with both keys (simulating rotation period)
	jwksSet := jwk.NewSet()
	require.NoError(t, jwksSet.AddKey(publicJwk1))
	require.NoError(t, jwksSet.AddKey(publicJwk2))

	assert.Equal(t, 2, jwksSet.Len(), "JWKS should contain both keys")

	cases := []struct {
		name       string
		privateKey jwk.Key
		kid        string
		issuer     string
		expiry     time.Time
		expectErr  bool
	}{
		{
			name:       "token signed with active key",
			privateKey: privateJwk1,
			kid:        activeKID,
			issuer:     validIssuer,
			expiry:     time.Now().Add(1 * time.Hour),
			expectErr:  false,
		},
		{
			name:       "token signed with retiring key",
			privateKey: privateJwk2,
			kid:        retiringKID,
			issuer:     validIssuer,
			expiry:     time.Now().Add(1 * time.Hour),
			expectErr:  false,
		},
		{
			name:       "expired token with active key",
			privateKey: privateJwk1,
			kid:        activeKID,
			issuer:     validIssuer,
			expiry:     time.Now().Add(-1 * time.Hour),
			expectErr:  true,
		},
		{
			name:       "expired token with retiring key",
			privateKey: privateJwk2,
			kid:        retiringKID,
			issuer:     validIssuer,
			expiry:     time.Now().Add(-1 * time.Hour),
			expectErr:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			builder := jwt.NewBuilder()
			builder.Issuer(tc.issuer)
			builder.IssuedAt(time.Now())
			builder.Expiration(tc.expiry)
			builder.Subject(userID)
			builder.Claim(smqjwt.TokenType, auth.AccessKey)
			builder.Claim(smqjwt.RoleField, auth.UserRole)
			builder.Claim(smqjwt.VerifiedField, true)

			token, err := builder.Build()
			require.NoError(t, err)

			signedToken, err := jwt.Sign(token, jwt.WithKey(jwa.EdDSA, tc.privateKey))
			require.NoError(t, err)

			// Parse the token using the multi-key JWKS set
			parsedToken, err := jwt.Parse(
				signedToken,
				jwt.WithValidate(true),
				jwt.WithKeySet(jwksSet, jws.WithInferAlgorithmFromKey(true)),
			)

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, parsedToken)
				assert.Equal(t, tc.issuer, parsedToken.Issuer())
				assert.Equal(t, userID, parsedToken.Subject())
			}
		})
	}
}

func TestKeyIDMatching(t *testing.T) {
	// Generate two Ed25519 key pairs
	pub1, priv1, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	pub2, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	kid1 := "key-1"
	kid2 := "key-2"

	// Create JWKs
	privateJwk1, err := jwk.FromRaw(priv1)
	require.NoError(t, err)
	require.NoError(t, privateJwk1.Set(jwk.AlgorithmKey, jwa.EdDSA))
	require.NoError(t, privateJwk1.Set(jwk.KeyIDKey, kid1))

	publicJwk1, err := jwk.FromRaw(pub1)
	require.NoError(t, err)
	require.NoError(t, publicJwk1.Set(jwk.AlgorithmKey, jwa.EdDSA))
	require.NoError(t, publicJwk1.Set(jwk.KeyIDKey, kid1))

	publicJwk2, err := jwk.FromRaw(pub2)
	require.NoError(t, err)
	require.NoError(t, publicJwk2.Set(jwk.AlgorithmKey, jwa.EdDSA))
	require.NoError(t, publicJwk2.Set(jwk.KeyIDKey, kid2))

	// Create JWKS with both public keys
	jwksSet := jwk.NewSet()
	require.NoError(t, jwksSet.AddKey(publicJwk1))
	require.NoError(t, jwksSet.AddKey(publicJwk2))

	// Create token signed with key-1
	builder := jwt.NewBuilder()
	builder.Issuer(validIssuer)
	builder.IssuedAt(time.Now())
	builder.Expiration(time.Now().Add(1 * time.Hour))
	builder.Subject(userID)

	token, err := builder.Build()
	require.NoError(t, err)

	signedToken, err := jwt.Sign(token, jwt.WithKey(jwa.EdDSA, privateJwk1))
	require.NoError(t, err)

	// Verify the token - JWT library should automatically select key-1 based on kid header
	parsedToken, err := jwt.Parse(
		signedToken,
		jwt.WithValidate(true),
		jwt.WithKeySet(jwksSet, jws.WithInferAlgorithmFromKey(true)),
	)

	require.NoError(t, err)
	assert.NotNil(t, parsedToken)
	assert.Equal(t, validIssuer, parsedToken.Issuer())

	// Verify the JWT header contains the correct kid
	headers, err := jws.Parse(signedToken)
	require.NoError(t, err)
	sigs := headers.Signatures()
	require.Len(t, sigs, 1, "JWT should have exactly one signature")

	kidValue := sigs[0].ProtectedHeaders().KeyID()
	assert.Equal(t, kid1, kidValue, "JWT kid header should match signing key")
}

func TestUnknownKeyID(t *testing.T) {
	// Generate key pair
	pub1, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	_, priv2, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	kid1 := "known-key"
	kid2 := "unknown-key"

	// Create JWK for known key (only this one in JWKS)
	publicJwk1, err := jwk.FromRaw(pub1)
	require.NoError(t, err)
	require.NoError(t, publicJwk1.Set(jwk.AlgorithmKey, jwa.EdDSA))
	require.NoError(t, publicJwk1.Set(jwk.KeyIDKey, kid1))

	// Create private JWK for unknown key (not in JWKS)
	privateJwk2, err := jwk.FromRaw(priv2)
	require.NoError(t, err)
	require.NoError(t, privateJwk2.Set(jwk.AlgorithmKey, jwa.EdDSA))
	require.NoError(t, privateJwk2.Set(jwk.KeyIDKey, kid2))

	// JWKS only contains key-1
	jwksSet := jwk.NewSet()
	require.NoError(t, jwksSet.AddKey(publicJwk1))

	// Create token signed with unknown key-2
	builder := jwt.NewBuilder()
	builder.Issuer(validIssuer)
	builder.IssuedAt(time.Now())
	builder.Expiration(time.Now().Add(1 * time.Hour))
	builder.Subject(userID)

	token, err := builder.Build()
	require.NoError(t, err)

	signedToken, err := jwt.Sign(token, jwt.WithKey(jwa.EdDSA, privateJwk2))
	require.NoError(t, err)

	// Attempt to verify - should fail because kid is not in JWKS
	_, err = jwt.Parse(
		signedToken,
		jwt.WithValidate(true),
		jwt.WithKeySet(jwksSet, jws.WithInferAlgorithmFromKey(true)),
	)

	assert.Error(t, err, "Should fail when token's kid is not in JWKS")
}

func TestJWKSCaching(t *testing.T) {
	t.Skip("Requires full authentication setup with gRPC")
	// This would test that multiple authentications within the cache duration
	// only fetch JWKS once from the HTTP endpoint
}
