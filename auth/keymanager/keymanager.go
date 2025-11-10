// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package keymanager

import (
	"crypto/rand"
	"crypto/rsa"
	"time"

	"github.com/absmach/supermq"
	"github.com/absmach/supermq/auth"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type keyPair struct {
	privateKey *rsa.PrivateKey
	jwkKey     jwk.Key
	createdAt  time.Time
}

type manager struct {
	idProvider supermq.IDProvider
	keySet     map[string]keyPair
	activeID   string
	retiredID  string
}

var _ auth.KeyManager = (*manager)(nil)

func NewKeyManager(idProvider supermq.IDProvider) (auth.KeyManager, error) {
	privateKey, publicKey, err := generateRSAKeyPair()
	if err != nil {
		return nil, err
	}
	kid, err := idProvider.ID()
	if err != nil {
		return nil, err
	}
	jwkKey, err := jwk.FromRaw(publicKey)
	if err != nil {
		return nil, err
	}
	jwkKey.Set(jwk.KeyIDKey, kid)
	jwkKey.Set(jwk.KeyTypeKey, "RSA")
	return &manager{
		idProvider: idProvider,
		keySet: map[string]keyPair{kid: {
			privateKey: privateKey,
			jwkKey:     jwkKey,
			createdAt:  time.Now(),
		}},
		activeID: kid,
	}, nil
}

func (km *manager) SignJWT(token jwt.Token) ([]byte, error) {
	privateKey := km.keySet[km.activeID].privateKey

	signedToken, err := jwt.Sign(token, jwt.WithKey(jwa.RS256, privateKey))
	if err != nil {
		return nil, err
	}
	return signedToken, nil
}

func (km *manager) ParseJWT(token string) (jwt.Token, error) {
	publicKey := km.keySet[km.activeID].privateKey.Public().(*rsa.PublicKey)

	tkn, err := jwt.Parse(
		[]byte(token),
		jwt.WithValidate(true),
		jwt.WithKey(jwa.RS256, publicKey),
	)
	if err != nil {
		return nil, err
	}
	return tkn, nil
}

func (km *manager) PublicJWKS() jwk.Set {
	set := jwk.NewSet()
	set.AddKey(km.keySet[km.activeID].jwkKey)
	if km.retiredID != "" {
		set.AddKey(km.keySet[km.retiredID].jwkKey)
	}
	return set
}

func (km *manager) Rotate() error {
	privateKey, publicKey, err := generateRSAKeyPair()
	if err != nil {
		return err
	}
	currentID := km.activeID
	newID, err := km.idProvider.ID()
	if err != nil {
		return err
	}
	jwkKey, err := jwk.FromRaw(publicKey)
	if err != nil {
		return err
	}
	jwkKey.Set(jwk.KeyIDKey, newID)
	jwkKey.Set(jwk.KeyTypeKey, "RSA")

	km.keySet[newID] = keyPair{
		privateKey: privateKey,
		jwkKey:     jwkKey,
		createdAt:  time.Now(),
	}
	km.retiredID = currentID
	km.activeID = newID

	return nil
}

func generateRSAKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	publicKey := &privateKey.PublicKey

	return privateKey, publicKey, nil
}
