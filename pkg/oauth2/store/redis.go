// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/absmach/supermq/pkg/oauth2"
	"github.com/redis/go-redis/v9"
)

const (
	deviceCodePrefix = "oauth:device:code:"
	userCodePrefix   = "oauth:device:user:"
)

// redisDeviceCodeStore is a Redis-based implementation of DeviceCodeStore.
type redisDeviceCodeStore struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisDeviceCodeStore creates a new Redis-based device code store.
func NewRedisDeviceCodeStore(ctx context.Context, client *redis.Client) oauth2.DeviceCodeStore {
	return &redisDeviceCodeStore{
		client: client,
		ctx:    ctx,
	}
}

func (s *redisDeviceCodeStore) Save(code oauth2.DeviceCode) error {
	data, err := json.Marshal(code)
	if err != nil {
		return fmt.Errorf("failed to marshal device code: %w", err)
	}

	// Store device code with expiry
	deviceKey := deviceCodePrefix + code.DeviceCode
	if err := s.client.Set(s.ctx, deviceKey, data, oauth2.DeviceCodeExpiry).Err(); err != nil {
		return fmt.Errorf("failed to save device code: %w", err)
	}

	// Store user code to device code mapping with expiry
	userKey := userCodePrefix + code.UserCode
	if err := s.client.Set(s.ctx, userKey, code.DeviceCode, oauth2.DeviceCodeExpiry).Err(); err != nil {
		return fmt.Errorf("failed to save user code mapping: %w", err)
	}

	return nil
}

func (s *redisDeviceCodeStore) Get(deviceCode string) (oauth2.DeviceCode, error) {
	deviceKey := deviceCodePrefix + deviceCode
	data, err := s.client.Get(s.ctx, deviceKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return oauth2.DeviceCode{}, oauth2.ErrDeviceCodeNotFound
		}
		return oauth2.DeviceCode{}, fmt.Errorf("failed to get device code: %w", err)
	}

	var code oauth2.DeviceCode
	if err := json.Unmarshal(data, &code); err != nil {
		return oauth2.DeviceCode{}, fmt.Errorf("failed to unmarshal device code: %w", err)
	}

	return code, nil
}

func (s *redisDeviceCodeStore) GetByUserCode(userCode string) (oauth2.DeviceCode, error) {
	// First, get the device code from user code mapping
	userKey := userCodePrefix + userCode
	deviceCode, err := s.client.Get(s.ctx, userKey).Result()
	if err != nil {
		if err == redis.Nil {
			return oauth2.DeviceCode{}, oauth2.ErrUserCodeNotFound
		}
		return oauth2.DeviceCode{}, fmt.Errorf("failed to get device code by user code: %w", err)
	}

	// Then, get the actual device code data
	return s.Get(deviceCode)
}

func (s *redisDeviceCodeStore) Update(code oauth2.DeviceCode) error {
	// Get the existing code to check if it exists
	existing, err := s.Get(code.DeviceCode)
	if err != nil {
		return err
	}

	// Preserve the creation time and user code from existing
	code.CreatedAt = existing.CreatedAt
	code.UserCode = existing.UserCode

	data, err := json.Marshal(code)
	if err != nil {
		return fmt.Errorf("failed to marshal device code: %w", err)
	}

	// Calculate remaining TTL
	deviceKey := deviceCodePrefix + code.DeviceCode
	ttl, err := s.client.TTL(s.ctx, deviceKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get TTL: %w", err)
	}

	// If TTL is negative (key doesn't exist or no expiry), use default
	if ttl < 0 {
		ttl = oauth2.DeviceCodeExpiry
	}

	// Update the device code with remaining TTL
	if err := s.client.Set(s.ctx, deviceKey, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to update device code: %w", err)
	}

	return nil
}

func (s *redisDeviceCodeStore) Delete(deviceCode string) error {
	// Get the code first to find the user code
	code, err := s.Get(deviceCode)
	if err != nil {
		return err
	}

	// Delete both device code and user code mapping
	deviceKey := deviceCodePrefix + deviceCode
	userKey := userCodePrefix + code.UserCode

	pipe := s.client.Pipeline()
	pipe.Del(s.ctx, deviceKey)
	pipe.Del(s.ctx, userKey)

	if _, err := pipe.Exec(s.ctx); err != nil {
		return fmt.Errorf("failed to delete device code: %w", err)
	}

	return nil
}
