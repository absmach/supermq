// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/things/clients"
)

var _ clients.ClientRepository = (*thingRepositoryMock)(nil)

type Connection struct {
	chanID    string
	thing     clients.Client
	connected bool
}

type thingRepositoryMock struct {
	mu      sync.Mutex
	counter uint64
	conns   chan Connection
	tconns  map[string]map[string]clients.Client
	things  map[string]clients.Client
}

// NewThingRepository creates in-memory thing repository.
func NewThingRepository(conns chan Connection) clients.ClientRepository {
	repo := &thingRepositoryMock{
		conns:  conns,
		things: make(map[string]clients.Client),
		tconns: make(map[string]map[string]clients.Client),
	}
	go func(conns chan Connection, repo *thingRepositoryMock) {
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

func (trm *thingRepositoryMock) Save(_ context.Context, clis ...clients.Client) ([]clients.Client, error) {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	for _, cli := range clis {
		for _, th := range trm.things {
			if th.Credentials.Secret == cli.Credentials.Secret {
				return []clients.Client{}, errors.ErrConflict
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

func (trm *thingRepositoryMock) Update(_ context.Context, thing clients.Client) (clients.Client, error) {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	dbKey := key(thing.Owner, thing.ID)

	if _, ok := trm.things[dbKey]; !ok {
		return clients.Client{}, errors.ErrNotFound
	}

	trm.things[dbKey] = thing

	return trm.things[dbKey], nil
}

func (trm *thingRepositoryMock) UpdateSecret(_ context.Context, client clients.Client) (clients.Client, error) {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	for _, th := range trm.things {
		if th.Credentials.Secret == client.Credentials.Secret {
			return clients.Client{}, errors.ErrConflict
		}
	}

	dbKey := key(client.Owner, client.ID)

	th, ok := trm.things[dbKey]
	if !ok {
		return clients.Client{}, errors.ErrNotFound
	}

	th.Credentials.Secret = client.Credentials.Secret
	trm.things[dbKey] = th

	return trm.things[dbKey], nil
}

func (trm *thingRepositoryMock) UpdateOwner(_ context.Context, client clients.Client) (clients.Client, error) {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	dbKey := key(client.Owner, client.ID)

	th, ok := trm.things[dbKey]
	if !ok {
		return clients.Client{}, errors.ErrNotFound
	}

	th.Owner = client.Owner
	trm.things[dbKey] = th

	return trm.things[dbKey], nil
}

func (trm *thingRepositoryMock) UpdateTags(_ context.Context, client clients.Client) (clients.Client, error) {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	dbKey := key(client.Owner, client.ID)

	th, ok := trm.things[dbKey]
	if !ok {
		return clients.Client{}, errors.ErrNotFound
	}

	th.Tags = client.Tags
	trm.things[dbKey] = th

	return trm.things[dbKey], nil
}

func (trm *thingRepositoryMock) RetrieveByID(_ context.Context, id string) (clients.Client, error) {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	if c, ok := trm.things[id]; ok {
		return c, nil
	}

	return clients.Client{}, errors.ErrNotFound
}

func (trm *thingRepositoryMock) RetrieveAll(_ context.Context, pm clients.Page) (clients.ClientsPage, error) {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	first := uint64(pm.Offset) + 1
	last := first + uint64(pm.Limit)

	var ths []clients.Client

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

	page := clients.ClientsPage{
		Clients: ths,
		Page: clients.Page{
			Total:  trm.counter,
			Offset: pm.Offset,
			Limit:  pm.Limit,
		},
	}

	return page, nil
}

func (trm *thingRepositoryMock) Members(_ context.Context, chID string, pm clients.Page) (clients.MembersPage, error) {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	if pm.Limit <= 0 {
		return clients.MembersPage{}, nil
	}

	first := uint64(pm.Offset) + 1
	last := first + uint64(pm.Limit)

	var ths []clients.Client

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

	page := clients.MembersPage{
		Members: ths,
		Page: clients.Page{
			Total:  trm.counter,
			Offset: pm.Offset,
			Limit:  pm.Limit,
		},
	}

	return page, nil
}

func (trm *thingRepositoryMock) ChangeStatus(_ context.Context, id string, status clients.Status) (clients.Client, error) {
	trm.mu.Lock()
	defer trm.mu.Unlock()
	th := trm.things[id]
	th.Status = status
	trm.things[id] = th
	return th, nil
}

func (trm *thingRepositoryMock) RetrieveBySecret(_ context.Context, key string) (clients.Client, error) {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	for _, thing := range trm.things {
		if thing.Credentials.Secret == key {
			return thing, nil
		}
	}

	return clients.Client{}, errors.ErrNotFound
}

func (trm *thingRepositoryMock) connect(conn Connection) {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	if _, ok := trm.tconns[conn.chanID]; !ok {
		trm.tconns[conn.chanID] = make(map[string]clients.Client)
	}
	trm.tconns[conn.chanID][conn.thing.ID] = conn.thing
}

func (trm *thingRepositoryMock) disconnect(conn Connection) {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	if conn.thing.ID == "" {
		delete(trm.tconns, conn.chanID)
		return
	}
	delete(trm.tconns[conn.chanID], conn.thing.ID)
}

type thingCacheMock struct {
	mu     sync.Mutex
	things map[string]string
}

// NewThingCache returns mock cache instance.
func NewThingCache() clients.ThingCache {
	return &thingCacheMock{
		things: make(map[string]string),
	}
}

func (tcm *thingCacheMock) Save(_ context.Context, key, id string) error {
	tcm.mu.Lock()
	defer tcm.mu.Unlock()

	tcm.things[key] = id
	return nil
}

func (tcm *thingCacheMock) ID(_ context.Context, key string) (string, error) {
	tcm.mu.Lock()
	defer tcm.mu.Unlock()

	id, ok := tcm.things[key]
	if !ok {
		return "", errors.ErrNotFound
	}

	return id, nil
}

func (tcm *thingCacheMock) Remove(_ context.Context, id string) error {
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
