// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package things

import (
	"context"

	"github.com/mainflux/mainflux/pkg/errors"

	"github.com/mainflux/mainflux"
)

var (
	// ErrUnauthorizedAccess indicates missing or invalid credentials provided
	// when accessing a protected resource.
	ErrUnauthorizedAccess = errors.New("missing or invalid credentials provided")

	// ErrCreateEntity indicates error in creating entity or entities
	ErrCreateEntity = errors.New("create entity failed")

	// ErrUpdateEntity indicates error in updating entity or entities
	ErrUpdateEntity = errors.New("update entity failed")

	// ErrViewEntity indicates error in viewing entity or entities
	ErrViewEntity = errors.New("view entity failed")

	// ErrRemoveEntity indicates error in removing entity
	ErrRemoveEntity = errors.New("remove entity failed")

	// ErrConnect indicates error in adding connection
	ErrConnect = errors.New("add connection failed")

	// ErrDisconnect indicates error in removing connection
	ErrDisconnect = errors.New("remove connection failed")

	// ErrDisconnect indicates error in database handling
	ErrDB = errors.New("database operation failed")

	// ErrDisconnect indicates error in cache handling
	ErrCache = errors.New("cache operation failed")
)

// Service specifies an API that must be fullfiled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	// CreateThings adds a list of things to the user identified by the provided key.
	CreateThings(ctx context.Context, token string, things ...Thing) ([]Thing, error)

	// UpdateThing updates the thing identified by the provided ID, that
	// belongs to the user identified by the provided key.
	UpdateThing(ctx context.Context, token string, thing Thing) error

	// UpdateKey updates key value of the existing thing. A non-nil error is
	// returned to indicate operation failure.
	UpdateKey(ctx context.Context, token, id, key string) error

	// ViewThing retrieves data about the thing identified with the provided
	// ID, that belongs to the user identified by the provided key.
	ViewThing(ctx context.Context, token, id string) (Thing, error)

	// ListThings retrieves data about subset of things that belongs to the
	// user identified by the provided key.
	ListThings(ctx context.Context, token string, offset, limit uint64, name string, metadata Metadata) (Page, error)

	// ListThingsByChannel retrieves data about subset of things that are
	// connected or not connected to specified channel and belong to the user identified by
	// the provided key.
	ListThingsByChannel(ctx context.Context, token, channel string, offset, limit uint64, connected bool) (Page, error)

	// RemoveThing removes the thing identified with the provided ID, that
	// belongs to the user identified by the provided key.
	RemoveThing(ctx context.Context, token, id string) error

	// CreateChannels adds a list of channels to the user identified by the provided key.
	CreateChannels(ctx context.Context, token string, channels ...Channel) ([]Channel, error)

	// UpdateChannel updates the channel identified by the provided ID, that
	// belongs to the user identified by the provided key.
	UpdateChannel(ctx context.Context, token string, channel Channel) error

	// ViewChannel retrieves data about the channel identified by the provided
	// ID, that belongs to the user identified by the provided key.
	ViewChannel(ctx context.Context, token, id string) (Channel, error)

	// ListChannels retrieves data about subset of channels that belongs to the
	// user identified by the provided key.
	ListChannels(ctx context.Context, token string, offset, limit uint64, name string, m Metadata) (ChannelsPage, error)

	// ListChannelsByThing retrieves data about subset of channels that have
	// specified thing connected or not connected to them and belong to the user identified by
	// the provided key.
	ListChannelsByThing(ctx context.Context, token, thing string, offset, limit uint64, connected bool) (ChannelsPage, error)

	// RemoveChannel removes the thing identified by the provided ID, that
	// belongs to the user identified by the provided key.
	RemoveChannel(ctx context.Context, token, id string) error

	// Connect adds things to the channel's list of connected things.
	Connect(ctx context.Context, token string, chIDs, thIDs []string) error

	// Disconnect removes thing from the channel's list of connected
	// things.
	Disconnect(ctx context.Context, token, chanID, thingID string) error

	// CanAccessByKey determines whether the channel can be accessed using the
	// provided key and returns thing's id if access is allowed.
	CanAccessByKey(ctx context.Context, chanID, key string) (string, error)

	// CanAccessByID determines whether the channel can be accessed by
	// the given thing and returns error if it cannot.
	CanAccessByID(ctx context.Context, chanID, thingID string) error

	// Identify returns thing ID for given thing key.
	Identify(ctx context.Context, key string) (string, error)
}

