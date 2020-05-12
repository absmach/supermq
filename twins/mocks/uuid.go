// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"fmt"
	"strings"
	"sync"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/twins"
)

const u4Pref = "123e4567-e89b-12d3-a456-"

var _ mainflux.UUIDProvider = (*uuidProviderMock)(nil)

type uuidProviderMock struct {
	mu      sync.Mutex
	counter int
}

func (up *uuidProviderMock) ID() (string, error) {
	up.mu.Lock()
	defer up.mu.Unlock()

	up.counter++
	return fmt.Sprintf("%s%012d", u4Pref, up.counter), nil
}

func (up *uuidProviderMock) IsValid(u4 string) error {
	if !strings.Contains(u4Pref, u4) {
		return twins.ErrMalformedEntity
	}

	return nil
}

// NewUUIDProvider creates "mirror" uuid provider, i.e. generated
// token will hold value provided by the caller.
func NewUUIDProvider() mainflux.UUIDProvider {
	return &uuidProviderMock{}
}
