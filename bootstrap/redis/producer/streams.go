package producer

import (
	"github.com/go-redis/redis"
	"github.com/mainflux/mainflux/bootstrap"
)

const (
	stream = "mainflux.things"
	group  = "mainflux.bootstrap"

	thingPrefix     = "thing."
	thingRemove     = thingPrefix + "remove"
	thingDisconnect = thingPrefix + "disconnect"

	channelPrefix = "channel."
	channelUpdate = channelPrefix + "update"
	channelRemove = channelPrefix + "remove"

	exists = "BUSYGROUP Consumer Group name already exists"
)

const (
	streamID  = "mainflux.bootstrap"
	streamLen = 1000
)

var _ bootstrap.Service = (*eventStore)(nil)

type eventStore struct {
	svc    bootstrap.Service
	client *redis.Client
}

// NewEventStoreMiddleware returns wrapper around things service that sends
// events to event store.
func NewEventStoreMiddleware(svc bootstrap.Service, client *redis.Client) bootstrap.Service {
	return eventStore{
		svc:    svc,
		client: client,
	}
}

func (es eventStore) Add(key string, cfg bootstrap.Config) (bootstrap.Config, error) {
	saved, err := es.svc.Add(key, cfg)
	if err != nil {
		return saved, err
	}

	var channels []string
	for _, ch := range saved.MFChannels {
		channels = append(channels, ch.ID)
	}

	event := createConfigEvent{
		mfThing:    saved.MFThing,
		owner:      saved.Owner,
		name:       saved.Name,
		mfChannels: channels,
		externalID: saved.ExternalID,
		content:    saved.Content,
	}
	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       event.Encode(),
	}
	es.client.XAdd(record).Err()

	return saved, err
}

func (es eventStore) View(key, id string) (bootstrap.Config, error) {
	return es.svc.View(key, id)
}

func (es eventStore) Update(key string, cfg bootstrap.Config) error {
	if err := es.svc.Update(key, cfg); err != nil {
		return err
	}

	event := updateConfigEvent{
		mfThing: cfg.MFThing,
		name:    cfg.Name,
		content: cfg.Content,
	}

	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       event.Encode(),
	}

	es.client.XAdd(record).Err()

	return nil
}

func (es eventStore) UpdateConnections(key, id string, connections []string) error {
	return es.svc.UpdateConnections(key, id, connections)
}

func (es eventStore) List(key string, filter bootstrap.Filter, offset, limit uint64) (bootstrap.ConfigsPage, error) {
	return es.svc.List(key, filter, offset, limit)
}

func (es eventStore) Remove(key, id string) error {
	if err := es.svc.Remove(key, id); err != nil {
		return err
	}

	event := removeConfigEvent{
		mfThing: id,
	}

	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       event.Encode(),
	}

	es.client.XAdd(record).Err()

	return nil
}

func (es eventStore) Bootstrap(externalKey, externalID string) (bootstrap.Config, error) {
	cfg, err := es.svc.Bootstrap(externalKey, externalID)
	if err != nil {
		return bootstrap.Config{}, err
	}

	event := bootstrapEvent{
		externalID: cfg.ExternalID,
	}

	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       event.Encode(),
	}

	es.client.XAdd(record).Err()

	return cfg, nil
}

func (es eventStore) ChangeState(key, id string, state bootstrap.State) error {
	if err := es.svc.ChangeState(key, id, state); err != nil {
		return err
	}

	event := changeStateEvent{
		externalID: id,
		state:      state,
	}

	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       event.Encode(),
	}

	es.client.XAdd(record).Err()

	return nil
}

func (es eventStore) UpdateChannelHandler(channel bootstrap.Channel) error {
	return es.UpdateChannelHandler(channel)
}

func (es eventStore) RemoveConfigHandler(id string) error {
	return es.svc.RemoveConfigHandler(id)
}

func (es eventStore) RemoveChannelHandler(id string) error {
	return es.svc.RemoveChannelHandler(id)
}

func (es eventStore) DisconnectThingHandler(channelID, thingID string) error {
	return es.svc.DisconnectThingHandler(channelID, thingID)
}
