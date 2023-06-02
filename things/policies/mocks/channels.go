package mocks

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/things/policies"
)

type channelCacheMock struct {
	mu       sync.Mutex
	policies map[string]string
}

// NewChannelCache returns mock cache instance.
func NewChannelCache() policies.Cache {
	return &channelCacheMock{
		policies: make(map[string]string),
	}
}

func (ccm *channelCacheMock) Put(_ context.Context, policy policies.Policy) error {
	ccm.mu.Lock()
	defer ccm.mu.Unlock()

	ccm.policies[fmt.Sprintf("%s:%s", policy.Subject, policy.Object)] = strings.Join(policy.Actions, ":")
	return nil
}

func (ccm *channelCacheMock) Get(_ context.Context, policy policies.Policy) (policies.Policy, error) {
	ccm.mu.Lock()
	defer ccm.mu.Unlock()
	actions := ccm.policies[fmt.Sprintf("%s:%s", policy.Subject, policy.Object)]

	if actions != "" {
		return policies.Policy{
			Subject: policy.Subject,
			Object:  policy.Object,
			Actions: strings.Split(actions, ":"),
		}, nil
	}

	return policies.Policy{}, errors.ErrNotFound
}

func (ccm *channelCacheMock) Remove(_ context.Context, policy policies.Policy) error {
	ccm.mu.Lock()
	defer ccm.mu.Unlock()

	delete(ccm.policies, fmt.Sprintf("%s:%s", policy.Subject, policy.Object))
	return nil
}
