package influxdb

import (
	"github.com/mainflux/mainflux/pkg/transformers/senml"
	"github.com/mainflux/mainflux/things"
)

type tags map[string]string

func interface2String(inter interface{}) string {
	switch inter := inter.(type) {
	case string:
		return inter
	}
	return ""
}

func senmlTags(msg senml.Message, deviceName string, meta things.Metadata) tags {
	return tags{
		"channel":   msg.Channel,
		"subtopic":  msg.Subtopic,
		"publisher": msg.Publisher,
		// Customized tags
		"device": deviceName,
		// Thing meta
		"system":         interface2String(meta["system-name"]),
		"building-group": interface2String(meta["building-group"]),
		"building":       interface2String(meta["building"]),
		"level":          interface2String(meta["level"]),
		"room":           interface2String(meta["room"]),
		"space":          interface2String(meta["space"]),
	}
}

// TODO: bring back json support
// func jsonTags(msg json.Message) tags {
// 	return tags{
// 		"channel":   msg.Channel,
// 		"subtopic":  msg.Subtopic,
// 		"publisher": msg.Publisher,
// 	}
// }
