package redis

import "encoding/json"

const (
	thingPrefix = "thing."
	thingCreate = thingPrefix + "create"
	thingUpdate = thingPrefix + "update"
	thingRemove = thingPrefix + "remove"
)

type event interface {
	Encode() map[string]interface{}
}

var (
	_ event = (*createClientEvent)(nil)
	_ event = (*updateClientEvent)(nil)
	_ event = (*removeClientEvent)(nil)
)

type createClientEvent struct {
	id       string
	owner    string
	name     string
	metadata map[string]interface{}
}

func (cte createClientEvent) Encode() map[string]interface{} {
	val := map[string]interface{}{
		"id":        cte.id,
		"owner":     cte.owner,
		"operation": thingCreate,
	}

	if cte.name != "" {
		val["name"] = cte.name
	}

	if cte.metadata != nil {
		metadata, err := json.Marshal(cte.metadata)
		if err != nil {
			return val
		}

		val["metadata"] = string(metadata)
	}

	return val
}

type updateClientEvent struct {
	id       string
	name     string
	owner    string
	tags     []string
	metadata map[string]interface{}
}

func (ute updateClientEvent) Encode() map[string]interface{} {
	val := map[string]interface{}{
		"id":        ute.id,
		"operation": thingUpdate,
	}

	if ute.name != "" {
		val["name"] = ute.name
	}

	if ute.metadata != nil {
		metadata, err := json.Marshal(ute.metadata)
		if err != nil {
			return val
		}

		val["metadata"] = string(metadata)
	}

	return val
}

type removeClientEvent struct {
	id string
}

func (rte removeClientEvent) Encode() map[string]interface{} {
	return map[string]interface{}{
		"id":        rte.id,
		"operation": thingRemove,
	}
}
