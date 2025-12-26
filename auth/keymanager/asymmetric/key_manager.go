// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package asymmetric

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"os"
	"path/filepath"

	"github.com/absmach/supermq"
	"github.com/absmach/supermq/auth"
	smqjwt "github.com/absmach/supermq/auth/jwt"
	"github.com/absmach/supermq/pkg/errors"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

var (
	errLoadingPrivateKey = errors.New("failed to load private key")
	errInvalidKeySize    = errors.New("invalid ED25519 key size")
	errParsingPrivateKey = errors.New("failed to parse private key")
	errInvalidKeyType    = errors.New("private key is not ED25519")
	errGeneratingKID     = errors.New("failed to generate key ID")
	errNoValidPublicKeys = errors.New("no valid public keys available")
)

type keyPair struct {
	id         string
	privateKey jwk.Key
	publicKey  jwk.Key
}

type manager struct {
	activeKeyID string
	// Field keys is populated during initialization and never modified afterward,
	// making it safe for concurrent reads from Sign(), Verify(), and PublicKeys().
	keys map[string]*keyPair
}

var _ auth.KeyManager = (*manager)(nil)

func NewKeyManager(privateKeyDir string, idProvider supermq.IDProvider) (auth.KeyManager, error) {
	keyDir := filepath.Dir(privateKeyDir)
	metadata, err := LoadKeysMetadata(keyDir)
	if err == errNoMetadata {
		return newSingleKeyManager(privateKeyDir, idProvider)
	}
	if err != nil {
		return nil, err
	}

	return newMultiKeyManager(keyDir, metadata)
}

// newSingleKeyManager creates a manager with a single key (backward compatibility).
func newSingleKeyManager(privateKeyPath string, idProvider supermq.IDProvider) (*manager, error) {
	kid, err := idProvider.ID()
	if err != nil {
		return nil, errors.Wrap(errGeneratingKID, err)
	}
	privateJwk, publicJwk, err := loadKeyPair(privateKeyPath, kid)
	if err != nil {
		return nil, err
	}

	keys := make(map[string]*keyPair)
	keys[kid] = &keyPair{
		id:         kid,
		privateKey: privateJwk,
		publicKey:  publicJwk,
	}

	return &manager{
		activeKeyID: kid,
		keys:        keys,
	}, nil
}

// newMultiKeyManager creates a manager with multiple keys from metadata.
func newMultiKeyManager(keyDir string, metadata *KeysMetadata) (*manager, error) {
	keys := make(map[string]*keyPair)

	for _, keyMeta := range metadata.GetValidKeys() {
		keyPath := filepath.Join(keyDir, keyMeta.File)

		privateJwk, publicJwk, err := loadKeyPair(keyPath, keyMeta.ID)
		if err != nil {
			continue
		}

		keys[keyMeta.ID] = &keyPair{
			id:         keyMeta.ID,
			privateKey: privateJwk,
			publicKey:  publicJwk,
		}
	}

	if _, ok := keys[metadata.ActiveKeyID]; !ok {
		return nil, errNoActiveKeyLoaded
	}

	return &manager{
		activeKeyID: metadata.ActiveKeyID,
		keys:        keys,
	}, nil
}

func (km *manager) Sign(key auth.Key) (string, error) {
	activeKey, ok := km.keys[km.activeKeyID]
	if !ok {
		return "", errNoActiveKeyLoaded
	}

	builder := jwt.NewBuilder()
	builder.
		Issuer(smqjwt.IssuerName).
		IssuedAt(key.IssuedAt).
		Claim(smqjwt.TokenType, key.Type).
		Expiration(key.ExpiresAt).
		Claim(smqjwt.RoleField, key.Role).
		Claim(smqjwt.VerifiedField, key.Verified)

	if key.Subject != "" {
		builder.Subject(key.Subject)
	}
	if key.ID != "" {
		builder.JwtID(key.ID)
	}

	tkn, err := builder.Build()
	if err != nil {
		return "", err
	}

	signedBytes, err := jwt.Sign(tkn, jwt.WithKey(jwa.EdDSA, activeKey.privateKey))
	if err != nil {
		return "", err
	}

	return string(signedBytes), nil
}

func (km *manager) Verify(tokenString string) (auth.Key, error) {
	set := jwk.NewSet()
	for _, kp := range km.keys {
		if err := set.AddKey(kp.publicKey); err != nil {
			return auth.Key{}, err
		}
	}

	tkn, err := jwt.Parse(
		[]byte(tokenString),
		jwt.WithValidate(true),
		jwt.WithKeySet(set, jws.WithInferAlgorithmFromKey(true)),
	)
	if err != nil {
		return auth.Key{}, err
	}

	if tkn.Issuer() != smqjwt.IssuerName {
		return auth.Key{}, smqjwt.ErrInvalidIssuer
	}

	return smqjwt.ToKey(tkn)
}

func (km *manager) PublicKeys() ([]auth.PublicKeyInfo, error) {
	publicKeys := make([]auth.PublicKeyInfo, 0, len(km.keys))
	for _, kp := range km.keys {
		var rawKey ed25519.PublicKey
		if err := kp.publicKey.Raw(&rawKey); err != nil {
			continue
		}

		publicKeys = append(publicKeys, auth.PublicKeyInfo{
			KeyID:     kp.id,
			KeyType:   "OKP",
			Algorithm: "EdDSA",
			Use:       "sig",
			Curve:     "Ed25519",
			X:         base64.RawURLEncoding.EncodeToString(rawKey),
		})
	}

	if len(publicKeys) == 0 {
		return nil, errNoValidPublicKeys
	}

	return publicKeys, nil
}

func loadKeyPair(privateKeyPath string, kid string) (jwk.Key, jwk.Key, error) {
	privateKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, nil, errors.Wrap(errLoadingPrivateKey, err)
	}

	var privateKey ed25519.PrivateKey
	block, _ := pem.Decode(privateKeyBytes)
	switch {
	case block != nil:
		parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, nil, errors.Wrap(errParsingPrivateKey, err)
		}
		var ok bool
		privateKey, ok = parsedKey.(ed25519.PrivateKey)
		if !ok {
			return nil, nil, errInvalidKeyType
		}
	default:
		if len(privateKeyBytes) != ed25519.PrivateKeySize {
			return nil, nil, errInvalidKeySize
		}
		privateKey = ed25519.PrivateKey(privateKeyBytes)
	}

	publicKey := privateKey.Public().(ed25519.PublicKey)

	privateJwk, err := jwk.FromRaw(privateKey)
	if err != nil {
		return nil, nil, err
	}
	if err := privateJwk.Set(jwk.AlgorithmKey, jwa.EdDSA); err != nil {
		return nil, nil, err
	}
	if err := privateJwk.Set(jwk.KeyIDKey, kid); err != nil {
		return nil, nil, err
	}

	publicJwk, err := jwk.FromRaw(publicKey)
	if err != nil {
		return nil, nil, err
	}
	if err := publicJwk.Set(jwk.AlgorithmKey, jwa.EdDSA); err != nil {
		return nil, nil, err
	}
	if err := publicJwk.Set(jwk.KeyIDKey, kid); err != nil {
		return nil, nil, err
	}

	return privateJwk, publicJwk, nil
}
