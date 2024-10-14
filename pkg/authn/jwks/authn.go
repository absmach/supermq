// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package jwks

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"time"

	"github.com/absmach/magistrala/pkg/authn"
	"github.com/absmach/magistrala/pkg/errors"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

var (
	// errJWTExpiryKey is used to check if the token is expired.
	errJWTExpiryKey = errors.New(`"exp" not satisfied`)
	// errInvalidIssuer indicates an invalid issuer value.
	errInvalidIssuer = errors.New("invalid token issuer value")
	// ErrValidateJWTToken indicates a failure to validate JWT token.
	ErrValidateJWTToken = errors.New("failed to validate jwt token")

	jwksCache = struct {
		jwks     mgauth.JWKS
		cachedAt time.Time
	}{}
)
var _ authn.Authentication = (*authentication)(nil)

type authentication struct{}

func NewAuthentication() authn.Authentication {
	return authentication{}
}

func (a authentication) Authenticate(ctx context.Context, token string) (authn.Session, error) {
	jwks, err := fetchJWKS()
	if err != nil {
		return auth.Session{}, err
	}

	publicKey, err := createPublicKey(jwks.Keys[0])
	if err != nil {
		return auth.Session{}, err
	}

	tkn, err := validateToken(token, publicKey)
	if err != nil {
		return auth.Session{}, err
	}

	res := auth.Session{DomainUserID: tkn.Subject()}
	pc := tkn.PrivateClaims()
	if pc["user"] != nil {
		res.UserID = pc["user"].(string)
	}
	if pc["domain"] != nil {
		res.DomainID = pc["domain"].(string)
	}

	return res, nil
}

func fetchJWKS() (mgauth.JWKS, error) {
	req, err := http.NewRequest("GET", client.jwksURL, nil)
	if err != nil {
		return mgauth.JWKS{}, err
	}
	req.Header.Set("Accept", "application/json")

	httpClient := &http.Client{}
	if time.Since(jwksCache.cachedAt) < cacheDuration && jwksCache.jwks.Keys != nil {
		return jwksCache.jwks, nil
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return mgauth.JWKS{}, err
	}
	defer resp.Body.Close()

	var jwks mgauth.JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return mgauth.JWKS{}, err
	}
	jwksCache.jwks = jwks
	jwksCache.cachedAt = time.Now()

	return jwks, nil
}

func validateToken(token string, publicKey *rsa.PublicKey) (jwt.Token, error) {
	tkn, err := jwt.Parse(
		[]byte(token),
		jwt.WithValidate(true),
		jwt.WithKey(jwa.RS256, publicKey),
	)
	if err != nil {
		if errors.Contains(err, errJWTExpiryKey) {
			return nil, mgauth.ErrExpiry
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

func createPublicKey(jwk mgauth.JWK) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, err
	}

	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	publicKey := &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}

	return publicKey, nil
}
