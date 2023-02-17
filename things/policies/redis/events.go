package redis

import "encoding/json"

const (
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

type event interface {
	Encode() map[string]interface{}
}

var (
	_ event = (*createChannelEvent)(nil)
	_ event = (*updateChannelEvent)(nil)
	_ event = (*removeChannelEvent)(nil)
	_ event = (*connectThingEvent)(nil)
	_ event = (*disconnectThingEvent)(nil)
)

type createChannelEvent struct {
	id       string
	owner    string
	name     string
	metadata map[string]interface{}
}

func (cce createChannelEvent) Encode() map[string]interface{} {
	val := map[string]interface{}{
		"id":        cce.id,
		"owner":     cce.owner,
		"operation": channelCreate,
	}

	if cce.name != "" {
		val["name"] = cce.name
	}

	if cce.metadata != nil {
		metadata, err := json.Marshal(cce.metadata)
		if err != nil {
			return val
		}

		val["metadata"] = string(metadata)
	}

	return val
}

type updateChannelEvent struct {
	id       string
	name     string
	metadata map[string]interface{}
}

func (uce updateChannelEvent) Encode() map[string]interface{} {
	val := map[string]interface{}{
		"id":        uce.id,
		"operation": channelUpdate,
	}

	if uce.name != "" {
		val["name"] = uce.name
	}

	if uce.metadata != nil {
		metadata, err := json.Marshal(uce.metadata)
		if err != nil {
			return val
		}

		val["metadata"] = string(metadata)
	}

	return val
}

type removeChannelEvent struct {
	id string
}

func (rce removeChannelEvent) Encode() map[string]interface{} {
	return map[string]interface{}{
		"id":        rce.id,
		"operation": channelRemove,
	}
}

type connectThingEvent struct {
	chanID  string
	thingID string
}

func (cte connectThingEvent) Encode() map[string]interface{} {
	return map[string]interface{}{
		"chan_id":   cte.chanID,
		"thing_id":  cte.thingID,
		"operation": thingConnect,
	}
}

type disconnectThingEvent struct {
	chanID  string
	thingID string
}

func (dte disconnectThingEvent) Encode() map[string]interface{} {
	return map[string]interface{}{
		"chan_id":   dte.chanID,
		"thing_id":  dte.thingID,
		"operation": thingDisconnect,
	}
}
