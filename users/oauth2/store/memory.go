// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"sync"
	"time"
)

// inMemoryDeviceCodeStore is an in-memory implementation of DeviceCodeStore.
type inMemoryDeviceCodeStore struct {
	mu          sync.RWMutex
	codes       map[string]DeviceCode
	userCodes   map[string]string // maps user code to device code
	cleanupDone chan struct{}
}

// NewInMemoryDeviceCodeStore creates a new in-memory device code store.
// It automatically starts a cleanup goroutine to remove expired codes.
func NewInMemoryDeviceCodeStore() DeviceCodeStore {
	store := &inMemoryDeviceCodeStore{
		codes:       make(map[string]DeviceCode),
		userCodes:   make(map[string]string),
		cleanupDone: make(chan struct{}),
	}
	go store.cleanup()
	return store
}

func (s *inMemoryDeviceCodeStore) Save(code DeviceCode) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.codes[code.DeviceCode] = code
	s.userCodes[code.UserCode] = code.DeviceCode
	return nil
}

func (s *inMemoryDeviceCodeStore) Get(deviceCode string) (DeviceCode, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	code, ok := s.codes[deviceCode]
	if !ok {
		return DeviceCode{}, ErrDeviceCodeNotFound
	}
	return code, nil
}

func (s *inMemoryDeviceCodeStore) GetByUserCode(userCode string) (DeviceCode, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	deviceCode, ok := s.userCodes[userCode]
	if !ok {
		return DeviceCode{}, ErrUserCodeNotFound
	}
	code, ok := s.codes[deviceCode]
	if !ok {
		return DeviceCode{}, ErrDeviceCodeNotFound
	}
	return code, nil
}

func (s *inMemoryDeviceCodeStore) Update(code DeviceCode) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.codes[code.DeviceCode]; !ok {
		return ErrDeviceCodeNotFound
	}
	s.codes[code.DeviceCode] = code
	return nil
}

func (s *inMemoryDeviceCodeStore) Delete(deviceCode string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if code, ok := s.codes[deviceCode]; ok {
		delete(s.userCodes, code.UserCode)
	}
	delete(s.codes, deviceCode)
	return nil
}

// cleanup periodically removes expired device codes.
func (s *inMemoryDeviceCodeStore) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			now := time.Now()
			for deviceCode, code := range s.codes {
				if now.Sub(code.CreatedAt) > DeviceCodeExpiry {
					delete(s.codes, deviceCode)
					delete(s.userCodes, code.UserCode)
				}
			}
			s.mu.Unlock()
		case <-s.cleanupDone:
			return
		}
	}
}
