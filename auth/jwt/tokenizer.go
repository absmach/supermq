// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package jwt

import (
	"context"
	"encoding/json"

	"github.com/absmach/supermq/auth"
	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

var (
	// ErrINvalidIssuer represents an invalid token issuer value.
	ErrInvalidIssuer = errors.New("invalid token issuer value")

	// ErrSignJWT indicates an error in signing jwt token.
	ErrSignJWT = errors.New("failed to sign jwt token")

	// ErrValidateJWTToken indicates a failure to validate JWT token.
	ErrValidateJWTToken = errors.New("failed to validate jwt token")

	// ErrJSONHandle indicates an error in handling JSON.
	ErrJSONHandle = errors.New("failed to perform operation JSON")

	errInvalidType     = errors.New("invalid token type")
	errInvalidRole     = errors.New("invalid role")
	errInvalidVerified = errors.New("invalid verified")
	errJWTExpiryKey    = errors.New(`"exp" not satisfied`)
)

const (
	IssuerName    = "supermq.auth"
	TokenType     = "type"
	RoleField     = "role"
	VerifiedField = "verified"
	patPrefix     = "pat"
)

type tokenizer struct {
	keyManager auth.KeyManager
}

var _ auth.Tokenizer = (*tokenizer)(nil)

// New instantiates an implementation of Tokenizer service.
func New(keyManager auth.KeyManager) auth.Tokenizer {
	return &tokenizer{
		keyManager: keyManager,
	}
}

func (tok *tokenizer) Issue(key auth.Key) (string, error) {
	signedToken, err := tok.keyManager.Sign(key)
	if err != nil {
		return "", errors.Wrap(ErrSignJWT, err)
	}
	return signedToken, nil
}

func (tok *tokenizer) Parse(ctx context.Context, token string) (auth.Key, error) {
	if len(token) >= 3 && token[:3] == patPrefix {
		return auth.Key{Type: auth.PersonalAccessToken}, nil
	}

	key, err := tok.keyManager.Verify(token)
	if err != nil {
		if errors.Contains(err, errJWTExpiryKey) {
			return auth.Key{}, errors.Wrap(svcerr.ErrAuthentication, auth.ErrExpiry)
		}
		return auth.Key{}, errors.Wrap(svcerr.ErrAuthentication, err)
	}

	return key, nil
}

func (tok *tokenizer) RetrieveJWKS() []auth.PublicKeyInfo {
	keys, err := tok.keyManager.PublicKeys()
	if err != nil {
		return nil
	}
	return keys
}

func ToKey(tkn jwt.Token) (auth.Key, error) {
	data, err := json.Marshal(tkn.PrivateClaims())
	if err != nil {
		return auth.Key{}, errors.Wrap(ErrJSONHandle, err)
	}
	var key auth.Key
	if err := json.Unmarshal(data, &key); err != nil {
		return auth.Key{}, errors.Wrap(ErrJSONHandle, err)
	}

	tType, ok := tkn.Get(TokenType)
	if !ok {
		return auth.Key{}, errInvalidType
	}
	kType, ok := tType.(float64)
	if !ok {
		return auth.Key{}, errInvalidType
	}
	kt := auth.KeyType(kType)
	if !kt.Validate() {
		return auth.Key{}, errInvalidType
	}

	tRole, ok := tkn.Get(RoleField)
	if !ok {
		return auth.Key{}, errInvalidRole
	}
	kRole, ok := tRole.(float64)
	if !ok {
		return auth.Key{}, errInvalidRole
	}

	tVerified, ok := tkn.Get(VerifiedField)
	if !ok {
		return auth.Key{}, errInvalidVerified
	}
	kVerified, ok := tVerified.(bool)
	if !ok {
		return auth.Key{}, errInvalidVerified
	}

	kr := auth.Role(kRole)
	if !kr.Validate() {
		return auth.Key{}, errInvalidRole
	}

	key.ID = tkn.JwtID()
	key.Type = auth.KeyType(kType)
	key.Role = auth.Role(kRole)
	key.Issuer = tkn.Issuer()
	key.Subject = tkn.Subject()
	key.IssuedAt = tkn.IssuedAt()
	key.ExpiresAt = tkn.Expiration()
	key.Verified = kVerified

	return key, nil
}
