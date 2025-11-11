// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package keymanager

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"time"

	"github.com/absmach/supermq"
	"github.com/absmach/supermq/auth"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

const (
	rotationInterval = 24 * time.Second
	gracePeriod      = 26 * time.Minute // token_ttl + skew + jwks_max_age
)

type keyPair struct {
	privateKey jwk.Key
	publicKey  jwk.Key
	retiredAt  time.Time
}

type manager struct {
	idProvider supermq.IDProvider
	keySet     map[string]keyPair
	activeID   string
	retiredID  string
}

var _ auth.KeyManager = (*manager)(nil)

func NewKeyManager(ctx context.Context, idProvider supermq.IDProvider) (auth.KeyManager, error) {
	kid, err := idProvider.ID()
	if err != nil {
		return nil, err
	}
	privateJwk, publicJwk, err := generateKeyPair(kid)
	if err != nil {
		return nil, err
	}

	km := &manager{
		idProvider: idProvider,
		keySet: map[string]keyPair{kid: {
			privateKey: privateJwk,
			publicKey:  publicJwk,
		}},
		activeID: kid,
	}

	go km.rotateHandler(ctx)

	return km, nil
}

func (km *manager) SignJWT(token jwt.Token) ([]byte, error) {
	jwkKey := km.keySet[km.activeID].privateKey

	return jwt.Sign(token, jwt.WithKey(jwa.RS256, jwkKey))
}

func (km *manager) ParseJWT(token string) (jwt.Token, error) {
	set := jwk.NewSet()
	set.AddKey(km.keySet[km.activeID].publicKey)
	if km.retiredID != "" {
		// Check if the retired key is still within the grace period
		if time.Since(km.keySet[km.retiredID].retiredAt) <= gracePeriod {
			set.AddKey(km.keySet[km.retiredID].publicKey)
		}
	}
	tkn, err := jwt.Parse(
		[]byte(token),
		jwt.WithValidate(true),
		jwt.WithKeySet(set, jws.WithInferAlgorithmFromKey(true)),
	)
	if err != nil {
		return nil, err
	}
	return tkn, nil
}

func (km *manager) PublicJWKS() []jwk.Key {
	keys := []jwk.Key{km.keySet[km.activeID].publicKey}
	if km.retiredID != "" {
		// Check if the retired key is still within the grace period
		if time.Since(km.keySet[km.retiredID].retiredAt) > gracePeriod {
			delete(km.keySet, km.retiredID)
			return keys
		}
		keys = append(keys, km.keySet[km.retiredID].publicKey)
	}
	return keys
}

func (km *manager) Rotate() error {
	currentID := km.activeID
	newID, err := km.idProvider.ID()
	if err != nil {
		return err
	}
	privateJwk, publicJwk, err := generateKeyPair(newID)
	if err != nil {
		return err
	}
	km.keySet[newID] = keyPair{
		privateKey: privateJwk,
		publicKey:  publicJwk,
	}
	km.retiredID = currentID
	km.activeID = newID

	if currentID != "" {
		kp, ok := km.keySet[currentID]
		if ok {
			t := time.Now().UTC()
			kp.retiredAt = t
			km.keySet[currentID] = kp
			km.retiredID = currentID
		}
	}

	return nil
}

func (km *manager) rotateHandler(ctx context.Context) error {
	ticker := time.NewTicker(rotationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := km.Rotate(); err != nil {
				return err
			}
		}
	}
}

func generateKeyPair(kid string) (jwk.Key, jwk.Key, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	publicKey := &privateKey.PublicKey

	privateJwk, err := jwk.FromRaw(privateKey)
	if err != nil {
		return nil, nil, err
	}
	privateJwk.Set(jwk.KeyIDKey, kid)

	publicJwk, err := jwk.FromRaw(publicKey)
	if err != nil {
		return nil, nil, err
	}
	publicJwk.Set(jwk.KeyIDKey, kid)
	publicJwk.Set(jwk.KeyTypeKey, "RSA")

	return privateJwk, publicJwk, nil
}
