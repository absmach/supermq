package bootstrap

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/mainflux/mainflux"
	mfsdk "github.com/mainflux/mainflux/sdk/go"
)

const (
	thingType = "device"
	chanName  = "channel"
)

var (
	// ErrNotFound indicates a non-existent entity request.
	ErrNotFound = errors.New("non-existent entity")

	// ErrMalformedEntity indicates malformed entity specification.
	ErrMalformedEntity = errors.New("malformed entity specification")

	// ErrUnauthorizedAccess indicates missing or invalid credentials provided
	// when accessing a protected resource.
	ErrUnauthorizedAccess = errors.New("missing or invalid credentials provided")

	// ErrInvalidID indicates that wrong ID is returned from the Mainflux.
	ErrInvalidID = errors.New("invalid Mainflux ID response")

	// ErrConflict indicates that entity with the same ID or external ID alredy exists.
	ErrConflict = errors.New("entity already exists")
)

var _ Service = (*bootstrapService)(nil)

// Service specifies an API that must be fulfilled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	// Add adds new Thing to the user identified by the provided key.
	Add(string, Thing) (Thing, error)

	// View returns Thing with given ID belonging to the user identified by the given key.
	View(string, string) (Thing, error)

	// Update updates editable fields of the provided Thing.
	Update(string, Thing) error

	// List returns subset of Things with given state that belong to the user identified by the given key.
	List(string, Filter, uint64, uint64) ([]Thing, error)

	// Remove removes Thing with specified key that belongs to the user identified by the given key.
	Remove(string, string) error

	// Bootstrap returns configuration to the Thing with provided external ID using external key.
	Bootstrap(string, string) (Config, error)

	// ChangeState changes state of the Thing with given ID and owner.
	ChangeState(string, string, State) error
}

// Config represents Thing configuration generated in bootstrapping process.
type Config struct {
	MFThing    string
	MFKey      string
	MFChannels []string
	Metadata   string
}

// ConfigReader is used to parse Config into format which will be encoded
// as a JSON and consumed from the client side.
type ConfigReader interface {
	ReadConfig(Config) (mainflux.Response, error)
}

type bootstrapService struct {
	things ThingRepository
	sdk    mfsdk.SDK
	users  mainflux.UsersServiceClient
}

// New returns new Bootstrap service.
func New(users mainflux.UsersServiceClient, things ThingRepository, sdk mfsdk.SDK) Service {
	return &bootstrapService{
		things: things,
		sdk:    sdk,
		users:  users,
	}
}

func (bs bootstrapService) Add(key string, thing Thing) (Thing, error) {
	owner, err := bs.identify(key)
	if err != nil {
		return Thing{}, err
	}

	// Check if channels exist.
	for _, c := range thing.MFChannels {
		if _, err := bs.sdk.Channel(c, key); err != nil {
			return Thing{}, ErrMalformedEntity
		}
	}

	resp, err := bs.sdk.CreateThing(mfsdk.Thing{Type: thingType}, key)
	if err != nil {
		return Thing{}, err
	}

	thingID, err := parseLocation(resp)
	if err != nil {
		return Thing{}, err
	}

	mfThing, err := bs.sdk.Thing(thingID, key)
	if err != nil {
		return Thing{}, bs.sdk.DeleteThing(thingID, key)
	}

	thing.Owner = owner
	thing.State = Created
	thing.MFThing = mfThing.ID
	thing.MFKey = mfThing.Key

	id, err := bs.things.Save(thing)
	if err != nil {
		return Thing{}, err
	}

	thing.ID = id
	return thing, nil
}

func (bs bootstrapService) View(key, id string) (Thing, error) {
	owner, err := bs.identify(key)
	if err != nil {
		return Thing{}, err
	}

	return bs.things.RetrieveByID(owner, id)
}

