// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package symmetric

import (
	"github.com/absmach/supermq/auth"
	smqjwt "github.com/absmach/supermq/auth/jwt"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type manager struct {
	algorithm jwa.KeyAlgorithm
	secret    []byte
}

var _ auth.KeyManager = (*manager)(nil)

func NewKeyManager(algorithm string, secret []byte) (auth.KeyManager, error) {
	alg := jwa.KeyAlgorithmFrom(algorithm)
	if _, ok := alg.(jwa.InvalidKeyAlgorithm); ok {
		return nil, auth.ErrUnsupportedKeyAlgorithm
	}
	if len(secret) == 0 {
		return nil, auth.ErrInvalidSymmetricKey
	}
	return &manager{
		secret:    secret,
		algorithm: alg,
	}, nil
}

func (km *manager) Sign(key auth.Key) (string, error) {
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

	signedBytes, err := jwt.Sign(tkn, jwt.WithKey(km.algorithm, km.secret))
	if err != nil {
		return "", err
	}

	return string(signedBytes), nil
}

func (km *manager) Verify(tokenString string) (auth.Key, error) {
	tkn, err := jwt.Parse(
		[]byte(tokenString),
		jwt.WithValidate(true),
		jwt.WithKey(km.algorithm, km.secret),
	)
	if err != nil {
		return auth.Key{}, err
	}

	// Validate issuer
	if tkn.Issuer() != smqjwt.IssuerName {
		return auth.Key{}, smqjwt.ErrInvalidIssuer
	}

	return smqjwt.ToKey(tkn)
}

func (km *manager) PublicKeys() ([]auth.PublicKeyInfo, error) {
	return nil, auth.ErrPublicKeysNotSupported
}
