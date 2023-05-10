// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"
	"fmt"
	"strings"
	"sync"

	mfclients "github.com/mainflux/mainflux/pkg/clients"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/things/clients"
)

var _ mfclients.Repository = (*clientRepoMock)(nil)

type Connection struct {
	chanID    string
	thing     mfclients.Client
	connected bool
}

type clientRepoMock struct {
	mu      sync.Mutex
	counter uint64
	conns   chan Connection
	tconns  map[string]map[string]mfclients.Client
	things  map[string]mfclients.Client
}

// NewThingRepository creates in-memory thing repository.
func NewThingRepository(conns chan Connection) mfclients.Repository {
	repo := &clientRepoMock{
		conns:  conns,
		things: make(map[string]mfclients.Client),
		tconns: make(map[string]map[string]mfclients.Client),
	}
	go func(conns chan Connection, repo *clientRepoMock) {
		for conn := range conns {
			if !conn.connected {
				repo.disconnect(conn)
				continue
			}
			repo.connect(conn)
		}
	}(conns, repo)

	return repo
}

func (*clientRepoMock) UpdateIdentity(ctx context.Context, client mfclients.Client) (mfclients.Client, error) {
	return mfclients.Client{}, nil
}

func (*clientRepoMock) RetrieveByIdentity(ctx context.Context, identity string) (mfclients.Client, error) {
	return mfclients.Client{}, nil
}

func (trm *clientRepoMock) Save(_ context.Context, clis ...mfclients.Client) ([]mfclients.Client, error) {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	for _, cli := range clis {
		for _, th := range trm.things {
			if th.Credentials.Secret == cli.Credentials.Secret {
				return []mfclients.Client{}, errors.ErrConflict
			}
		}

		trm.counter++
		if cli.ID == "" {
			cli.ID = fmt.Sprintf("%03d", trm.counter)
		}
		trm.things[key(cli.Owner, cli.ID)] = cli
	}
	return clis, nil
}

func (trm *clientRepoMock) Update(_ context.Context, thing mfclients.Client) (mfclients.Client, error) {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	dbKey := key(thing.Owner, thing.ID)

	if _, ok := trm.things[dbKey]; !ok {
		return mfclients.Client{}, errors.ErrNotFound
	}

	trm.things[dbKey] = thing

	return trm.things[dbKey], nil
}

func (trm *clientRepoMock) UpdateSecret(_ context.Context, client mfclients.Client) (mfclients.Client, error) {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	for _, th := range trm.things {
		if th.Credentials.Secret == client.Credentials.Secret {
			return mfclients.Client{}, errors.ErrConflict
		}
	}

	dbKey := key(client.Owner, client.ID)

	th, ok := trm.things[dbKey]
	if !ok {
		return mfclients.Client{}, errors.ErrNotFound
	}

	th.Credentials.Secret = client.Credentials.Secret
	trm.things[dbKey] = th

	return trm.things[dbKey], nil
}

func (trm *clientRepoMock) UpdateOwner(_ context.Context, client mfclients.Client) (mfclients.Client, error) {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	dbKey := key(client.Owner, client.ID)

	th, ok := trm.things[dbKey]
	if !ok {
		return mfclients.Client{}, errors.ErrNotFound
	}

	th.Owner = client.Owner
	trm.things[dbKey] = th

	return trm.things[dbKey], nil
}

func (trm *clientRepoMock) UpdateTags(_ context.Context, client mfclients.Client) (mfclients.Client, error) {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	dbKey := key(client.Owner, client.ID)

	th, ok := trm.things[dbKey]
	if !ok {
		return mfclients.Client{}, errors.ErrNotFound
	}

	th.Tags = client.Tags
	trm.things[dbKey] = th

	return trm.things[dbKey], nil
}

func (trm *clientRepoMock) RetrieveByID(_ context.Context, id string) (mfclients.Client, error) {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	if c, ok := trm.things[id]; ok {
		return c, nil
	}

	return mfclients.Client{}, errors.ErrNotFound
}

