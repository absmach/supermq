package bootstrap

import (
	"errors"
	"strings"

	mfsdk "github.com/mainflux/mainflux/sdk/go"
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
)

// Service specifies an API that must be fulfilled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	// Add adds new Thing to the user identified by the provided key.
	Add(string, Thing) (Thing, error)

	// View returns Thing with given ID belonging to the user identified by the given key.
	View(string, string) (Thing, error)

	// Remove removes Thing with specified key that belongs to the user identified by the given key.
	Remove(string, string) error

	// Bootstrap returns initial configuration to the Thing with provided external ID.
	Bootstrap(string) error

	// ChangeStatus changes status of the Thing with given ID and owner.
	ChangeStatus(string, string, Status) error
}

var _ Service = (*bootstrapService)(nil)

type bootstrapService struct {
	things ThingRepository
	sdk    mfsdk.SDK
	apiKey string
}

// New returns new Bootstrap service.
func New(things ThingRepository, apiKey string, sdk mfsdk.SDK) Service {
	return &bootstrapService{
		things: things,
		apiKey: apiKey,
		sdk:    sdk,
	}
}

func (bs bootstrapService) Add(key string, thing Thing) (Thing, error) {
	thing.Owner = key
	thing.Status = Created
	id, err := bs.things.Save(thing)
	if err != nil {
		return Thing{}, err
	}
	thing.ID = id
	return thing, nil
}

func (bs bootstrapService) View(id, key string) (Thing, error) {
	return bs.things.RetrieveByID(id, key)
}

func (bs bootstrapService) Remove(id, key string) error {
	return bs.Remove(id, key)
}

func (bs bootstrapService) Bootstrap(externID string) error {
	thing, err := bs.things.RetrieveByExternalID(externID)
	if err != nil {
		return ErrUnauthorizedAccess
	}

	if thing.Status != Created {
		return ErrMalformedEntity
	}

	resp, err := bs.sdk.CreateThing(mfsdk.Thing{Type: "device"}, bs.apiKey)
	if err != nil {
		return err
	}

	mfID, err := parseLocation(resp)
	if err != nil {
		return err
	}

	mfThing, err := bs.sdk.Thing(mfID, bs.apiKey)
	if err != nil {
		return err
	}

	thing.MainfluxID = mfID
	thing.Key = mfThing.Key

	mfChan, err := bs.sdk.CreateChannel(mfsdk.Channel{Name: "NOV"}, bs.apiKey)
	if err != nil {
		return err
	}

	chanID, err := parseLocation(mfChan)
	if err != nil {
		return err
	}

	thing.ChannelID = chanID
	thing.Status = Inactive
	if err := bs.things.Update(thing); err != nil {
		return err
	}

	return nil
}

func (bs bootstrapService) ChangeStatus(id, owner string, status Status) error {
	return bs.things.ChangeStatus(id, owner, status)
}

func parseLocation(location string) (string, error) {
	mfPath := strings.Split(location, "/")
	n := len(mfPath)
	if n != 3 {
		return "", ErrInvalidID
	}
	return mfPath[n-1], nil
}