func (bs bootstrapService) Update(key string, thing Thing) error {
	owner, err := bs.identify(key)
	if err != nil {
		return err
	}

	thing.Owner = owner

	t, err := bs.things.RetrieveByID(owner, thing.ID)
	id := t.MFThing

	if err != nil {
		return err
	}

	if t.State == Active {
		tmp := make(map[string]bool)
		for _, c := range t.MFChannels {
			tmp[c] = true
		}

		for _, c := range thing.MFChannels {
			if !tmp[c] {
				if err := bs.sdk.ConnectThing(id, c, key); err != nil {
					return err
				}
				continue
			}

			delete(tmp, c)
		}

		for c := range tmp {
			if err := bs.sdk.DisconnectThing(id, c, key); err != nil {
				return err
			}
		}
	}

	return bs.things.Update(thing)
}

func (bs bootstrapService) List(key string, filter Filter, offset, limit uint64) ([]Thing, error) {
	owner, err := bs.identify(key)
	if err != nil {
		return []Thing{}, err
	}

	// All the Things with state other than NewThing have an owner.
	if state := filter["state"]; state != NewThing.String() {
		filter["owner"] = owner
	}

	return bs.things.RetrieveAll(filter, offset, limit), nil
}

func (bs bootstrapService) Remove(key, id string) error {
	owner, err := bs.identify(key)
	if err != nil {
		return err
	}

	thing, err := bs.things.RetrieveByID(owner, id)
	if err != nil {
		return err
	}

	if err := bs.sdk.DeleteThing(thing.MFThing, key); err != nil {
		return err
	}

	return bs.things.Remove(owner, id)
}

func (bs bootstrapService) Bootstrap(externalKey, externalID string) (Config, error) {
	thing, err := bs.things.RetrieveByExternalID(externalKey, externalID)
	if err == ErrNotFound {
		t := Thing{
			ExternalID:  externalID,
			ExternalKey: externalKey,
			State:       NewThing,
		}
		_, err := bs.things.Save(t)
		return Config{}, err
	}

	if err != nil {
		return Config{}, ErrUnauthorizedAccess
	}

	thing.State = Inactive
	if err := bs.things.ChangeState(thing.Owner, thing.ID, thing.State); err != nil {
		return Config{}, err
	}

	config := Config{
		MFThing:    thing.MFThing,
		MFKey:      thing.MFKey,
		MFChannels: thing.MFChannels,
		Metadata:   thing.Config,
	}

	return config, nil
}

func (bs bootstrapService) ChangeState(key, id string, state State) error {
	if state == NewThing || state == Created {
		return ErrMalformedEntity
	}

	owner, err := bs.identify(key)
	if err != nil {
		return err
	}

	thing, err := bs.things.RetrieveByID(owner, id)
	if err != nil {
		return err
	}

	switch state {
	case Active:
		for i, c := range thing.MFChannels {
			if err := bs.sdk.ConnectThing(thing.MFThing, c, key); err != nil {
				bs.connectionFallback(thing.MFThing, key, thing.MFChannels[:i], false)
				return err
			}
		}
	case Inactive:
		for i, c := range thing.MFChannels {
			if err := bs.sdk.DisconnectThing(thing.MFThing, c, key); err != nil {
				bs.connectionFallback(thing.MFThing, key, thing.MFChannels[:i], true)
				return err
			}
		}
	}

	return bs.things.ChangeState(owner, id, state)
}

func parseLocation(location string) (string, error) {
	mfPath := strings.Split(location, "/")
	n := len(mfPath)
	if n != 3 {
		return "", ErrInvalidID
	}

	return mfPath[n-1], nil
}

func (bs bootstrapService) identify(token string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := bs.users.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return "", ErrUnauthorizedAccess
	}

	return res.GetValue(), nil
}

func (bs bootstrapService) connectionFallback(id string, key string, channels []string, connect bool) {
	for _, c := range channels {
		if connect {
			bs.sdk.ConnectThing(id, c, key)
			continue
		}

		bs.sdk.DisconnectThing(id, c, key)
	}
}
