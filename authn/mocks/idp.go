// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"fmt"
	"sync"

	"github.com/mainflux/mainflux"
)

var _ mainflux.UUIDProvider = (*uuidProviderMock)(nil)

type uuidProviderMock struct {
	mu      sync.Mutex
	counter int
}

func (idp *uuidProviderMock) ID() (string, error) {
	idp.mu.Lock()
	defer idp.mu.Unlock()

	idp.counter++
	return fmt.Sprintf("%s%012d", "123e4567-e89b-12d3-a456-", idp.counter), nil
}

// NewUUIDProvider creates "mirror" uuid provider, i.e. generated
// token will hold value provided by the caller.
func NewUUIDProvider() mainflux.UUIDProvider {
	return &uuidProviderMock{}
}
