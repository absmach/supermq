// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package jwks

import (
	"context"
	"io"
	"net/http"
	"time"

	smqauth "github.com/absmach/supermq/auth"
	smqjwt "github.com/absmach/supermq/auth/jwt"
	"github.com/absmach/supermq/pkg/authn"
	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

const (
	issuerName    = "supermq.auth"
	cacheDuration = 5 * time.Minute
)

var (
	// errJWTExpiryKey is used to check if the token is expired.
	errJWTExpiryKey = errors.New(`"exp" not satisfied`)
	// errFetchJWKS indicates an error fetching JWKS from URL.
	errFetchJWKS = errors.New("failed to fetch jwks")
	// errInvalidIssuer indicates an invalid issuer value.
	errInvalidIssuer = errors.New("invalid token issuer value")
	// ErrValidateJWTToken indicates a failure to validate JWT token.
	errValidateJWTToken = errors.New("failed to validate jwt token")

	jwksCache = struct {
		jwks     jwk.Set
		cachedAt time.Time
	}{}
)

var _ authn.Authentication = (*authentication)(nil)

type authentication struct {
	jwksURL string
}

func NewAuthentication(jwksURL string) authn.Authentication {
	return authentication{
		jwksURL: jwksURL,
	}
}

func (a authentication) Authenticate(ctx context.Context, token string) (authn.Session, error) {
	jwks, err := a.fetchJWKS()
	if err != nil {
		return authn.Session{}, errors.Wrap(svcerr.ErrAuthentication, err)
	}
	tkn, err := validateToken(token, jwks)
	if err != nil {
		return authn.Session{}, errors.Wrap(svcerr.ErrAuthentication, err)
	}
	key, err := smqjwt.ToKey(tkn)
	if err != nil {
		return authn.Session{}, errors.Wrap(svcerr.ErrAuthentication, err)
	}

	return authn.Session{
		Type:     authn.AccessToken,
		UserID:   key.Subject,
		Role:     authn.Role(key.Role),
		Verified: key.Verified,
	}, nil
}

func (a authentication) fetchJWKS() (jwk.Set, error) {
	req, err := http.NewRequest("GET", a.jwksURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	httpClient := &http.Client{}
	if time.Since(jwksCache.cachedAt) < cacheDuration && jwksCache.jwks.Len() > 0 {
		return jwksCache.jwks, nil
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errFetchJWKS
	}

	data, _ := io.ReadAll(resp.Body)
	set, err := jwk.Parse(data)
	if err != nil {
		return nil, err
	}
	jwksCache.jwks = set
	jwksCache.cachedAt = time.Now()

	return set, nil
}

func validateToken(token string, jwks jwk.Set) (jwt.Token, error) {
	tkn, err := jwt.Parse(
		[]byte(token),
		jwt.WithValidate(true),
		jwt.WithKeySet(jwks, jws.WithInferAlgorithmFromKey(true)),
	)
	if err != nil {
		if errors.Contains(err, errJWTExpiryKey) {
			return nil, smqauth.ErrExpiry
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
		return nil, errors.Wrap(errValidateJWTToken, err)
	}

	return tkn, nil
}
