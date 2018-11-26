package redis

import (
	"github.com/go-redis/redis"
	"github.com/mainflux/mainflux/lora"
)

const (
	group  = "mainflux.lora"
	stream = "mainflux.things"

	thingPrefix     = "thing."
	thingCreate     = thingPrefix + "create"
	thingUpdate     = thingPrefix + "update"
	thingRemove     = thingPrefix + "remove"
	thingConnect    = thingPrefix + "connect"
	thingDisconnect = thingPrefix + "disconnect"

	channelPrefix = "channel."
	channelCreate = channelPrefix + "create"
	channelUpdate = channelPrefix + "update"
	channelRemove = channelPrefix + "remove"
)

// EventStore represents event source for things and channels provisioning.
type EventStore interface {
	// Subscribes to geven subject and receives events.
	Subscribe(string)
}

type eventStore struct {
	client   *redis.Client
	consumer string
	svc      lora.Service
}

// NewEventStore returns new event store instance.
func NewEventStore(client *redis.Client, consumer string) EventStore {
	return eventStore{
		client:   client,
		consumer: consumer,
	}
}

func (es eventStore) Subscribe(subject string) {
	es.client.XGroupCreate(stream, group, "$").Err()
	for {
		streams, err := es.client.XReadGroup(&redis.XReadGroupArgs{
			Group:    group,
			Consumer: es.consumer,
			Streams:  []string{stream, ">"},
			Count:    100,
		}).Result()
		if err != nil || len(streams) == 0 {
			continue
		}

		for _, msg := range streams[0].Messages {
			event := msg.Values

			var err error
			switch event["operation"] {
			case thingCreate:
				cte := decodeCreateThing(event)
				err = es.handleCreateThing(cte)
			case thingUpdate:
				ute := decodeUpdateThing(event)
				err = es.handleUpdateThing(ute)
			case thingRemove:
				rte := decodeRemoveThing(event)
				err = es.handleRemoveThing(rte)
			case channelCreate:
				cce := decodeCreateChannel(event)
				err = es.handleCreateChannel(cce)
			case channelUpdate:
				uce := decodeUpdateChannel(event)
				err = es.handleUpdateChannel(uce)
			case channelRemove:
				rce := decodeRemoveChannel(event)
				err = es.handleRemoveChannel(rce)
			case thingConnect:
				cte := decodeConnectThing(event)
				err = es.handleConnect(cte)
			case thingDisconnect:
				dte := decodeDisconnectThing(event)
				err = es.handleDisconnect(dte)
			}
			if err != nil {
				// TODO: add error logging
				break
			}
			es.client.XAck(stream, group, msg.ID)
		}
	}
}

func decodeCreateThing(event map[string]interface{}) createThingEvent {
	return createThingEvent{
		id:       event["id"].(string),
		kind:     event["type"].(string),
		name:     event["name"].(string),
		owner:    event["owner"].(string),
		metadata: event["metadata"].(string),
	}
}

func decodeUpdateThing(event map[string]interface{}) updateThingEvent {
	return updateThingEvent{
		id:       event["id"].(string),
		kind:     event["type"].(string),
		name:     event["name"].(string),
		metadata: event["metadata"].(string),
	}
}

func decodeRemoveThing(event map[string]interface{}) removeThingEvent {
	return removeThingEvent{
		id: event["id"].(string),
	}
}

func decodeCreateChannel(event map[string]interface{}) createChannelEvent {
	return createChannelEvent{
		id:    event["id"].(string),
		owner: event["owner"].(string),
		name:  event["name"].(string),
	}
}

func decodeUpdateChannel(event map[string]interface{}) updateChannelEvent {
	return updateChannelEvent{
		id:   event["id"].(string),
		name: event["name"].(string),
	}
}

func decodeRemoveChannel(event map[string]interface{}) removeChannelEvent {
	return removeChannelEvent{
		id: event["id"].(string),
	}
}

func decodeConnectThing(event map[string]interface{}) connectThingEvent {
	return connectThingEvent{
		thingID: event["thing_id"].(string),
		chanID:  event["chan_id"].(string),
	}
}

func decodeDisconnectThing(event map[string]interface{}) disconnectThingEvent {
	return disconnectThingEvent{
		thingID: event["thing_id"].(string),
		chanID:  event["chan_id"].(string),
	}
}

func (es eventStore) handleCreateThing(cte createThingEvent) error {
	// TODO: es.svc.CreateThing()
	return nil
}

func (es eventStore) handleUpdateThing(ute updateThingEvent) error {
	// TODO: es.svc.UpdateThing()
	return nil
}

func (es eventStore) handleRemoveThing(rte removeThingEvent) error {
	// TODO: es.svc.RemoveThing()
	return nil
}

func (es eventStore) handleCreateChannel(cce createChannelEvent) error {
	// TODO: es.svc.CreateChannel()
	return nil
}

func (es eventStore) handleUpdateChannel(uce updateChannelEvent) error {
	// TODO: es.svc.UpdateChannel()
	return nil
}

func (es eventStore) handleRemoveChannel(rce removeChannelEvent) error {
	// TODO: es.svc.RemoveChannel()
	return nil
}

func (es eventStore) handleConnect(cte connectThingEvent) error {
	// TODO: es.svc.Connect()
	return nil
}

func (es eventStore) handleDisconnect(dte disconnectThingEvent) error {
	// TODO: es.svc.Disconnect()
	return nil
}
