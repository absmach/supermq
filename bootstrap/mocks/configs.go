package mocks

import (
	"nov/bootstrap"
	"sort"
	"strconv"
	"sync"
)

var _ bootstrap.ConfigRepository = (*configRepositoryMock)(nil)

type configRepositoryMock struct {
	mu       sync.Mutex
	counter  uint64
	configs  map[string]bootstrap.Config
	unknowns map[string]string
}

// NewThingsRepository creates in-memory thing repository.
func NewThingsRepository() bootstrap.ConfigRepository {
	return &configRepositoryMock{
		configs:  make(map[string]bootstrap.Config),
		unknowns: make(map[string]string),
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

func (trm *configRepositoryMock) RetrieveByID(key, id string) (bootstrap.Config, error) {
	c, ok := trm.configs[id]
	if !ok {
		return bootstrap.Config{}, bootstrap.ErrNotFound
	}
	if c.Owner != key {
		return bootstrap.Config{}, bootstrap.ErrUnauthorizedAccess
	}

	return c, nil

}

func (trm *configRepositoryMock) RetrieveAll(key string, filter bootstrap.Filter, offset, limit uint64) []bootstrap.Config {
	configs := make([]bootstrap.Config, 0)

	if offset < 0 || limit <= 0 {
		return configs
	}

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
			if (state == -1 || v.State == state) && v.Owner == key {
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

func (trm *configRepositoryMock) Remove(key, id string) error {
	for k, v := range trm.configs {
		if v.Owner == key && k == id {
			delete(trm.configs, k)
			break
		}
	}

	return nil
}

func (trm *configRepositoryMock) ChangeState(key, id string, state bootstrap.State) error {
	trm.mu.Lock()
	defer trm.mu.Unlock()

	config, ok := trm.configs[id]
	if !ok {
		return bootstrap.ErrNotFound
	}
	if config.Owner != key {
		return bootstrap.ErrUnauthorizedAccess
	}

	config.State = state
	trm.configs[id] = config
	return nil
}

func (trm *configRepositoryMock) RetrieveUnknown(offset, limit uint64) []bootstrap.Config {
	return []bootstrap.Config{}
}

func (trm *configRepositoryMock) RemoveUnknown(string, string) error {
	return nil
}

func (trm *configRepositoryMock) SaveUnknown(key, id string) error {
	return nil
}
