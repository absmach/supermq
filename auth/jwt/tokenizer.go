// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package jwt

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"strconv"

	"github.com/absmach/magistrala/auth"
	"github.com/absmach/magistrala/pkg/errors"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	"github.com/absmach/magistrala/pkg/uuid"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

var (
	errInvalidIssuer = errors.New("invalid token issuer value")
	// errJWTExpiryKey is used to check if the token is expired.
	errJWTExpiryKey = errors.New(`"exp" not satisfied`)
	// ErrSignJWT indicates an error in signing jwt token.
	ErrSignJWT = errors.New("failed to sign jwt token")
	// ErrValidateJWTToken indicates a failure to validate JWT token.
	ErrValidateJWTToken = errors.New("failed to validate jwt token")
	// ErrJSONHandle indicates an error in handling JSON.
	ErrJSONHandle = errors.New("failed to perform operation JSON")
	// errRevokedToken indicates that the token is revoked.
	errRevokedToken = errors.New("token is revoked")
)

const (
	issuerName             = "magistrala.auth"
	tokenType              = "type"
	userField              = "user"
	domainField            = "domain"
	oauthProviderField     = "oauth_provider"
	oauthAccessTokenField  = "access_token"
	oauthRefreshTokenField = "refresh_token"
)

var _ auth.Tokenizer = (*tokenizer)(nil)

type tokenizer struct {
	secret []byte
	cache  auth.Cache
	repo   auth.TokenRepository
}

// NewRepository instantiates an implementation of Token repository.
func New(secret []byte, repo auth.TokenRepository, cache auth.Cache) auth.Tokenizer {
	return &tokenizer{
		secret: secret,
		repo:   repo,
		cache:  cache,
	}
}

func (tok *tokenizer) Issue(key auth.Key) (string, error) {
	builder := jwt.NewBuilder()
	builder.
		Issuer(issuerName).
		IssuedAt(key.IssuedAt).
		Claim(tokenType, key.Type).
		Expiration(key.ExpiresAt)
	builder.Claim(userField, key.User)
	if key.Domain != "" {
		builder.Claim(domainField, key.Domain)
	}
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

func (tok *tokenizer) Parse(ctx context.Context, token string) (auth.Key, error) {
	tkn, err := tok.validateToken(token)
	if err != nil {
		return auth.Key{}, errors.Wrap(svcerr.ErrAuthentication, err)
	}

	key, err := toKey(tkn)
	if err != nil {
		return auth.Key{}, errors.Wrap(svcerr.ErrAuthentication, err)
	}

	if key.Type == auth.RefreshKey {
		switch tok.cache.Contains(ctx, "", key.ID) {
		case true:
			return auth.Key{}, errors.Wrap(svcerr.ErrAuthentication, errRevokedToken)
		default:
			if ok := tok.repo.Contains(ctx, key.ID); ok {
				if err := tok.cache.Save(ctx, "", key.ID); err != nil {
					return auth.Key{}, errors.Wrap(svcerr.ErrAuthentication, err)
				}

				return auth.Key{}, errors.Wrap(svcerr.ErrAuthentication, errRevokedToken)
			}
		}
	}

	return key, nil
}

func (tok *tokenizer) Revoke(ctx context.Context, token string) error {
	tkn, err := tok.validateToken(token)
	if err != nil {
		return errors.Wrap(svcerr.ErrAuthentication, err)
	}

	key, err := toKey(tkn)
	if err != nil {
		return errors.Wrap(svcerr.ErrAuthentication, err)
	}

	if key.Type == auth.RefreshKey {
		if err := tok.repo.Save(ctx, key.ID); err != nil {
			return errors.Wrap(svcerr.ErrAuthentication, err)
		}

		if err := tok.cache.Save(ctx, "", key.ID); err != nil {
			return errors.Wrap(svcerr.ErrAuthentication, err)
		}
	}

	return nil
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

func (tok *tokenizer) RetrieveJWKS() auth.JWKS {
	jwk := auth.JWK{
		Kty: "RSA",
		Kid: tok.keyID,
		N:   base64.RawURLEncoding.EncodeToString(tok.publicKey.N.Bytes()),
		E:   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(tok.publicKey.E)).Bytes()),
	}

	jwks := auth.JWKS{
		Keys: []auth.JWK{jwk},
	}
	return jwks
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
		return auth.Key{}, err
	}
	ktype, err := strconv.ParseInt(fmt.Sprintf("%v", tType), 10, 64)
	if err != nil {
		return auth.Key{}, err
	}

	key.ID = tkn.JwtID()
	key.Type = auth.KeyType(ktype)
	key.Issuer = tkn.Issuer()
	key.Subject = tkn.Subject()
	key.IssuedAt = tkn.IssuedAt()
	key.ExpiresAt = tkn.Expiration()

	return key, nil
}

func loadPrivateKey(filePath string) (*rsa.PrivateKey, error) {
	privKeyBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrap(errReadPrivateKeyFile, err)
	}

	block, _ := pem.Decode(privKeyBytes)
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, errors.Wrap(errReadPrivateKeyFile, errors.New("invalid file type"))
	}

	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.Wrap(errReadPrivateKeyFile, err)
	}

	rsaKey, ok := privKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.Wrap(errReadPrivateKeyFile, errors.New("invalid private key type"))
	}

	return rsaKey, nil
}
