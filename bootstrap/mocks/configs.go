package mocks

import (
	"nov/bootstrap"
	"sort"
	"strconv"
	"sync"
)

var _ bootstrap.ConfigRepository = (*configRepositoryMock)(nil)

type configRepositoryMock struct {
	mu      sync.Mutex
	counter uint64
	configs map[string]bootstrap.Config
}

// NewThingsRepository creates in-memory thing repository.
func NewThingsRepository() bootstrap.ConfigRepository {
	return &configRepositoryMock{
		configs: make(map[string]bootstrap.Config),
	}
}

func (trm *configRepositoryMock) Save(config bootstrap.Config) (string, error) {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	trm.counter++
	config.ID = strconv.FormatUint(trm.counter, 10)
	trm.configs[config.ID] = config

	return config.ID, nil
}

func (trm *configRepositoryMock) RetrieveByID(owner, id string) (bootstrap.Config, error) {
	c, ok := trm.configs[id]
	if !ok {
		return bootstrap.Config{}, bootstrap.ErrNotFound
	}
	if c.Owner != owner {
		return bootstrap.Config{}, bootstrap.ErrUnauthorizedAccess
	}

	return c, nil

}

func (trm *configRepositoryMock) RetrieveAll(filter bootstrap.Filter, offset, limit uint64) []bootstrap.Config {
	configs := make([]bootstrap.Config, 0)

	if offset < 0 || limit <= 0 {
		return configs
	}

	owner := filter["owner"]
	first := uint64(offset) + 1
	last := first + uint64(limit)
	var state bootstrap.State = -1
	if s, ok := filter["state"]; ok {
		val, _ := strconv.Atoi(s)
		state = bootstrap.State(val)
	}

	for _, v := range trm.configs {
		id, _ := strconv.ParseUint(v.ID, 10, 64)
		if id >= first && id < last {
			if (state == -1 || v.State == state) && (owner == "" || v.Owner == owner) {
				configs = append(configs, v)
			}
		}
	}

	sort.SliceStable(configs, func(i, j int) bool {
		return configs[i].ID < configs[j].ID
	})

	return configs
}

func (trm *configRepositoryMock) RetrieveByExternalID(externalKey, externalID string) (bootstrap.Config, error) {
	for _, thing := range trm.configs {
		if thing.ExternalID == externalID && thing.ExternalKey == externalKey {
			return thing, nil
		}
	}

	return bootstrap.Config{}, bootstrap.ErrNotFound
}

func (trm *configRepositoryMock) Update(config bootstrap.Config) error {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	if _, ok := trm.configs[config.ID]; !ok {
		return bootstrap.ErrNotFound
	}

	trm.configs[config.ID] = config

	return nil
}

func (trm *configRepositoryMock) Remove(owner, id string) error {
	for k, v := range trm.configs {
		if v.Owner == owner && k == id {
			delete(trm.configs, k)
			break
		}
	}

	return nil
}

func (trm *configRepositoryMock) ChangeState(owner, id string, state bootstrap.State) error {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	config, ok := trm.configs[id]
	if !ok {
		return bootstrap.ErrNotFound
	}
	if config.Owner != owner {
		return bootstrap.ErrUnauthorizedAccess
	}

	config.State = state
	trm.configs[id] = config
	return nil
}

func (trm *configRepositoryMock) Assign(bootstrap.Config) error {
	return nil
}
