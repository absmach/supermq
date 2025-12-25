// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"errors"
)

var (
	ErrUnsupportedKeyAlgorithm = errors.New("unsupported key algorithm")
	ErrInvalidSymmetricKey     = errors.New("invalid symmetric key")
	ErrPublicKeysNotSupported  = errors.New("public keys not supported for symmetric algorithm")
)

// PublicKeyInfo represents a public key for external distribution via JWKS.
// This follows RFC 7517 (JSON Web Key) specification.
type PublicKeyInfo struct {
	KeyID     string `json:"kid"`
	KeyType   string `json:"kty"`
	Algorithm string `json:"alg"`
	Use       string `json:"use,omitempty"`

	// EdDSA (Ed25519) fields
	Curve string `json:"crv,omitempty"`
	X     string `json:"x,omitempty"`

	// Future: RSA fields (n, e), ECDSA fields (x, y, crv), etc.
}

// KeyManager manages cryptographic keys for JWT operations.
// Implementations handle both signing/verification and key distribution.
type KeyManager interface {
	// Sign creates a signed JWT token string from the given key claims.
	Sign(key Key) (signedToken string, err error)

	// Verify verifies and parses a JWT token string, returning the extracted claims.
	Verify(tokenString string) (key Key, err error)

	// PublicKeys returns public keys for distribution via JWKS endpoint.
	// Returns ErrPublicKeysNotSupported for symmetric key managers (HMAC).
	PublicKeys() ([]PublicKeyInfo, error)
}

// IsSymmetricAlgorithm determines if the given algorithm is symmetric (HMAC-based).
// Returns true for HMAC algorithms (HS256, HS384, HS512).
// Returns false for asymmetric algorithms (EdDSA).
// Returns error for unsupported algorithms.
func IsSymmetricAlgorithm(alg string) (bool, error) {
	switch alg {
	case "EdDSA":
		return false, nil
	case "HS256", "HS384", "HS512":
		return true, nil
	default:
		return false, ErrUnsupportedKeyAlgorithm
	}
}