// PageMetadata contains page metadata that helps navigation.
type PageMetadata struct {
	Total  uint64
	Offset uint64
	Limit  uint64
	Name   string
}

var _ Service = (*thingsService)(nil)

type thingsService struct {
	auth         mainflux.AuthNServiceClient
	things       ThingRepository
	channels     ChannelRepository
	channelCache ChannelCache
	thingCache   ThingCache
	uuidProvider mainflux.UUIDProvider
}

// New instantiates the things service implementation.
func New(auth mainflux.AuthNServiceClient, things ThingRepository, channels ChannelRepository, ccache ChannelCache, tcache ThingCache, up mainflux.UUIDProvider) Service {
	return &thingsService{
		auth:         auth,
		things:       things,
		channels:     channels,
		channelCache: ccache,
		thingCache:   tcache,
		uuidProvider: up,
	}
}

func (ts *thingsService) CreateThings(ctx context.Context, token string, things ...Thing) ([]Thing, error) {
	res, err := ts.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return []Thing{}, errors.Wrap(ErrUnauthorizedAccess, err)
	}

	for i := range things {
		things[i].ID, err = ts.uuidProvider.ID()
		if err != nil {
			return []Thing{}, errors.Wrap(ErrCreateEntity, err)
		}

		things[i].Owner = res.GetValue()

		if things[i].Key == "" {
			things[i].Key, err = ts.uuidProvider.ID()
			if err != nil {
				return []Thing{}, errors.Wrap(ErrCreateEntity, err)
			}
		}
	}

	things, err = ts.things.Save(ctx, things...)
	if err != nil {
		return []Thing{}, errors.Wrap(ErrCreateEntity, errors.Wrap(ErrDB, err))
	}
	return things, nil
}

func (ts *thingsService) UpdateThing(ctx context.Context, token string, thing Thing) error {
	res, err := ts.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return errors.Wrap(ErrUnauthorizedAccess, err)
	}

	thing.Owner = res.GetValue()

	if err = ts.things.Update(ctx, thing); err != nil {
		return errors.Wrap(ErrUpdateEntity, errors.Wrap(ErrDB, err))
	}
	return nil
}

func (ts *thingsService) UpdateKey(ctx context.Context, token, id, key string) error {
	res, err := ts.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return errors.Wrap(ErrUnauthorizedAccess, err)
	}

	owner := res.GetValue()

	if err = ts.things.UpdateKey(ctx, owner, id, key); err != nil {
		return errors.Wrap(ErrUpdateEntity, errors.Wrap(ErrDB, err))
	}
	return nil
}

func (ts *thingsService) ViewThing(ctx context.Context, token, id string) (Thing, error) {
	res, err := ts.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Thing{}, errors.Wrap(ErrUnauthorizedAccess, err)
	}

	thing, err := ts.things.RetrieveByID(ctx, res.GetValue(), id)
	if err != nil {
		return Thing{}, errors.Wrap(ErrViewEntity, errors.Wrap(ErrDB, err))
	}
	return thing, nil
}

func (ts *thingsService) ListThings(ctx context.Context, token string, offset, limit uint64, name string, metadata Metadata) (Page, error) {
	res, err := ts.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Page{}, errors.Wrap(ErrUnauthorizedAccess, err)
	}

	page, err := ts.things.RetrieveAll(ctx, res.GetValue(), offset, limit, name, metadata)
	if err != nil {
		errors.Wrap(ErrViewEntity, errors.Wrap(ErrDB, err))
	}
	return page, nil
}

