// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package asymmetric_test

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/absmach/supermq/auth"
	"github.com/absmach/supermq/auth/tokenizer/asymmetric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type incrementingIDProvider struct {
	counter int
}

func (p *incrementingIDProvider) ID() (string, error) {
	p.counter++
	return fmt.Sprintf("key-id-%d", p.counter), nil
}

func TestTwoKeyRotation(t *testing.T) {
	tmpDir := t.TempDir()

	// Generate two key pairs
	_, activePriv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	_, retiringPriv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	// Save keys
	activeKeyPath := filepath.Join(tmpDir, "active.key")
	retiringKeyPath := filepath.Join(tmpDir, "retiring.key")

	saveKey(t, activePriv, activeKeyPath)
	saveKey(t, retiringPriv, retiringKeyPath)

	// Create tokenizer with both keys (using incrementing ID provider for unique IDs)
	idProvider := &incrementingIDProvider{}
	tokenizer, err := asymmetric.NewTokenizer(activeKeyPath, retiringKeyPath, idProvider, newTestLogger())
	require.NoError(t, err)

	// Test issuing token (should use active key)
	testKey := auth.Key{
		ID:        "test-key",
		Type:      auth.AccessKey,
		Subject:   "user-123",
		Role:      auth.UserRole,
		IssuedAt:  time.Now().UTC(),
		ExpiresAt: time.Now().Add(1 * time.Hour).UTC(),
		Verified:  true,
	}

	token, err := tokenizer.Issue(testKey)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Test parsing token (should work with both keys)
	verified, err := tokenizer.Parse(context.Background(), token)
	require.NoError(t, err)
	assert.Equal(t, testKey.Subject, verified.Subject)

	// Test JWKS returns both keys
	publicKeys, err := tokenizer.RetrieveJWKS()
	require.NoError(t, err)
	assert.Len(t, publicKeys, 2, "Should return both active and retiring keys")

	// Verify both keys have different IDs
	keyIDs := make(map[string]bool)
	for _, pk := range publicKeys {
		keyIDs[pk.KeyID] = true
	}
	assert.Len(t, keyIDs, 2, "Both keys should have unique IDs")
}

func TestSingleKeyMode(t *testing.T) {
	tmpDir := t.TempDir()

	// Generate one key
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	keyPath := filepath.Join(tmpDir, "single.key")
	saveKey(t, privateKey, keyPath)

	// Create tokenizer with only active key (no retiring key)
	idProvider := &mockIDProvider{id: "single-id"}
	tokenizer, err := asymmetric.NewTokenizer(keyPath, "", idProvider, newTestLogger())
	require.NoError(t, err)

	// Test issuing and parsing
	testKey := auth.Key{
		ID:        "test",
		Type:      auth.AccessKey,
		Subject:   "user",
		Role:      auth.UserRole,
		IssuedAt:  time.Now().UTC(),
		ExpiresAt: time.Now().Add(1 * time.Hour).UTC(),
	}

	token, err := tokenizer.Issue(testKey)
	require.NoError(t, err)

	_, err = tokenizer.Parse(context.Background(), token)
	require.NoError(t, err)

	// Test JWKS returns only one key
	publicKeys, err := tokenizer.RetrieveJWKS()
	require.NoError(t, err)
	assert.Len(t, publicKeys, 1, "Should return only the active key")
}

func TestMissingRetiringKey(t *testing.T) {
	tmpDir := t.TempDir()

	// Generate only active key
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	activeKeyPath := filepath.Join(tmpDir, "active.key")
	saveKey(t, privateKey, activeKeyPath)

	// Specify retiring key path that doesn't exist
	retiringKeyPath := filepath.Join(tmpDir, "nonexistent.key")

	// Should succeed with just active key (warning logged)
	idProvider := &mockIDProvider{id: "test-id"}
	tokenizer, err := asymmetric.NewTokenizer(activeKeyPath, retiringKeyPath, idProvider, newTestLogger())
	require.NoError(t, err, "Should succeed even if retiring key is missing")

	// Should work with active key only
	testKey := auth.Key{
		ID:        "test",
		Type:      auth.AccessKey,
		Subject:   "user",
		Role:      auth.UserRole,
		IssuedAt:  time.Now().UTC(),
		ExpiresAt: time.Now().Add(1 * time.Hour).UTC(),
	}

	token, err := tokenizer.Issue(testKey)
	require.NoError(t, err)

	_, err = tokenizer.Parse(context.Background(), token)
	require.NoError(t, err)

	// JWKS should have only active key
	publicKeys, err := tokenizer.RetrieveJWKS()
	require.NoError(t, err)
	assert.Len(t, publicKeys, 1, "Should return only active key when retiring key is missing")
}

func saveKey(t *testing.T, privateKey ed25519.PrivateKey, path string) {
	pkcs8Key, err := x509.MarshalPKCS8PrivateKey(privateKey)
	require.NoError(t, err)

	pemBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Key,
	}

	err = os.WriteFile(path, pem.EncodeToMemory(pemBlock), 0o600)
	require.NoError(t, err)
}
