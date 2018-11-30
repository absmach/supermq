package bootstrap

import (
	"fmt"
	"net/http"

	"github.com/mainflux/mainflux"
)

// bootstrapRes represent Mainflux Response to the Bootatrap request.
// This is used as a response from ConfigReader and can easily be
// replace with any other response format.
type bootstrapRes struct {
	MFKey        string `json:"key"`
	MQTTUsername string `json:"mf_mqtt_username"`
	MQTTRcvTopic string `json:"mf_mqtt_rcv_topic"`
	MQTTSndTopic string `json:"mf_mqtt_snd_topic"`
	GWID         string `json:"nov_gw_id"`
	Metadata     string `json:"metadata"`
}

func (res bootstrapRes) Code() int {
	return http.StatusOK
}

func (res bootstrapRes) Headers() map[string]string {
	return map[string]string{}
}

func (res bootstrapRes) Empty() bool {
	return false
}

type reader struct{}

// NewConfigReader return new reader which is used to generate response
// from the config.
func NewConfigReader() ConfigReader {
	return reader{}
}

func (r reader) ReadConfig(cfg Config) (mainflux.Response, error) {
	res := bootstrapRes{
		MFKey:        cfg.MFKey,
		GWID:         cfg.MFID,
		MQTTUsername: cfg.MFID,
		MQTTRcvTopic: fmt.Sprintf("channels/%s/messages", cfg.MFID),
		MQTTSndTopic: fmt.Sprintf("channels/%s/messages", cfg.MFID),
		Metadata:     cfg.Metadata,
	}
	return res, nil
}