func (ts *thingsService) ListThingsByChannel(ctx context.Context, token, channel string, offset, limit uint64, connected bool) (Page, error) {
	res, err := ts.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Page{}, errors.Wrap(ErrUnauthorizedAccess, err)
	}

	page, err := ts.things.RetrieveByChannel(ctx, res.GetValue(), channel, offset, limit, connected)
	if err != nil {
		errors.Wrap(ErrViewEntity, errors.Wrap(ErrDB, err))
	}
	return page, nil
}

func (ts *thingsService) RemoveThing(ctx context.Context, token, id string) error {
	res, err := ts.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return errors.Wrap(ErrUnauthorizedAccess, err)
	}

	if err := ts.thingCache.Remove(ctx, id); err != nil {
		return errors.Wrap(ErrRemoveEntity, errors.Wrap(ErrCache, err))
	}
	if err = ts.things.Remove(ctx, res.GetValue(), id); err != nil {
		return errors.Wrap(ErrRemoveEntity, errors.Wrap(ErrDB, err))
	}
	return nil
}

func (ts *thingsService) CreateChannels(ctx context.Context, token string, channels ...Channel) ([]Channel, error) {
	res, err := ts.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return []Channel{}, errors.Wrap(ErrUnauthorizedAccess, err)
	}

	for i := range channels {
		channels[i].ID, err = ts.uuidProvider.ID()
		if err != nil {
			return []Channel{}, errors.Wrap(ErrCreateEntity, err)
		}

		channels[i].Owner = res.GetValue()
	}

	channels, err = ts.channels.Save(ctx, channels...)
	if err != nil {
		return []Channel{}, errors.Wrap(ErrCreateEntity, errors.Wrap(ErrDB, err))
	}
	return channels, nil
}

func (ts *thingsService) UpdateChannel(ctx context.Context, token string, channel Channel) error {
	res, err := ts.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return errors.Wrap(ErrUnauthorizedAccess, err)
	}

	channel.Owner = res.GetValue()
	if err := ts.channels.Update(ctx, channel); err != nil {
		return errors.Wrap(ErrUpdateEntity, errors.Wrap(ErrDB, err))
	}
	return nil
}

func (ts *thingsService) ViewChannel(ctx context.Context, token, id string) (Channel, error) {
	res, err := ts.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Channel{}, errors.Wrap(ErrUnauthorizedAccess, err)
	}

	channel, err := ts.channels.RetrieveByID(ctx, res.GetValue(), id)
	if err != nil {
		errors.Wrap(ErrViewEntity, errors.Wrap(ErrDB, err))
	}
	return channel, nil
}

func (ts *thingsService) ListChannels(ctx context.Context, token string, offset, limit uint64, name string, m Metadata) (ChannelsPage, error) {
	res, err := ts.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return ChannelsPage{}, errors.Wrap(ErrUnauthorizedAccess, err)
	}

	page, err := ts.channels.RetrieveAll(ctx, res.GetValue(), offset, limit, name, m)
	if err != nil {
		errors.Wrap(ErrViewEntity, errors.Wrap(ErrDB, err))
	}
	return page, nil
}

func (ts *thingsService) ListChannelsByThing(ctx context.Context, token, thing string, offset, limit uint64, connected bool) (ChannelsPage, error) {
	res, err := ts.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return ChannelsPage{}, errors.Wrap(ErrUnauthorizedAccess, err)
	}

	page, err := ts.channels.RetrieveByThing(ctx, res.GetValue(), thing, offset, limit, connected)
	if err != nil {
		errors.Wrap(ErrViewEntity, errors.Wrap(ErrDB, err))
	}
	return page, nil
}

