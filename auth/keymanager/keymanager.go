// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package keymanager

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/absmach/supermq"
	"github.com/absmach/supermq/auth"
	"github.com/absmach/supermq/pkg/errors"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

const (
	skewDuration = 15 * time.Second
	keysDir      = "keys"
)

var (
	errEmptyKerDir = errors.New("key directory cannot be empty when saving to file")
)

type keyPair struct {
	privateKey jwk.Key
	publicKey  jwk.Key
	retiredAt  time.Time
}

type manager struct {
	mu               sync.RWMutex
	idProvider       supermq.IDProvider
	keySet           map[string]keyPair
	activeID         string
	nextID           string
	retiredID        string
	gracePeriod      time.Duration
	rotationInterval time.Duration
	saveToFile       bool
	keyDir           string
	statePath        string
}

type KeyManagerConfig struct {
	RotationInterval time.Duration `env:"ROTATION_INTERVAL" envDefault:"24h"`
	SaveToFile       bool          `env:"SAVE_TO_FILE"      envDefault:"false"`
	LoginDuration    time.Duration
}

type diskKey struct {
	Private   json.RawMessage `json:"private_jwk"`
	Public    json.RawMessage `json:"public_jwk"`
	RetiredAt string          `json:"retired_at,omitempty"`
}

type diskState struct {
	ActiveID  string             `json:"active_id"`
	NextID    string             `json:"next_id"`
	RetiredID string             `json:"retired_id,omitempty"`
	Keys      map[string]diskKey `json:"keys"`
}

var _ auth.KeyManager = (*manager)(nil)

func NewKeyManager(ctx context.Context, cfg KeyManagerConfig, idProvider supermq.IDProvider) (auth.KeyManager, error) {
	km := &manager{
		idProvider:       idProvider,
		keySet:           make(map[string]keyPair),
		gracePeriod:      cfg.LoginDuration + skewDuration,
		rotationInterval: cfg.RotationInterval,
		saveToFile:       cfg.SaveToFile,
		keyDir:           keysDir,
	}

	if km.saveToFile {
		if km.keyDir == "" {
			return nil, errEmptyKerDir
		}
		if err := os.MkdirAll(km.keyDir, 0o700); err != nil {
			return nil, err
		}
		km.statePath = filepath.Join(km.keyDir, "keys.json")
		// Try to load existing state
		if exists(km.statePath) {
			if err := km.loadFromDisk(); err != nil {
				return nil, err
			}
		}
	}

	if km.activeID == "" || km.nextID == "" || len(km.keySet) == 0 {
		activeKid, err := idProvider.ID()
		if err != nil {
			return nil, err
		}
		activePair, err := generateKeyPair(activeKid)
		if err != nil {
			return nil, err
		}
		nextKid, err := idProvider.ID()
		if err != nil {
			return nil, err
		}
		nextPair, err := generateKeyPair(nextKid)
		if err != nil {
			return nil, err
		}
		km.keySet[activeKid] = activePair
		km.keySet[nextKid] = nextPair
		km.activeID = activeKid
		km.nextID = nextKid

		if km.saveToFile {
			if err := km.saveToDisk(); err != nil {
				return nil, err
			}
		}
	}

	// Start rotation loop if interval > 0
	if km.rotationInterval > 0 {
		go km.rotateHandler(ctx)
	}

	return km, nil
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
	set.AddKey(km.keySet[km.activeID].publicKey)
	if km.retiredID != "" {
		// Check if the retired key is still within the grace period
		if time.Since(km.keySet[km.retiredID].retiredAt) <= km.gracePeriod {
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
	km.mu.Lock()
	defer km.mu.Unlock()

	keys := []jwk.Key{km.keySet[km.activeID].publicKey}
	if km.retiredID != "" {
		kp := km.keySet[km.retiredID]
		if time.Since(kp.retiredAt) > km.gracePeriod {
			delete(km.keySet, km.retiredID)
			km.retiredID = ""
			// persist state change if needed
			if km.saveToFile {
				_ = km.saveToDisk()
			}
			return keys
		}
		keys = append(keys, kp.publicKey)
	}

	if km.nextID != "" {
		keys = append(keys, km.keySet[km.nextID].publicKey)
	}
	return keys
}

func (km *manager) Rotate() error {
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
		}
	}

	if km.saveToFile {
		if err := km.saveToDisk(); err != nil {
			return err
		}
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
			if err := km.Rotate(); err != nil {
				return err
			}
		}
	}
}

func (km *manager) loadFromDisk() error {
	b, err := os.ReadFile(km.statePath)
	if err != nil {
		return err
	}
	var ds diskState
	if err := json.Unmarshal(b, &ds); err != nil {
		return err
	}

	keys := make(map[string]keyPair, len(ds.Keys))
	for kid, dk := range ds.Keys {
		priv, err := jwk.ParseKey(dk.Private)
		if err != nil {
			return err
		}
		pub, err := jwk.ParseKey(dk.Public)
		if err != nil {
			return err
		}
		var rt time.Time
		if dk.RetiredAt != "" {
			rt, err = time.Parse(time.RFC3339, dk.RetiredAt)
			if err != nil {
				return err
			}
		}
		keys[kid] = keyPair{privateKey: priv, publicKey: pub, retiredAt: rt}
	}

	km.keySet = keys
	km.activeID = ds.ActiveID
	km.nextID = ds.NextID
	km.retiredID = ds.RetiredID
	return nil
}

func (km *manager) saveToDisk() error {
	ds := diskState{
		ActiveID:  km.activeID,
		NextID:    km.nextID,
		RetiredID: km.retiredID,
		Keys:      make(map[string]diskKey, len(km.keySet)),
	}

	for kid, kp := range km.keySet {
		pb, err := json.Marshal(kp.publicKey)
		if err != nil {
			return err
		}
		sb, err := json.Marshal(kp.privateKey)
		if err != nil {
			return err
		}
		dk := diskKey{
			Private: sb,
			Public:  pb,
		}
		if !kp.retiredAt.IsZero() {
			dk.RetiredAt = kp.retiredAt.Format(time.RFC3339)
		}
		ds.Keys[kid] = dk
	}

	data, err := json.MarshalIndent(ds, "", "  ")
	if err != nil {
		return err
	}
	tmp := km.statePath + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, km.statePath)
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
	privateJwk.Set(jwk.KeyIDKey, kid)

	publicJwk, err := jwk.FromRaw(publicKey)
	if err != nil {
		return keyPair{}, err
	}
	publicJwk.Set(jwk.KeyIDKey, kid)
	publicJwk.Set(jwk.KeyTypeKey, "RSA")

	return keyPair{
		privateKey: privateJwk,
		publicKey:  publicJwk,
	}, nil
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
