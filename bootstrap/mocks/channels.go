// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"
	"strconv"
	"sync"

	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/things/groups"
	upolicies "github.com/mainflux/mainflux/users/policies"
)

var _ groups.Service = (*mainfluxChannels)(nil)

type mainfluxChannels struct {
	mu       sync.Mutex
	counter  uint64
	channels map[string]groups.Group
	auth     upolicies.AuthServiceClient
}

// NewChannelsService returns Mainflux Things service mock.
// Only methods used by SDK are mocked.
func NewChannelsService(channels map[string]groups.Group, auth upolicies.AuthServiceClient) groups.Service {
	return &mainfluxChannels{
		channels: channels,
		auth:     auth,
	}
}

func (svc *mainfluxChannels) CreateGroups(_ context.Context, token string, chs ...groups.Group) ([]groups.Group, error) {
	svc.mu.Lock()
	defer svc.mu.Unlock()

	userID, err := svc.auth.Identify(context.Background(), &upolicies.Token{Value: token})
	if err != nil {
		return []groups.Group{}, errors.ErrAuthentication
	}
	for i := range chs {
		svc.counter++
		chs[i].Owner = userID.GetId()
		chs[i].ID = strconv.FormatUint(svc.counter, 10)
		svc.channels[chs[i].ID] = chs[i]
	}

	return chs, nil
}

func (svc *mainfluxChannels) ViewGroup(_ context.Context, owner, id string) (groups.Group, error) {
	if c, ok := svc.channels[id]; ok {
		return c, nil
	}
	return groups.Group{}, errors.ErrNotFound
}

func (svc *mainfluxChannels) ListGroups(context.Context, string, groups.GroupsPage) (groups.GroupsPage, error) {
	panic("not implemented")
}

func (svc *mainfluxChannels) ListMemberships(context.Context, string, string, groups.GroupsPage) (groups.MembershipsPage, error) {
	panic("not implemented")
}

func (svc *mainfluxChannels) UpdateGroup(context.Context, string, groups.Group) (groups.Group, error) {
	panic("not implemented")
}

func (svc *mainfluxChannels) EnableGroup(ctx context.Context, token, id string) (groups.Group, error) {
	svc.mu.Lock()
	defer svc.mu.Unlock()

	userID, err := svc.auth.Identify(context.Background(), &upolicies.Token{Value: token})
	if err != nil {
		return groups.Group{}, errors.ErrAuthentication
	}

	if t, ok := svc.channels[id]; !ok || t.Owner != userID.GetId() {
		return groups.Group{}, errors.ErrNotFound
	}
	if t, ok := svc.channels[id]; ok && t.Owner == userID.GetId() {
		t.Status = groups.EnabledStatus
		return t, nil
	}
	return groups.Group{}, nil
}

func (svc *mainfluxChannels) DisableGroup(ctx context.Context, token, id string) (groups.Group, error) {
	svc.mu.Lock()
	defer svc.mu.Unlock()

	userID, err := svc.auth.Identify(context.Background(), &upolicies.Token{Value: token})
	if err != nil {
		return groups.Group{}, errors.ErrAuthentication
	}

	if t, ok := svc.channels[id]; !ok || t.Owner != userID.GetId() {
		return groups.Group{}, errors.ErrNotFound
	}
	if t, ok := svc.channels[id]; ok && t.Owner == userID.GetId() {
		t.Status = groups.DisabledStatus
		return t, nil
	}
	return groups.Group{}, nil
}
