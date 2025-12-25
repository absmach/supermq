// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package asymmetric_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/absmach/supermq/auth"
	smqjwt "github.com/absmach/supermq/auth/jwt"
	"github.com/absmach/supermq/auth/keymanager/asymmetric"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyRotation(t *testing.T) {
	tmpDir := t.TempDir()

	// Generate two key pairs
	pub1, priv1, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	pub2, priv2, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	// Save keys
	key1Path := filepath.Join(tmpDir, "key-active.pem")
	key2Path := filepath.Join(tmpDir, "key-retiring.pem")

	saveKey(t, priv1, key1Path)
	saveKey(t, priv2, key2Path)

	// Create metadata with overlapping keys
	metadata := map[string]interface{}{
		"active_key_id": "key-active",
		"keys": []map[string]interface{}{
			{
				"id":         "key-active",
				"file":       "key-active.pem",
				"created_at": time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
				"status":     "active",
			},
			{
				"id":         "key-retiring",
				"file":       "key-retiring.pem",
				"created_at": time.Now().Add(-8 * 24 * time.Hour).Format(time.RFC3339),
				"status":     "retiring",
				"expires_at": time.Now().Add(6 * 24 * time.Hour).Format(time.RFC3339), // 6 days left
			},
		},
	}

	metadataPath := filepath.Join(tmpDir, "keys.json")
	data, err := json.MarshalIndent(metadata, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(metadataPath, data, 0600)
	require.NoError(t, err)

	// Create key manager (will use multi-key mode)
	km, err := asymmetric.NewKeyManager(filepath.Join(tmpDir, "dummy.pem"), &mockIDProvider{id: "ignored"})
	require.NoError(t, err)

	// Test 1: Sign with active key
	testKey := auth.Key{
		ID:        "test-key",
		Type:      auth.AccessKey,
		Subject:   "user-123",
		Role:      auth.UserRole,
		IssuedAt:  time.Now().UTC(),
		ExpiresAt: time.Now().Add(1 * time.Hour).UTC(),
		Verified:  true,
	}

	token, err := km.Sign(testKey)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Test 2: Verify with active key (the key used for signing)
	verified, err := km.Verify(token)
	require.NoError(t, err)
	assert.Equal(t, testKey.Subject, verified.Subject)

	// Test 3: Create a token signed with the retiring key (simulating old token)
	// Manually create a token with the retiring key
	oldToken := createTokenWithKey(t, priv2, "key-retiring", testKey)

	// Test 4: Verify old token signed with retiring key (should succeed during grace period)
	verifiedOld, err := km.Verify(oldToken)
	require.NoError(t, err, "Token signed with retiring key should still be valid during grace period")
	assert.Equal(t, testKey.Subject, verifiedOld.Subject)

	// Test 5: Public keys should return both active and retiring keys
	publicKeys, err := km.PublicKeys()
	require.NoError(t, err)
	assert.Len(t, publicKeys, 2, "Should return both active and retiring keys")

	// Verify both key IDs are present
	keyIDs := make(map[string]bool)
	for _, pk := range publicKeys {
		keyIDs[pk.KeyID] = true
	}
	assert.True(t, keyIDs["key-active"], "Active key should be in JWKS")
	assert.True(t, keyIDs["key-retiring"], "Retiring key should be in JWKS")

	// Keep for future use
	_ = pub1
	_ = pub2
}

func TestRetiredKey(t *testing.T) {
	tmpDir := t.TempDir()

	// Generate keys
	_, priv1, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	_, priv2, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	// Save keys
	key1Path := filepath.Join(tmpDir, "key-active.pem")
	key2Path := filepath.Join(tmpDir, "key-retired.pem")

	saveKey(t, priv1, key1Path)
	saveKey(t, priv2, key2Path)

	// Create metadata with retired key (expired retiring key)
	metadata := map[string]interface{}{
		"active_key_id": "key-active",
		"keys": []map[string]interface{}{
			{
				"id":         "key-active",
				"file":       "key-active.pem",
				"created_at": time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
				"status":     "active",
			},
			{
				"id":         "key-retired",
				"file":       "key-retired.pem",
				"created_at": time.Now().Add(-20 * 24 * time.Hour).Format(time.RFC3339),
				"status":     "retiring",
				"expires_at": time.Now().Add(-1 * time.Hour).Format(time.RFC3339), // Expired 1 hour ago
			},
		},
	}

	metadataPath := filepath.Join(tmpDir, "keys.json")
	data, err := json.MarshalIndent(metadata, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(metadataPath, data, 0600)
	require.NoError(t, err)

	// Create key manager
	km, err := asymmetric.NewKeyManager(filepath.Join(tmpDir, "dummy.pem"), &mockIDProvider{id: "ignored"})
	require.NoError(t, err)

	// Public keys should only return active key (retired key filtered out)
	publicKeys, err := km.PublicKeys()
	require.NoError(t, err)
	assert.Len(t, publicKeys, 1, "Should only return active key, not retired")
	assert.Equal(t, "key-active", publicKeys[0].KeyID)
}

func TestExplicitRetiredStatus(t *testing.T) {
	tmpDir := t.TempDir()

	// Generate keys
	_, priv1, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	_, priv2, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	// Save keys
	key1Path := filepath.Join(tmpDir, "key-active.pem")
	key2Path := filepath.Join(tmpDir, "key-retired.pem")

	saveKey(t, priv1, key1Path)
	saveKey(t, priv2, key2Path)

	// Create metadata with explicit retired status
	metadata := map[string]interface{}{
		"active_key_id": "key-active",
		"keys": []map[string]interface{}{
			{
				"id":         "key-active",
				"file":       "key-active.pem",
				"created_at": time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
				"status":     "active",
			},
			{
				"id":         "key-retired",
				"file":       "key-retired.pem",
				"created_at": time.Now().Add(-30 * 24 * time.Hour).Format(time.RFC3339),
				"status":     "retired",
			},
		},
	}

	metadataPath := filepath.Join(tmpDir, "keys.json")
	data, err := json.MarshalIndent(metadata, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(metadataPath, data, 0600)
	require.NoError(t, err)

	// Create key manager
	km, err := asymmetric.NewKeyManager(filepath.Join(tmpDir, "dummy.pem"), &mockIDProvider{id: "ignored"})
	require.NoError(t, err)

	// Public keys should only return active key (retired status keys filtered out)
	publicKeys, err := km.PublicKeys()
	require.NoError(t, err)
	assert.Len(t, publicKeys, 1, "Should only return active key, keys with retired status filtered out")
	assert.Equal(t, "key-active", publicKeys[0].KeyID)
}

func TestRetiringKeyWithoutExpiry(t *testing.T) {
	tmpDir := t.TempDir()

	// Generate keys
	_, priv1, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	_, priv2, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	// Save keys
	key1Path := filepath.Join(tmpDir, "key-active.pem")
	key2Path := filepath.Join(tmpDir, "key-retiring.pem")

	saveKey(t, priv1, key1Path)
	saveKey(t, priv2, key2Path)

	// Create metadata with retiring key missing expires_at
	metadata := map[string]interface{}{
		"active_key_id": "key-active",
		"keys": []map[string]interface{}{
			{
				"id":         "key-active",
				"file":       "key-active.pem",
				"created_at": time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
				"status":     "active",
			},
			{
				"id":         "key-retiring",
				"file":       "key-retiring.pem",
				"created_at": time.Now().Add(-8 * 24 * time.Hour).Format(time.RFC3339),
				"status":     "retiring",
				// Missing expires_at - should fail validation
			},
		},
	}

	metadataPath := filepath.Join(tmpDir, "keys.json")
	data, err := json.MarshalIndent(metadata, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(metadataPath, data, 0600)
	require.NoError(t, err)

	// Create key manager - should fail because retiring key lacks expires_at
	_, err = asymmetric.NewKeyManager(filepath.Join(tmpDir, "dummy.pem"), &mockIDProvider{id: "ignored"})
	assert.Error(t, err, "Should fail when retiring key missing expires_at")
	assert.Contains(t, err.Error(), "retiring keys must have expires_at set")
}

func TestBackwardCompatibility(t *testing.T) {
	// Test that old single-key mode still works without keys.json
	tmpDir := t.TempDir()

	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	keyPath := filepath.Join(tmpDir, "private.key")
	saveKey(t, privateKey, keyPath)

	// Create without keys.json (should fall back to single key mode)
	km, err := asymmetric.NewKeyManager(keyPath, &mockIDProvider{id: "test-kid"})
	require.NoError(t, err)

	// Should work normally
	testKey := auth.Key{
		ID:        "test",
		Type:      auth.AccessKey,
		Subject:   "user",
		Role:      auth.UserRole,
		IssuedAt:  time.Now().UTC(),
		ExpiresAt: time.Now().Add(1 * time.Hour).UTC(),
	}

	token, err := km.Sign(testKey)
	require.NoError(t, err)

	_, err = km.Verify(token)
	require.NoError(t, err)

	publicKeys, err := km.PublicKeys()
	require.NoError(t, err)
	assert.Len(t, publicKeys, 1)
}

// Helper functions

func saveKey(t *testing.T, privateKey ed25519.PrivateKey, path string) {
	pkcs8Key, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)

	pemBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Key,
	}

	err = os.WriteFile(path, pem.EncodeToMemory(pemBlock), 0600)
	require.NoError(t, err)
}

func createTokenWithKey(t *testing.T, privateKey ed25519.PrivateKey, kid string, key auth.Key) string {
	// Create JWK from private key
	privateJwk, err := jwk.FromRaw(privateKey)
	require.NoError(t, err)
	require.NoError(t, privateJwk.Set(jwk.AlgorithmKey, jwa.EdDSA))
	require.NoError(t, privateJwk.Set(jwk.KeyIDKey, kid))

	// Build JWT token
	builder := jwt.NewBuilder()
	builder.
		Issuer(smqjwt.IssuerName).
		IssuedAt(key.IssuedAt).
		Expiration(key.ExpiresAt).
		Subject(key.Subject).
		JwtID(key.ID).
		Claim(smqjwt.TokenType, key.Type).
		Claim(smqjwt.RoleField, key.Role).
		Claim(smqjwt.VerifiedField, key.Verified)

	tkn, err := builder.Build()
	require.NoError(t, err)

	// Sign the token
	signedBytes, err := jwt.Sign(tkn, jwt.WithKey(jwa.EdDSA, privateJwk))
	require.NoError(t, err)

	return string(signedBytes)
}
