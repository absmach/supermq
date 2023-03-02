package mocks

import (
	"context"
	"fmt"
	"sync"

	"github.com/mainflux/mainflux/things/policies"
)

type channelCacheMock struct {
	mu       sync.Mutex
	policies map[string]policies.Policy
}

// NewChannelCache returns mock cache instance.
func NewChannelCache() policies.Cache {
	return &channelCacheMock{
		policies: make(map[string]policies.Policy),
	}
}

func (ccm *channelCacheMock) AddPolicy(_ context.Context, policy policies.Policy) error {
	ccm.mu.Lock()
	defer ccm.mu.Unlock()

	ccm.policies[fmt.Sprintf("%s:%s", policy.Subject, policy.Object)] = policy
	return nil
}

func (ccm *channelCacheMock) Evaluate(_ context.Context, policy policies.Policy) bool {
	ccm.mu.Lock()
	defer ccm.mu.Unlock()

	for _, a := range ccm.policies[fmt.Sprintf("%s:%s", policy.Subject, policy.Object)].Actions {
		if a == policy.Actions[0] {
			return true
		}
	}

	return false
}

func (ccm *channelCacheMock) DeletePolicy(_ context.Context, policy policies.Policy) error {
	ccm.mu.Lock()
	defer ccm.mu.Unlock()

	delete(ccm.policies, fmt.Sprintf("%s:%s", policy.Subject, policy.Object))
	return nil
}
