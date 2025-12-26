// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package symmetric

import (
	"context"

	"github.com/absmach/supermq/auth"
	smqjwt "github.com/absmach/supermq/auth/tokenizer/util"
	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

const (
	patPrefix = "pat"
)

var errJWTExpiryKey = errors.New(`"exp" not satisfied`)

type tokenizer struct {
	algorithm jwa.KeyAlgorithm
	secret    []byte
}

var _ auth.Tokenizer = (*tokenizer)(nil)

func NewTokenizer(algorithm string, secret []byte) (auth.Tokenizer, error) {
	alg := jwa.KeyAlgorithmFrom(algorithm)
	if _, ok := alg.(jwa.InvalidKeyAlgorithm); ok {
		return nil, auth.ErrUnsupportedKeyAlgorithm
	}
	if len(secret) == 0 {
		return nil, auth.ErrInvalidSymmetricKey
	}
	return &tokenizer{
		secret:    secret,
		algorithm: alg,
	}, nil
}

func (km *tokenizer) Issue(key auth.Key) (string, error) {
	tkn, err := smqjwt.BuildToken(key)
	if err != nil {
		return "", err
	}

	signedBytes, err := jwt.Sign(tkn, jwt.WithKey(km.algorithm, km.secret))
	if err != nil {
		return "", err
	}

	return string(signedBytes), nil
}

func (km *tokenizer) Parse(ctx context.Context, tokenString string) (auth.Key, error) {
	if len(tokenString) >= 3 && tokenString[:3] == patPrefix {
		return auth.Key{Type: auth.PersonalAccessToken}, nil
	}

	tkn, err := jwt.Parse(
		[]byte(tokenString),
		jwt.WithValidate(true),
		jwt.WithKey(km.algorithm, km.secret),
	)
	if err != nil {
		if errors.Contains(err, errJWTExpiryKey) {
			return auth.Key{}, errors.Wrap(svcerr.ErrAuthentication, auth.ErrExpiry)
		}
		return auth.Key{}, errors.Wrap(svcerr.ErrAuthentication, err)
	}

	if tkn.Issuer() != smqjwt.IssuerName {
		return auth.Key{}, smqjwt.ErrInvalidIssuer
	}

	return smqjwt.ToKey(tkn)
}

func (km *tokenizer) RetrieveJWKS() ([]auth.PublicKeyInfo, error) {
	return nil, auth.ErrPublicKeysNotSupported
}
