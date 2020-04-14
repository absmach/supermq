package mocks

//
import (
	"sync"

	"github.com/gofrs/uuid"
	provSDK "github.com/mainflux/mainflux/provision/sdk"
	mfSDK "github.com/mainflux/mainflux/sdk/go"
)

const (
	validEmail   = "test@example.com"
	validPass    = "test"
	invalid      = "invalid"
	validToken   = "valid_token"
	invalidToken = "invalid_token"
)

var thingIDs = []string{"ids"}

// SDK is fake sdk for mocking
type mockSDK struct {
	things      map[string]provSDK.Thing
	channels    map[string]provSDK.Channel
	connections map[string][]string
	configs     map[string]provSDK.BSConfig
	mu          sync.Mutex
}

// NewSDK returns new mock SDK for testing purposes.
func NewSDK() provSDK.SDK {
	sdk := &mockSDK{}
	sdk.channels = make(map[string]provSDK.Channel)
	sdk.connections = make(map[string][]string)
	sdk.configs = make(map[string]provSDK.BSConfig)

	th := provSDK.Thing{ID: "predefined", Name: "ID"}
	sdk.things = map[string]provSDK.Thing{"predefined": th}
	sdk.mu = sync.Mutex{}

	return sdk
}

// CreateToken receives credentials and returns user token.
func (s *mockSDK) CreateToken(email, pass string) (string, error) {
	if email != validEmail || pass != validPass {
		return "", mfSDK.ErrUnauthorized
	}
	return validToken, nil
}

func (s *mockSDK) Cert(thingID, thingKey string, token string) (provSDK.Cert, error) {
	if thingID == invalid || thingKey == invalid {
		return provSDK.Cert{}, provSDK.ErrCerts
	}
	return provSDK.Cert{}, nil
}

func (s *mockSDK) SaveConfig(data provSDK.BSConfig, token string) error {
	if data.ThingID == invalid {
		return mfSDK.ErrFailedCreation
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.configs[data.ExternalID]; ok {
		return provSDK.ErrConflict
	}
	s.configs[data.ExternalID] = data
	return nil
}

func (s *mockSDK) Whitelist(thingID string, data map[string]int, token string) error {
	if thingID == invalid {
		return provSDK.ErrWhitelist
	}
	return nil
}

func (s *mockSDK) RemoveConfig(id string, token string) error {
	if id == invalid {
		return provSDK.ErrConfigRemove
	}
	return nil
}

func (s *mockSDK) RemoveCert(key string, token string) error {
	if key == invalid {
		return provSDK.ErrCertsRemove
	}
	return nil
}

func (s *mockSDK) CreateThing(externalID string, name string, token string) (string, error) {
	if token != validToken {
		return "", mfSDK.ErrUnauthorized
	}

	id, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	key, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	newThing := provSDK.Thing{ID: id.String(), Name: name, Key: key.String(), Metadata: map[string]interface{}{"ExternalID": externalID}}
	s.things[newThing.ID] = newThing

	return newThing.ID, nil
}

func (s *mockSDK) Thing(id, token string) (provSDK.Thing, error) {
	t := provSDK.Thing{}

	if token != validToken {
		return t, mfSDK.ErrUnauthorized
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if t, ok := s.things[id]; ok {
		return t, nil
	}

	return t, mfSDK.ErrNotFound

}

func (s *mockSDK) DeleteThing(id string, token string) error {
	if id == invalid {
		return mfSDK.ErrFailedRemoval
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.things, id)
	return nil
}

func (s *mockSDK) CreateChannel(ch mfSDK.Channel, token string) (provSDK.Channel, error) {
	if token != validToken {
		return provSDK.Channel{}, mfSDK.ErrUnauthorized
	}

	id, err := uuid.NewV4()
	if err != nil {
		return provSDK.Channel{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	newChan := provSDK.Channel{ID: id.String(), Name: name, Metadata: map[string]interface{}{"Type": chantype}}
	s.channels[newChan.ID] = newChan

	return newChan, nil
}

func (s *mockSDK) DeleteChannel(id string, token string) error {
	if id == invalid {
		return mfSDK.ErrFailedRemoval
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.channels, id)
	return nil
}

// ConnectThing connects thing to specified channel by id.
func (s *mockSDK) Connect(thingID, chanID, token string) error {
	if token != validToken {
		return mfSDK.ErrUnauthorized
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.things[thingID]; !ok {
		return mfSDK.ErrNotFound
	}
	if _, ok := s.channels[chanID]; !ok {
		return mfSDK.ErrNotFound
	}

	conns := s.connections[thingID]
	conns = append(conns, chanID)
	s.connections[thingID] = conns
	return nil
}
