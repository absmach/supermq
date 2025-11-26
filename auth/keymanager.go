// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

// KeyManager is the high-level contract the Auth service depends on.
type KeyManager interface {
	SignJWT(token jwt.Token) ([]byte, error)

	ParseJWT(token string) (jwt.Token, error)

	PublicJWKS() []jwk.Key

	Rotate() error
}
