package producer

import "github.com/mainflux/mainflux/bootstrap"

type createConfigEvent struct {
	mfThing    string
	owner      string
	name       string
	mfChannels []string
	externalID string
	content    string
}

func (cce createConfigEvent) Encode() map[string]interface{} {
	return map[string]interface{}{
		"id":         cce.mfThing,
		"owner":      cce.owner,
		"name":       cce.name,
		"channels":   cce.mfChannels,
		"externalID": cce.externalID,
		"content":    cce.content,
	}
}

type updateConfigEvent struct {
	mfThing string
	name    string
	content string
}

func (uce updateConfigEvent) Encode() map[string]interface{} {
	return map[string]interface{}{
		"id":      uce.mfThing,
		"name":    uce.name,
		"content": uce.content,
	}
}

type removeConfigEvent struct {
	mfThing string
}

func (rce removeConfigEvent) Encode() map[string]interface{} {
	return map[string]interface{}{
		"id": rce.mfThing,
	}
}

type bootstrapEvent struct {
	externalID string
}

func (be bootstrapEvent) Encode() map[string]interface{} {
	return map[string]interface{}{
		"externalID": be.externalID,
	}
}

type changeStateEvent struct {
	externalID string
	state      bootstrap.State
}

func (cse changeStateEvent) Encode() map[string]interface{} {
	return map[string]interface{}{
		"externalID": cse.externalID,
		"state":      cse.state.String(),
	}
}
