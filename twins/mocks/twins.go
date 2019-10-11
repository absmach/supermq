//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package mocks

import (
	"context"
	"sort"
	"strconv"
	"sync"

	"github.com/mainflux/mainflux/twins"
)

var _ twins.TwinRepository = (*twinRepositoryMock)(nil)

type twinRepositoryMock struct {
	mu      sync.Mutex
	counter uint64
	twins   map[string]twins.Twin
	tconns  map[string]map[string]twins.Twin
}

// NewTwinRepository creates in-memory twin repository.
func NewTwinRepository() twins.TwinRepository {
	return &twinRepositoryMock{
		twins:  make(map[string]twins.Twin),
		tconns: make(map[string]map[string]twins.Twin),
	}
}

func (trm *twinRepositoryMock) Save(ctx context.Context, twin twins.Twin) (string, error) {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	for _, tw := range trm.twins {
		if tw.ID == twin.ID {
			return "", twins.ErrConflict
		}
	}

	for _, tw := range trm.twins {
		if tw.Key == twin.Key {
			return "", twins.ErrConflict
		}
	}

	trm.counter++
	twin.ID = strconv.FormatUint(trm.counter, 10)
	trm.twins[key(twin.Owner, twin.ID)] = twin

	return twin.ID, nil
}

func (trm *twinRepositoryMock) Update(ctx context.Context, twin twins.Twin) error {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	dbKey := key(twin.Owner, twin.ID)

	if _, ok := trm.twins[dbKey]; !ok {
		return twins.ErrNotFound
	}

	trm.twins[dbKey] = twin

	return nil
}

func (trm *twinRepositoryMock) UpdateKey(ctx context.Context, owner, id, val string) error {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	for _, tw := range trm.twins {
		if tw.Key == val {
			return twins.ErrConflict
		}
	}

	dbKey := key(owner, id)

	tw, ok := trm.twins[dbKey]
	if !ok {
		return twins.ErrNotFound
	}

	tw.Key = val
	trm.twins[dbKey] = tw

	return nil
}

func (trm *twinRepositoryMock) RetrieveByID(_ context.Context, owner, id string) (twins.Twin, error) {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	if tw, ok := trm.twins[key(owner, id)]; ok {
		return tw, nil
	}

	return twins.Twin{}, twins.ErrNotFound
}

func (trm *twinRepositoryMock) RetrieveByKey(_ context.Context, key string) (string, error) {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	for _, twin := range trm.twins {
		if twin.Key == key {
			return twin.ID, nil
		}
	}

	return "", twins.ErrNotFound
}

func (trm *twinRepositoryMock) RetrieveByChannel(_ context.Context, chanID string, limit uint64) (twins.TwinsSet, error) {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	items := make([]twins.Twin, 0)

	if limit <= 0 {
		return twins.TwinsSet{}, nil
	}

	tws, ok := trm.tconns[chanID]
	if !ok {
		return twins.TwinsSet{}, nil
	}

	for _, v := range tws {
		id, _ := strconv.ParseUint(v.ID, 10, 64)
		if id < limit {
			items = append(items, v)
		}
	}

	sort.SliceStable(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})

	page := twins.TwinsSet{
		Twins: items,
		SetMetadata: twins.SetMetadata{
			Total: trm.counter,
			Limit: limit,
		},
	}

	return page, nil
}

func (trm *twinRepositoryMock) Remove(ctx context.Context, owner, id string) error {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	delete(trm.twins, key(owner, id))

	return nil
}