func (ts *thingsService) RemoveChannel(ctx context.Context, token, id string) error {
	res, err := ts.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return errors.Wrap(ErrUnauthorizedAccess, err)
	}

	if err := ts.channelCache.Remove(ctx, id); err != nil {
		return errors.Wrap(ErrRemoveEntity, errors.Wrap(ErrCache, err))
	}

	if err := ts.channels.Remove(ctx, res.GetValue(), id); err != nil {
		return errors.Wrap(ErrRemoveEntity, errors.Wrap(ErrDB, err))
	}
	return nil
}

func (ts *thingsService) Connect(ctx context.Context, token string, chIDs, thIDs []string) error {
	res, err := ts.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return errors.Wrap(ErrUnauthorizedAccess, err)
	}

	if err := ts.channels.Connect(ctx, res.GetValue(), chIDs, thIDs); err != nil {
		return errors.Wrap(ErrConnect, errors.Wrap(ErrDB, err))
	}
	return nil
}

func (ts *thingsService) Disconnect(ctx context.Context, token, chanID, thingID string) error {
	res, err := ts.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return errors.Wrap(ErrUnauthorizedAccess, err)
	}

	if err := ts.channelCache.Disconnect(ctx, chanID, thingID); err != nil {
		return errors.Wrap(ErrDisconnect, errors.Wrap(ErrCache, err))
	}

	if err := ts.channels.Disconnect(ctx, res.GetValue(), chanID, thingID); err != nil {
		return errors.Wrap(ErrDisconnect, errors.Wrap(ErrDB, err))
	}
	return nil
}

func (ts *thingsService) CanAccessByKey(ctx context.Context, chanID, key string) (string, error) {
	thingID, err := ts.hasThing(ctx, chanID, key)
	if err == nil {
		return thingID, nil
	}

	thingID, err = ts.channels.HasThing(ctx, chanID, key)
	if err != nil {
		return "", errors.Wrap(ErrUnauthorizedAccess, errors.Wrap(ErrDB, err))
	}

	if err := ts.thingCache.Save(ctx, key, thingID); err != nil {
		return "", errors.Wrap(ErrUnauthorizedAccess, errors.Wrap(ErrCache, err))
	}
	if err := ts.channelCache.Connect(ctx, chanID, thingID); err != nil {
		return "", errors.Wrap(ErrUnauthorizedAccess, errors.Wrap(ErrCache, err))
	}
	return thingID, nil
}

func (ts *thingsService) CanAccessByID(ctx context.Context, chanID, thingID string) error {
	if connected := ts.channelCache.HasThing(ctx, chanID, thingID); connected {
		return nil
	}

	if err := ts.channels.HasThingByID(ctx, chanID, thingID); err != nil {
		return errors.Wrap(ErrUnauthorizedAccess, errors.Wrap(ErrDB, err))
	}

	if err := ts.channelCache.Connect(ctx, chanID, thingID); err != nil {
		return errors.Wrap(ErrUnauthorizedAccess, errors.Wrap(ErrCache, err))
	}
	return nil
}

func (ts *thingsService) Identify(ctx context.Context, key string) (string, error) {
	id, err := ts.thingCache.ID(ctx, key)
	if err == nil {
		return id, nil
	}

	id, err = ts.things.RetrieveByKey(ctx, key)
	if err != nil {
		return "", errors.Wrap(ErrUnauthorizedAccess, errors.Wrap(ErrDB, err))
	}

	if err := ts.thingCache.Save(ctx, key, id); err != nil {
		return "", errors.Wrap(ErrUnauthorizedAccess, errors.Wrap(ErrCache, err))
	}
	return id, nil
}

func (ts *thingsService) hasThing(ctx context.Context, chanID, key string) (string, error) {
	thingID, err := ts.thingCache.ID(ctx, key)
	if err != nil {
		return "", errors.Wrap(ErrUnauthorizedAccess, err)
	}

	if connected := ts.channelCache.HasThing(ctx, chanID, thingID); !connected {
		return "", errors.Wrap(ErrUnauthorizedAccess, errors.Wrap(ErrCache, err))
	}

	return thingID, nil
}
