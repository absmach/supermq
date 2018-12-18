package mocks

import (
	"nov/bootstrap"
	"sort"
	"strconv"
	"sync"
)

var _ bootstrap.ThingRepository = (*thingRepositoryMock)(nil)

type thingRepositoryMock struct {
	mu      sync.Mutex
	counter uint64
	things  map[string]bootstrap.Thing
}

// NewThingsRepository creates in-memory thing repository.
func NewThingsRepository() bootstrap.ThingRepository {
	return &thingRepositoryMock{
		things: make(map[string]bootstrap.Thing),
	}
}

func (trm *thingRepositoryMock) Save(thing bootstrap.Thing) (string, error) {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	trm.counter++
	thing.ID = strconv.FormatUint(trm.counter, 10)
	trm.things[thing.ID] = thing

	return thing.ID, nil
}

func (trm *thingRepositoryMock) RetrieveByID(owner, id string) (bootstrap.Thing, error) {
	c, ok := trm.things[id]
	if !ok {
		return bootstrap.Thing{}, bootstrap.ErrNotFound
	}
	if c.Owner != owner {
		return bootstrap.Thing{}, bootstrap.ErrUnauthorizedAccess
	}

	return c, nil

}

func (trm *thingRepositoryMock) RetrieveAll(filter bootstrap.Filter, offset, limit uint64) []bootstrap.Thing {
	things := make([]bootstrap.Thing, 0)

	if offset < 0 || limit <= 0 {
		return things
	}

	owner := filter["owner"]
	first := uint64(offset) + 1
	last := first + uint64(limit)
	var state bootstrap.State = -1
	if s, ok := filter["state"]; ok {
		val, _ := strconv.Atoi(s)
		state = bootstrap.State(val)
	}

	for _, v := range trm.things {
		id, _ := strconv.ParseUint(v.ID, 10, 64)
		if id >= first && id < last {
			if (state == -1 || v.State == state) && (owner == "" || v.Owner == owner) {
				things = append(things, v)
			}
		}
	}

	sort.SliceStable(things, func(i, j int) bool {
		return things[i].ID < things[j].ID
	})

	return things
}

func (trm *thingRepositoryMock) RetrieveByExternalID(externalKey, externalID string) (bootstrap.Thing, error) {
	for _, thing := range trm.things {
		if thing.ExternalID == externalID && thing.ExternalKey == externalKey {
			return thing, nil
		}
	}

	return bootstrap.Thing{}, bootstrap.ErrNotFound
}

func (trm *thingRepositoryMock) Update(thing bootstrap.Thing) error {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	if _, ok := trm.things[thing.ID]; !ok {
		return bootstrap.ErrNotFound
	}

	trm.things[thing.ID] = thing

	return nil
}

func (trm *thingRepositoryMock) Remove(owner, id string) error {
	for k, v := range trm.things {
		if v.Owner == owner && k == id {
			delete(trm.things, k)
			break
		}
	}

	return nil
}

func (trm *thingRepositoryMock) ChangeState(owner, id string, state bootstrap.State) error {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	thing, ok := trm.things[id]
	if !ok {
		return bootstrap.ErrNotFound
	}
	if thing.Owner != owner {
		return bootstrap.ErrUnauthorizedAccess
	}

	thing.State = state
	trm.things[id] = thing
	return nil
}

func (trm *thingRepositoryMock) Assign(bootstrap.Thing) error {
	return nil
}
