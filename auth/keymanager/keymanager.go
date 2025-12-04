// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package keymanager

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"sync"
	"time"

	"github.com/absmach/supermq"
	"github.com/absmach/supermq/auth"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

const (
	skewDuration    = 15 * time.Second
	cleanupInterval = 10 * time.Minute // How often to clean up expired keys
)

type keyPair struct {
	privateKey jwk.Key
	publicKey  jwk.Key
	retiredAt  time.Time
}

type manager struct {
	mu               sync.RWMutex
	idProvider       supermq.IDProvider
	repo             auth.PublicKeyRepository
	keySet           map[string]keyPair
	activeID         string
	nextID           string
	retiredID        string
	gracePeriod      time.Duration
	rotationInterval time.Duration
}

type KeyManagerConfig struct {
	RotationInterval time.Duration `env:"ROTATION_INTERVAL" envDefault:"24h"`
	LoginDuration    time.Duration
}

var _ auth.KeyManager = (*manager)(nil)

func NewKeyManager(ctx context.Context, cfg KeyManagerConfig, idProvider supermq.IDProvider, repo auth.PublicKeyRepository) (auth.KeyManager, error) {
	km := &manager{
		idProvider:       idProvider,
		repo:             repo,
		keySet:           make(map[string]keyPair),
		gracePeriod:      cfg.LoginDuration + skewDuration,
		rotationInterval: cfg.RotationInterval,
	}

	if err := km.loadFromDatabase(ctx); err != nil {
		return nil, err
	}

	if km.activeID == "" || km.nextID == "" || len(km.keySet) == 0 {
		if err := km.initializeKeys(ctx); err != nil {
			return nil, err
		}
	}

	if km.rotationInterval > 0 {
		go func() {
			if err := km.rotateHandler(ctx); err != nil {
				return
			}
		}()
	}

	go func() {
		if err := km.cleanupHandler(ctx); err != nil {
			return
		}
	}()

	return km, nil
}

func (km *manager) loadFromDatabase(ctx context.Context) error {
	activeKeys, err := km.repo.RetrieveActive(ctx)
	if err != nil {
		return nil
	}

	for _, key := range activeKeys {
		km.keySet[key.Kid] = keyPair{
			publicKey: key.JWKData,
			retiredAt: time.Time{},
		}

		if km.activeID == "" {
			km.activeID = key.Kid
		} else if km.nextID == "" {
			km.nextID = key.Kid
		}
	}

	return nil
}

func (km *manager) initializeKeys(ctx context.Context) error {
	activeKid, err := km.idProvider.ID()
	if err != nil {
		return err
	}
	activePair, err := generateKeyPair(activeKid)
	if err != nil {
		return err
	}

	nextKid, err := km.idProvider.ID()
	if err != nil {
		return err
	}
	nextPair, err := generateKeyPair(nextKid)
	if err != nil {
		return err
	}

	km.keySet[activeKid] = activePair
	km.keySet[nextKid] = nextPair
	km.activeID = activeKid
	km.nextID = nextKid

	now := time.Now().UTC()
	activePublicKey := auth.PublicKey{
		Kid:       activeKid,
		JWKData:   activePair.publicKey,
		CreatedAt: now,
		Status:    auth.ActiveKeyStatus,
	}
	if err := km.repo.Save(ctx, activePublicKey); err != nil {
		return err
	}

	nextPublicKey := auth.PublicKey{
		Kid:       nextKid,
		JWKData:   nextPair.publicKey,
		CreatedAt: now,
		Status:    auth.ActiveKeyStatus,
	}
	if err := km.repo.Save(ctx, nextPublicKey); err != nil {
		return err
	}

	return nil
}

func (km *manager) SignJWT(token jwt.Token) ([]byte, error) {
	km.mu.RLock()
	jwkKey := km.keySet[km.activeID].privateKey
	km.mu.RUnlock()

	return jwt.Sign(token, jwt.WithKey(jwa.RS256, jwkKey))
}

func (km *manager) ParseJWT(token string) (jwt.Token, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	set := jwk.NewSet()
	if err := set.AddKey(km.keySet[km.activeID].publicKey); err != nil {
		return nil, err
	}
	if km.retiredID != "" {
		if time.Since(km.keySet[km.retiredID].retiredAt) <= km.gracePeriod {
			if err := set.AddKey(km.keySet[km.retiredID].publicKey); err != nil {
				return nil, err
			}
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
	km.mu.Lock()
	defer km.mu.Unlock()

	keys := []jwk.Key{km.keySet[km.activeID].publicKey}
	if km.retiredID != "" {
		kp := km.keySet[km.retiredID]
		if time.Since(kp.retiredAt) > km.gracePeriod {
			delete(km.keySet, km.retiredID)
			km.retiredID = ""
			return keys
		}
		keys = append(keys, kp.publicKey)
	}

	if km.nextID != "" {
		keys = append(keys, km.keySet[km.nextID].publicKey)
	}
	return keys
}

func (km *manager) Rotate(ctx context.Context) error {
	km.mu.Lock()
	defer km.mu.Unlock()

	currentID := km.activeID
	nextID := km.nextID

	newID, err := km.idProvider.ID()
	if err != nil {
		return err
	}
	newPair, err := generateKeyPair(newID)
	if err != nil {
		return err
	}

	km.keySet[newID] = newPair
	km.activeID = nextID
	km.nextID = newID

	if currentID != "" {
		kp, ok := km.keySet[currentID]
		if ok {
			t := time.Now().UTC()
			kp.retiredAt = t
			km.keySet[currentID] = kp
			km.retiredID = currentID

			if err := km.repo.Retire(ctx, currentID); err != nil {
				return err
			}
		}
	}

	newPublicKey := auth.PublicKey{
		Kid:       newID,
		JWKData:   newPair.publicKey,
		CreatedAt: time.Now().UTC(),
		Status:    auth.ActiveKeyStatus,
	}
	if err := km.repo.Save(ctx, newPublicKey); err != nil {
		return err
	}

	return nil
}

func (km *manager) rotateHandler(ctx context.Context) error {
	ticker := time.NewTicker(km.rotationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := km.Rotate(ctx); err != nil {
				return err
			}
		}
	}
}

func (km *manager) cleanupHandler(ctx context.Context) error {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			expiredBefore := time.Now().UTC().Add(-km.gracePeriod)
			if err := km.repo.PurgeExpired(ctx, expiredBefore); err != nil {
				continue
			}
		}
	}
}

func generateKeyPair(kid string) (keyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return keyPair{}, err
	}
	publicKey := &privateKey.PublicKey

	privateJwk, err := jwk.FromRaw(privateKey)
	if err != nil {
		return keyPair{}, err
	}
	if err := privateJwk.Set(jwk.KeyIDKey, kid); err != nil {
		return keyPair{}, err
	}

	publicJwk, err := jwk.FromRaw(publicKey)
	if err != nil {
		return keyPair{}, err
	}
	if err := publicJwk.Set(jwk.KeyIDKey, kid); err != nil {
		return keyPair{}, err
	}
	if err := publicJwk.Set(jwk.KeyTypeKey, "RSA"); err != nil {
		return keyPair{}, err
	}

	return keyPair{
		privateKey: privateJwk,
		publicKey:  publicJwk,
	}, nil
}
