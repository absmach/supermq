// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package jwt

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"

	"github.com/absmach/supermq/auth"
	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	"github.com/absmach/supermq/pkg/uuid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

var (
	// errInvalidIssuer is returned when the issuer is not supermq.auth.
	errInvalidIssuer = errors.New("invalid token issuer value")
	// errInvalidType is returned when there is no type field.
	errInvalidType = errors.New("invalid token type")
	// errInvalidRole is returned when the role is invalid.
	errInvalidRole = errors.New("invalid role")
	// errInvalidVerified is returned when the verified is invalid.
	errInvalidVerified = errors.New("invalid verified")
	// errJWTExpiryKey is used to check if the token is expired.
	errJWTExpiryKey = errors.New(`"exp" not satisfied`)
	// ErrSignJWT indicates an error in signing jwt token.
	ErrSignJWT = errors.New("failed to sign jwt token")
	// ErrValidateJWTToken indicates a failure to validate JWT token.
	ErrValidateJWTToken = errors.New("failed to validate jwt token")
	// ErrJSONHandle indicates an error in handling JSON.
	ErrJSONHandle = errors.New("failed to perform operation JSON")
)

const (
	issuerName    = "supermq.auth"
	tokenType     = "type"
	RoleField     = "role"
	VerifiedField = "verified"
	rsaKeyType    = "RSA"
)

type tokenizer struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

var _ auth.Tokenizer = (*tokenizer)(nil)

// New instantiates an implementation of Tokenizer service.
func New(privateKey *rsa.PrivateKey) auth.Tokenizer {
	publicKey := privateKey.Public().(*rsa.PublicKey)
	return &tokenizer{
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}

func (tok *tokenizer) Issue(key auth.Key) (string, error) {
	builder := jwt.NewBuilder()
	builder.
		Issuer(issuerName).
		IssuedAt(key.IssuedAt).
		Claim(tokenType, key.Type).
		Expiration(key.ExpiresAt)
	builder.Claim(RoleField, key.Role)
	builder.Claim(VerifiedField, key.Verified)
	if key.Subject != "" {
		builder.Subject(key.Subject)
	}
	if key.ID != "" {
		builder.JwtID(key.ID)
	}
	tkn, err := builder.Build()
	if err != nil {
		return "", errors.Wrap(svcerr.ErrAuthentication, err)
	}
	signedTkn, err := jwt.Sign(tkn, jwt.WithKey(jwa.RS256, tok.privateKey))
	if err != nil {
		return "", errors.Wrap(ErrSignJWT, err)
	}
	return string(signedTkn), nil
}

func (tok *tokenizer) Parse(token string) (auth.Key, error) {
	tkn, err := tok.validateToken(token)
	if err != nil {
		return auth.Key{}, errors.Wrap(svcerr.ErrAuthentication, err)
	}

	key, err := toKey(tkn)
	if err != nil {
		return auth.Key{}, errors.Wrap(svcerr.ErrAuthentication, err)
	}

	return key, nil
}

func (tok *tokenizer) validateToken(token string) (jwt.Token, error) {
	tkn, err := jwt.Parse(
		[]byte(token),
		jwt.WithValidate(true),
		jwt.WithKey(jwa.RS256, tok.publicKey),
	)
	if err != nil {
		if errors.Contains(err, errJWTExpiryKey) {
			return nil, auth.ErrExpiry
		}

		return nil, err
	}
	validator := jwt.ValidatorFunc(func(_ context.Context, t jwt.Token) jwt.ValidationError {
		if t.Issuer() != issuerName {
			return jwt.NewValidationError(errInvalidIssuer)
		}
		return nil
	})
	if err := jwt.Validate(tkn, jwt.WithValidator(validator)); err != nil {
		return nil, errors.Wrap(ErrValidateJWTToken, err)
	}

	return tkn, nil
}

func (tok *tokenizer) RetrieveJWKS() (auth.JWKS, error) {
	keyID, err := uuid.New().ID()
	if err != nil {
		return auth.JWKS{}, err
	}
	jwk := auth.JWK{
		Kty: rsaKeyType,
		Kid: keyID,
		N:   base64.RawURLEncoding.EncodeToString(tok.publicKey.N.Bytes()),
		E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(tok.publicKey.E)).Bytes()),
	}
	jwks := auth.JWKS{
		Keys: []auth.JWK{jwk},
	}

	return jwks, nil
}

func toKey(tkn jwt.Token) (auth.Key, error) {
	data, err := json.Marshal(tkn.PrivateClaims())
	if err != nil {
		return auth.Key{}, errors.Wrap(ErrJSONHandle, err)
	}
	var key auth.Key
	if err := json.Unmarshal(data, &key); err != nil {
		return auth.Key{}, errors.Wrap(ErrJSONHandle, err)
	}

	tType, ok := tkn.Get(tokenType)
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
