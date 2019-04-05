package producer

import (
	"strings"
	"time"

	"github.com/mainflux/mainflux/bootstrap"
)

const (
	configPrefix = "config."
	configCreate = configPrefix + "create"
	configUpdate = configPrefix + "update"
	configRemove = configPrefix + "remove"

	thingPrefix      = "thing."
	thingStateChange = thingPrefix + "state_change"
	thingBootstrap   = thingPrefix + "bootstrap"
)

type event interface {
	encode() map[string]interface{}
}

var (
	_ event = (*createConfigEvent)(nil)
	_ event = (*updateConfigEvent)(nil)
	_ event = (*removeConfigEvent)(nil)
	_ event = (*bootstrapEvent)(nil)
	_ event = (*changeStateEvent)(nil)
)

type createConfigEvent struct {
	mfThing    string
	owner      string
	name       string
	mfChannels []string
	externalID string
	content    string
	timestamp  time.Time
}

func (cce createConfigEvent) encode() map[string]interface{} {
	return map[string]interface{}{
		"id":         cce.mfThing,
		"owner":      cce.owner,
		"name":       cce.name,
		"channels":   strings.Join(cce.mfChannels, ", "),
		"externalID": cce.externalID,
		"content":    cce.content,
		"timestamp":  cce.timestamp.Unix(),
		"operation":  configCreate,
	}
}

type updateConfigEvent struct {
	mfThing   string
	name      string
	content   string
	timestamp time.Time
}

func (uce updateConfigEvent) encode() map[string]interface{} {
	return map[string]interface{}{
		"id":        uce.mfThing,
		"name":      uce.name,
		"content":   uce.content,
		"timestamp": uce.timestamp.Unix(),
		"operation": configUpdate,
	}
}

type removeConfigEvent struct {
	mfThing   string
	timestamp time.Time
}

func (rce removeConfigEvent) encode() map[string]interface{} {
	return map[string]interface{}{
		"id":        rce.mfThing,
		"timestamp": rce.timestamp.Unix(),
		"operation": configRemove,
	}
}

type bootstrapEvent struct {
	externalID string
	timestamp  time.Time
}

func (be bootstrapEvent) encode() map[string]interface{} {
	return map[string]interface{}{
		"externalID": be.externalID,
		"timestamp":  be.timestamp.Unix(),
		"operation":  thingBootstrap,
	}
}

type changeStateEvent struct {
	externalID string
	state      bootstrap.State
	timestamp  time.Time
}

func (cse changeStateEvent) encode() map[string]interface{} {
	return map[string]interface{}{
		"externalID": cse.externalID,
		"state":      cse.state.String(),
		"timestamp":  cse.timestamp.Unix(),
		"operation":  thingStateChange,
	}
}
