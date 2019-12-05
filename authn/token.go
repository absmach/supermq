// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package authn

import (
	"github.com/dgrijalva/jwt-go"
)

type claims struct {
	jwt.StandardClaims
	Type *uint32 `json:"type,omitempty"`
}

func (svc authService) issue(key Key) (string, error) {
	claims := claims{
		StandardClaims: jwt.StandardClaims{
			Issuer:   key.Issuer,
			Subject:  key.Secret,
			IssuedAt: key.IssuedAt.Unix(),
		},
		Type: &key.Type,
	}

	if key.ExpiresAt != nil {
		claims.ExpiresAt = key.ExpiresAt.Unix()
	}
	if key.ID != "" {
		claims.Id = key.ID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(svc.secret))
}

func (svc authService) parse(token string) (claims, error) {
	c := claims{}
	_, err := jwt.ParseWithClaims(token, &c, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnauthorizedAccess
		}
		return []byte(svc.secret), nil
	})

	if err != nil {
		if e, ok := err.(*jwt.ValidationError); ok && e.Errors == jwt.ValidationErrorExpired {
			if c.Type != nil && *c.Type == UserKey {
				return c, nil
			}
			return claims{}, ErrKeyExpired
		}
		return claims{}, ErrUnauthorizedAccess
	}

	return c, nil
}
