package bootstrap

import (
	"errors"
	"strings"

	"github.com/mainflux/mainflux"
	mfsdk "github.com/mainflux/mainflux/sdk/go"
)

const (
	thingType = "device"
	chanName  = "NOV"
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

	// ErrInvalidStatus indicates attempt to change status of the unconfigured entity.
	// Entity needs to be bootstrapped in order to be able to change status.
	ErrInvalidStatus = errors.New("can change status of unconfigured entity")
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

	// List returns subset of Things that belong to the user identified by the given key.
	List(string, uint64, uint64) ([]Thing, error)

	// Remove removes Thing with specified key that belongs to the user identified by the given key.
	Remove(string, string) error

	// Bootstrap returns initial configuration to the Thing with provided external ID.
	Bootstrap(string) (Config, error)

	// ChangeStatus changes status of the Thing with given ID and owner.
	ChangeStatus(string, string, Status) error
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
	idp    IdentityProvider
	config string
}

// New returns new Bootstrap service.
func New(things ThingRepository, idp IdentityProvider, sdk mfsdk.SDK, cfgPath string) Service {
	return &bootstrapService{
		things: things,
		sdk:    sdk,
		idp:    idp,
		config: cfgPath,
	}
}

func (bs bootstrapService) Add(key string, thing Thing) (Thing, error) {
	owner, err := bs.idp.ExtractKey(key)
	if err != nil {
		return Thing{}, err
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
	thing.Status = Created
	thing.MFThing = thingID
	thing.MFKey = mfThing.Key

	id, err := bs.things.Save(thing)
	if err != nil {
		return Thing{}, err
	}

	thing.ID = id
	return thing, nil
}

func (bs bootstrapService) View(key, id string) (Thing, error) {
	owner, err := bs.idp.ExtractKey(key)
	if err != nil {
		return Thing{}, err
	}

	return bs.things.RetrieveByID(owner, id)
}

func (bs bootstrapService) Update(key string, thing Thing) error {
	owner, err := bs.idp.Identify(key)
	if err != nil {
		return err
	}

	thing.Owner = owner

	t, err := bs.things.RetrieveByID(owner, thing.ID)
	id := t.MFThing

	if err != nil {
		return err
	}

	if t.Status == Active {
		intersect := []string{}
		for _, c := range t.MFChannels {
			if contains(thing.MFChannels, c) {
				intersect = append(intersect, c)
			}
		}

		for _, c := range t.MFChannels {
			if !contains(intersect, c) {
				if err := bs.sdk.DisconnectThing(id, c, key); err != nil {
					return err
				}
			}
		}

		for _, c := range thing.MFChannels {
			if !contains(intersect, c) {
				if err := bs.sdk.ConnectThing(id, c, key); err != nil {
					return err
				}
			}
		}
	}

	return bs.things.Update(thing)
}

func (bs bootstrapService) List(key string, offset, limit uint64) ([]Thing, error) {
	owner, err := bs.idp.ExtractKey(key)
	if err != nil {
		return []Thing{}, err
	}

	return bs.things.RetrieveAll(owner, offset, limit), nil
}

func (bs bootstrapService) Remove(key, id string) error {
	owner, err := bs.idp.ExtractKey(key)
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

func (bs bootstrapService) Bootstrap(id string) (Config, error) {
	thing, err := bs.things.RetrieveByExternalID(id)
	if err != nil {
		return Config{}, ErrUnauthorizedAccess
	}

	if thing.Status == Active {
		return Config{}, ErrMalformedEntity
	}

	thing.Status = Inactive
	if err := bs.things.ChangeStatus(thing.Owner, thing.ID, thing.Status); err != nil {
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

func (bs bootstrapService) ChangeStatus(key, id string, status Status) error {
	owner, err := bs.idp.ExtractKey(key)
	if err != nil {
		return err
	}

	thing, err := bs.things.RetrieveByID(owner, id)
	if err != nil {
		return err
	}

	if thing.Status == Created {
		return ErrInvalidStatus
	}

	switch status {
	case Active:
		for _, c := range thing.MFChannels {
			if err := bs.sdk.ConnectThing(thing.MFThing, c, key); err != nil {
				return err
			}
		}
	case Inactive:
		for _, c := range thing.MFChannels {
			if err := bs.sdk.DisconnectThing(thing.MFThing, c, key); err != nil {
				return err
			}
		}
	}

	return bs.things.ChangeStatus(owner, id, status)
}

func parseLocation(location string) (string, error) {
	mfPath := strings.Split(location, "/")
	n := len(mfPath)
	if n != 3 {
		return "", ErrInvalidID
	}

	return mfPath[n-1], nil
}

func contains(l []string, e string) bool {
	for _, v := range l {
		if v == e {
			return true
		}
	}

	return false
}