func (trm *clientRepoMock) RetrieveAll(_ context.Context, pm mfclients.Page) (mfclients.ClientsPage, error) {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	first := uint64(pm.Offset) + 1
	last := first + uint64(pm.Limit)

	var ths []mfclients.Client

	// This obscure way to examine map keys is enforced by the key structure
	// itself (see mocks/commons.go).
	prefix := "owner"
	for k, v := range trm.things {
		id := parseID(v.ID)
		if strings.HasPrefix(k, prefix) && id >= first && id < last {
			ths = append(ths, v)
		}
	}

	// Sort Things list
	ths = sortThings(pm, ths)

	page := mfclients.ClientsPage{
		Clients: ths,
		Page: mfclients.Page{
			Total:  trm.counter,
			Offset: pm.Offset,
			Limit:  pm.Limit,
		},
	}

	return page, nil
}

func (trm *clientRepoMock) Members(_ context.Context, chID string, pm mfclients.Page) (mfclients.MembersPage, error) {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	if pm.Limit <= 0 {
		return mfclients.MembersPage{}, nil
	}

	first := uint64(pm.Offset) + 1
	last := first + uint64(pm.Limit)

	var ths []mfclients.Client

	// Append connected or not connected channels
	switch pm.Disconnected {
	case false:
		for _, co := range trm.tconns[chID] {
			id := parseID(co.ID)
			if id >= first && id < last {
				ths = append(ths, co)
			}
		}
	default:
		for _, th := range trm.things {
			conn := false
			id := parseID(th.ID)
			if id >= first && id < last {
				for _, co := range trm.tconns[chID] {
					if th.ID == co.ID {
						conn = true
					}
				}

				// Append if not found in connections list
				if !conn {
					ths = append(ths, th)
				}
			}
		}
	}

	// Sort Things by Channel list
	ths = sortThings(pm, ths)

	page := mfclients.MembersPage{
		Members: ths,
		Page: mfclients.Page{
			Total:  trm.counter,
			Offset: pm.Offset,
			Limit:  pm.Limit,
		},
	}

	return page, nil
}

func (trm *clientRepoMock) ChangeStatus(_ context.Context, client mfclients.Client) (mfclients.Client, error) {
	trm.mu.Lock()
	defer trm.mu.Unlock()
	th := trm.things[client.ID]
	th.Status = client.Status
	trm.things[client.ID] = th
	return th, nil
}

func (trm *clientRepoMock) RetrieveBySecret(_ context.Context, key string) (mfclients.Client, error) {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	for _, thing := range trm.things {
		if thing.Credentials.Secret == key {
			return thing, nil
		}
	}

	return mfclients.Client{}, errors.ErrNotFound
}

func (trm *clientRepoMock) connect(conn Connection) {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	if _, ok := trm.tconns[conn.chanID]; !ok {
		trm.tconns[conn.chanID] = make(map[string]mfclients.Client)
	}
	trm.tconns[conn.chanID][conn.thing.ID] = conn.thing
}

func (trm *clientRepoMock) disconnect(conn Connection) {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	if conn.thing.ID == "" {
		delete(trm.tconns, conn.chanID)
		return
	}
	delete(trm.tconns[conn.chanID], conn.thing.ID)
}

type clientCacheMock struct {
	mu     sync.Mutex
	things map[string]string
}

// NewClientCache returns mock cache instance.
func NewClientCache() clients.ClientCache {
	return &clientCacheMock{
		things: make(map[string]string),
	}
}

func (tcm *clientCacheMock) Save(_ context.Context, key, id string) error {
	tcm.mu.Lock()
	defer tcm.mu.Unlock()

	tcm.things[key] = id
	return nil
}

func (tcm *clientCacheMock) ID(_ context.Context, key string) (string, error) {
	tcm.mu.Lock()
	defer tcm.mu.Unlock()

	id, ok := tcm.things[key]
	if !ok {
		return "", errors.ErrNotFound
	}

	return id, nil
}

func (tcm *clientCacheMock) Remove(_ context.Context, id string) error {
	tcm.mu.Lock()
	defer tcm.mu.Unlock()

	for key, val := range tcm.things {
		if val == id {
			delete(tcm.things, key)
			return nil
		}
	}

	return nil
}
