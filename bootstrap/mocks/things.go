// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"
	"strconv"
	"sync"

	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/things/clients"
	"github.com/mainflux/mainflux/things/groups"
	"github.com/mainflux/mainflux/users/policies"
)

var _ clients.Service = (*mainfluxThings)(nil)

type mainfluxThings struct {
	mu          sync.Mutex
	counter     uint64
	things      map[string]clients.Client
	channels    map[string]groups.Group
	auth        policies.AuthServiceClient
	connections map[string][]string
}

// NewThingsService returns Mainflux Things service mock.
// Only methods used by SDK are mocked.
func NewThingsService(things map[string]clients.Client, channels map[string]groups.Group, auth policies.AuthServiceClient) clients.Service {
	return &mainfluxThings{
		things:      things,
		channels:    channels,
		auth:        auth,
		connections: make(map[string][]string),
	}
}

func (svc *mainfluxThings) CreateThings(_ context.Context, owner string, ths ...clients.Client) ([]clients.Client, error) {
	svc.mu.Lock()
	defer svc.mu.Unlock()

	userID, err := svc.auth.Identify(context.Background(), &policies.Token{Value: owner})
	if err != nil {
		return []clients.Client{}, errors.ErrAuthentication
	}
	for i := range ths {
		svc.counter++
		ths[i].Owner = userID.Email
		ths[i].ID = strconv.FormatUint(svc.counter, 10)
		ths[i].Credentials.Secret = ths[i].ID
		svc.things[ths[i].ID] = ths[i]
	}

	return ths, nil
}

func (svc *mainfluxThings) ViewClient(_ context.Context, owner, id string) (clients.Client, error) {
	svc.mu.Lock()
	defer svc.mu.Unlock()

	userID, err := svc.auth.Identify(context.Background(), &policies.Token{Value: owner})
	if err != nil {
		return clients.Client{}, errors.ErrAuthentication
	}

	if t, ok := svc.things[id]; ok && t.Owner == userID.Email {
		return t, nil

	}

	return clients.Client{}, errors.ErrNotFound
}

func (svc *mainfluxThings) Connect(_ context.Context, owner string, chIDs, thIDs []string) error {
	svc.mu.Lock()
	defer svc.mu.Unlock()

	userID, err := svc.auth.Identify(context.Background(), &policies.Token{Value: owner})
	if err != nil {
		return errors.ErrAuthentication
	}
	for _, chID := range chIDs {
		if svc.channels[chID].Owner != userID.Email {
			return errors.ErrAuthentication
		}
		svc.connections[chID] = append(svc.connections[chID], thIDs...)
	}

	return nil
}

func (svc *mainfluxThings) Disconnect(_ context.Context, owner string, chIDs, thIDs []string) error {
	svc.mu.Lock()
	defer svc.mu.Unlock()

	userID, err := svc.auth.Identify(context.Background(), &policies.Token{Value: owner})
	if err != nil {
		return errors.ErrAuthentication
	}

	for _, chID := range chIDs {
		if svc.channels[chID].Owner != userID.Email {
			return errors.ErrAuthentication
		}

		ids := svc.connections[chID]
		var count int
		var newConns []string
		for _, thID := range thIDs {
			for _, id := range ids {
				if id == thID {
					count++
					continue
				}
				newConns = append(newConns, id)
			}

			if len(newConns)-len(ids) != count {
				return errors.ErrNotFound
			}
			svc.connections[chID] = newConns
		}
	}
	return nil
}

func (svc *mainfluxThings) DisableClient(_ context.Context, owner, id string) (clients.Client, error) {
	panic("not implemented")
}

func (svc *mainfluxThings) EnableClient(_ context.Context, owner, id string) (clients.Client, error) {
	panic("not implemented")
}

func (svc *mainfluxThings) ViewChannel(_ context.Context, owner, id string) (groups.Group, error) {
	if c, ok := svc.channels[id]; ok {
		return c, nil
	}
	return groups.Group{}, errors.ErrNotFound
}

func (svc *mainfluxThings) UpdateClient(context.Context, string, clients.Client) (clients.Client, error) {
	panic("not implemented")
}

func (svc *mainfluxThings) UpdateClientOwner(context.Context, string, clients.Client) (clients.Client, error) {
	panic("not implemented")
}

func (svc *mainfluxThings) UpdateClientTags(context.Context, string, clients.Client) (clients.Client, error) {
	panic("not implemented")
}

func (svc *mainfluxThings) UpdateClientSecret(context.Context, string, string, string) (clients.Client, error) {
	panic("not implemented")
}

func (svc *mainfluxThings) ListClients(context.Context, string, clients.Page) (clients.ClientsPage, error) {
	panic("not implemented")
}

func (svc *mainfluxThings) ListChannelsByThing(context.Context, string, string, groups.Page) (groups.MembershipsPage, error) {
	panic("not implemented")
}

func (svc *mainfluxThings) ListThingsByChannel(context.Context, string, string, clients.Page) (clients.MembersPage, error) {
	panic("not implemented")
}

func (svc *mainfluxThings) CreateChannels(_ context.Context, owner string, chs ...groups.Group) ([]groups.Group, error) {
	svc.mu.Lock()
	defer svc.mu.Unlock()

	userID, err := svc.auth.Identify(context.Background(), &policies.Token{Value: owner})
	if err != nil {
		return []groups.Group{}, errors.ErrAuthentication
	}
	for i := range chs {
		svc.counter++
		chs[i].Owner = userID.Email
		chs[i].ID = strconv.FormatUint(svc.counter, 10)
		svc.channels[chs[i].ID] = chs[i]
	}

	return chs, nil
}

func (svc *mainfluxThings) UpdateChannel(context.Context, string, groups.Group) error {
	panic("not implemented")
}

func (svc *mainfluxThings) ListGroups(context.Context, string, groups.Page) (groups.GroupsPage, error) {
	panic("not implemented")
}

func (svc *mainfluxThings) RemoveChannel(context.Context, string, string) error {
	panic("not implemented")
}

func (svc *mainfluxThings) CanAccessByKey(context.Context, string, string) (string, error) {
	panic("not implemented")
}

func (svc *mainfluxThings) CanAccessByID(context.Context, string, string) error {
	panic("not implemented")
}

func (svc *mainfluxThings) IsChannelOwner(context.Context, string, string) error {
	panic("not implemented")
}

func (svc *mainfluxThings) Identify(context.Context, string) (string, error) {
	panic("not implemented")
}

func (svc *mainfluxThings) ShareThing(ctx context.Context, token, thingID string, actions []string, userIDs []string) error {
	panic("not implemented")
}

func findIndex(list []string, val string) int {
	for i, v := range list {
		if v == val {
			return i
		}
	}

	return -1
}

func (svc *mainfluxThings) ListMembers(ctx context.Context, token, groupID string, pm clients.Page) (clients.MembersPage, error) {
	panic("not implemented")
}
