package influxdb

import (
	"errors"
	"strings"

	"github.com/mainflux/mainflux/pkg/transformers/senml"
)

func senmlBasename(msg senml.Message) (deviceName string, measurement string, err error) {
	splitted := strings.Split(msg.Name, ":")
	if len(splitted) == 2 {
		deviceName = splitted[0]
		measurement = splitted[1]
	} else {
		err = errors.New("malformed message name")
	}
	return
}
