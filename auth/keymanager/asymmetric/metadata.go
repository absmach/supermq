// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package asymmetric

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

const (
	defaultMetadataFile = "keys.json"
	defaultGracePeriod  = 168 * time.Hour // 7 days
)

var (
	errNoActiveKey       = errors.New("no active key found in metadata")
	errInvalidKeyStatus  = errors.New("invalid key status")
	errNoActiveKeyLoaded = errors.New("active key not loaded successfully")
	errLoadingMetadata   = errors.New("failed to load key metadata")
	errParsingMetadata   = errors.New("failed to parse key metadata")
	errNoMetadata        = errors.New("no metadata file in private key dir")
)

// KeyStatus represents the lifecycle state of a key.
type KeyStatus string

const (
	// KeyStatusActive indicates the key is currently used for signing and verification.
	KeyStatusActive KeyStatus = "active"
	// KeyStatusRetiring indicates the key is only used for verification (within grace period).
	KeyStatusRetiring KeyStatus = "retiring"
	// KeyStatusRetired indicates the key is no longer valid (beyond grace period).
	KeyStatusRetired KeyStatus = "retired"
)

// KeyMetadataEntry represents a single key's metadata.
type KeyMetadataEntry struct {
	ID        string    `json:"id"`
	File      string    `json:"file"`
	CreatedAt time.Time `json:"created_at"`
	Status    KeyStatus `json:"status"`
	ExpiresAt time.Time `json:"expires_at,omitempty"` // Set when status changes to retiring
}

// KeysMetadata represents the complete key configuration.
type KeysMetadata struct {
	ActiveKeyID      string             `json:"active_key_id"`
	GracePeriodHours int                `json:"grace_period_hours"`
	Keys             []KeyMetadataEntry `json:"keys"`
}

// IsExpired checks if the key is beyond its expiration time.
func (k *KeyMetadataEntry) IsExpired() bool {
	if k.Status != KeyStatusRetiring {
		return false
	}
	if k.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(k.ExpiresAt)
}

// IsValid checks if the key is active or retiring (not expired).
func (k *KeyMetadataEntry) IsValid() bool {
	if k.IsExpired() {
		return false
	}
	switch k.Status {
	case KeyStatusActive, KeyStatusRetiring:
		return true
	case KeyStatusRetired:
		return false
	default:
		return false
	}
}

// LoadKeysMetadata loads the keys metadata from a JSON file.
func LoadKeysMetadata(keysDir string) (*KeysMetadata, error) {
	metadataPath := filepath.Join(keysDir, defaultMetadataFile)

	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		return nil, errNoMetadata
	}
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, errors.Join(errLoadingMetadata, err)
	}

	var metadata KeysMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, errors.Join(errParsingMetadata, err)
	}

	// Validate metadata
	if err := metadata.Validate(); err != nil {
		return nil, err
	}

	// Set default grace period if not specified
	if metadata.GracePeriodHours == 0 {
		metadata.GracePeriodHours = int(defaultGracePeriod.Hours())
	}

	return &metadata, nil
}

// Validate checks if the metadata is valid.
func (m *KeysMetadata) Validate() error {
	if m.ActiveKeyID == "" {
		return errNoActiveKey
	}

	// Check that active key exists
	activeKeyExists := false
	for _, key := range m.Keys {
		if key.ID == m.ActiveKeyID {
			activeKeyExists = true
			if key.Status != KeyStatusActive {
				return errors.New("active_key_id must reference a key with 'active' status")
			}
		}

		// Validate key status
		if key.Status != KeyStatusActive &&
			key.Status != KeyStatusRetiring &&
			key.Status != KeyStatusRetired {
			return errInvalidKeyStatus
		}
	}

	if !activeKeyExists {
		return errors.New("active_key_id references non-existent key")
	}

	return nil
}

// GetActiveKey returns the active key metadata.
func (m *KeysMetadata) GetActiveKey() *KeyMetadataEntry {
	for i := range m.Keys {
		if m.Keys[i].ID == m.ActiveKeyID {
			return &m.Keys[i]
		}
	}
	return nil
}

// GetValidKeys returns all keys that are valid (active or retiring, not expired).
func (m *KeysMetadata) GetValidKeys() []KeyMetadataEntry {
	var validKeys []KeyMetadataEntry
	for _, key := range m.Keys {
		if key.IsValid() {
			validKeys = append(validKeys, key)
		}
	}
	return validKeys
}
