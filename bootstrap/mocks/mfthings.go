package mocks

import (
	"strconv"
	"sync"

	"github.com/mainflux/mainflux/things"
	"github.com/mainflux/mainflux/users"
)

var _ things.Service = (*mainfluxThings)(nil)

type mainfluxThings struct {
	mu          sync.Mutex
	counter     uint64
	things      map[string]things.Thing
	channels    map[string]things.Channel
	users       map[string]users.User
	connections map[string][]string
}

// NewMainfluxThingsService returns (svc *mainfluxThings) things service mock.
// Only methods used by SDK are mocked.
func NewMainfluxThingsService(things map[string]things.Thing, channels map[string]things.Channel, users map[string]users.User, connections map[string][]string) *mainfluxThings {
	return &mainfluxThings{
		things:   things,
		channels: channels,
		users:    users,
	}
}

func (svc *mainfluxThings) AddThing(owner string, thing things.Thing) (things.Thing, error) {
	svc.mu.Lock()
	defer svc.mu.Unlock()

	_, ok := svc.users[owner]
	if !ok {
		return things.Thing{}, things.ErrUnauthorizedAccess
	}

	svc.counter++
	thing.Owner = owner
	thing.ID = strconv.FormatUint(svc.counter, 10)
	svc.things[key(owner, thing.ID)] = thing
	return thing, nil
}

func (svc *mainfluxThings) ViewThing(owner, id string) (things.Thing, error) {
	svc.mu.Lock()
	defer svc.mu.Unlock()

	_, ok := svc.users[owner]
	if !ok {
		return things.Thing{}, things.ErrUnauthorizedAccess
	}

	if t, ok := svc.things[id]; ok {
		return t, nil
	}

	return things.Thing{}, things.ErrNotFound
}

func (svc *mainfluxThings) Connect(owner, thingID, chanID string) error {
	svc.mu.Lock()
	defer svc.mu.Unlock()

	_, ok := svc.users[owner]
	if !ok {
		return things.ErrUnauthorizedAccess
	}
	svc.connections[chanID] = append(svc.connections[chanID], thingID)
	return nil
}

func (svc *mainfluxThings) Disconnect(owner, thingID, chanID string) error {
	svc.mu.Lock()
	defer svc.mu.Unlock()

	_, ok := svc.users[owner]
	if !ok {
		return things.ErrUnauthorizedAccess
	}

	ids := svc.connections[chanID]
	i := 0
	for _, t := range ids {
		if t == thingID {
			break
		}
		i++
	}
	if i == len(ids) {
		return things.ErrNotFound
	}

	var tmp []string
	if i != len(ids)-2 {
		tmp = ids[i+1:]
	}
	ids = append(ids[:i], tmp...)
	svc.connections[chanID] = ids

	return nil
}

func (svc *mainfluxThings) RemoveThing(owner, id string) error {
	svc.mu.Lock()
	defer svc.mu.Unlock()

	_, ok := svc.users[owner]
	if !ok {
		return things.ErrUnauthorizedAccess
	}

	dbKey := key(owner, id)
	if _, ok := svc.things[dbKey]; !ok {
		return things.ErrNotFound
	}

	delete(svc.things, dbKey)
	conns := make(map[string][]string)
	for k, v := range svc.connections {
		idx := findIndex(v, dbKey)
		if idx != -1 {
			var tmp []string
			if idx != len(v)-2 {
				tmp = v[idx+1:]
			}
			conns[k] = append(v[:idx], tmp...)
		}
	}

	svc.connections = conns
	return nil
}

func findIndex(list []string, val string) int {
	for i, v := range list {
		if v == val {
			return i
		}
	}

	return -1
}

func (svc *mainfluxThings) UpdateThing(string, things.Thing) error {
	panic("not implemented")
}

func (svc *mainfluxThings) ListThings(string, uint64, uint64) ([]things.Thing, error) {
	panic("not implemented")
}

func (svc *mainfluxThings) CreateChannel(string, things.Channel) (things.Channel, error) {
	panic("not implemented")
}

func (svc *mainfluxThings) UpdateChannel(string, things.Channel) error {
	panic("not implemented")
}

func (svc *mainfluxThings) ViewChannel(string, string) (things.Channel, error) {
	panic("not implemented")
}

func (svc *mainfluxThings) ListChannels(string, uint64, uint64) ([]things.Channel, error) {
	panic("not implemented")
}

func (svc *mainfluxThings) RemoveChannel(string, string) error {
	panic("not implemented")
}

func (svc *mainfluxThings) CanAccess(string, string) (string, error) {
	panic("not implemented")
}

func (svc *mainfluxThings) Identify(string) (string, error) {
	panic("not implemented")
}
