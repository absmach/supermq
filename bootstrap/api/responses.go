package api

import (
	"fmt"
	"net/http"

	"github.com/mainflux/mainflux/bootstrap"

	"github.com/mainflux/mainflux"
)

var (
	_ mainflux.Response = (*identityRes)(nil)
	_ mainflux.Response = (*removeRes)(nil)
	_ mainflux.Response = (*configRes)(nil)
	_ mainflux.Response = (*stateRes)(nil)
	_ mainflux.Response = (*viewRes)(nil)
	_ mainflux.Response = (*listRes)(nil)
)

type identityRes struct {
	id uint64
}

func (res identityRes) Headers() map[string]string {
	return map[string]string{
		"X-thing-id": fmt.Sprint(res.id),
	}
}

func (res identityRes) Code() int {
	return http.StatusOK
}

func (res identityRes) Empty() bool {
	return true
}

type removeRes struct{}

func (res removeRes) Code() int {
	return http.StatusNoContent
}

func (res removeRes) Headers() map[string]string {
	return map[string]string{}
}

func (res removeRes) Empty() bool {
	return true
}

type configRes struct {
	id      string
	created bool
}

func (res configRes) Code() int {
	if res.created {
		return http.StatusCreated
	}

	return http.StatusOK
}

func (res configRes) Headers() map[string]string {
	if res.created {
		return map[string]string{
			"Location": fmt.Sprintf("/configs/%s", res.id),
		}
	}

	return map[string]string{}
}

func (res configRes) Empty() bool {
	return true
}

type viewRes struct {
	ID          string          `json:"id,omitempty"`
	MFKey       string          `json:"mainflux_key,omitempty"`
	MFThing     string          `json:"mainflux_id,omitempty"`
	MFChannels  []string        `json:"mainflux_channels,omitempty"`
	ExternalID  string          `json:"external_id"`
	ExternalKey string          `json:"external_key,omitempty"`
	Content     string          `json:"content,omitempty"`
	State       bootstrap.State `json:"state,omitempty"`
}

func (res viewRes) Code() int {
	return http.StatusOK
}

func (res viewRes) Headers() map[string]string {
	return map[string]string{}
}

func (res viewRes) Empty() bool {
	return false
}

type listRes struct {
	Configs []viewRes `json:"configs"`
}

func (res listRes) Code() int {
	return http.StatusOK
}

func (res listRes) Headers() map[string]string {
	return map[string]string{}
}

func (res listRes) Empty() bool {
	return false
}

type bootstrapRes struct {
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

type stateRes struct {
	State bootstrap.State `json:"state"`
}

func (res stateRes) Code() int {
	return http.StatusOK
}

func (res stateRes) Headers() map[string]string {
	return map[string]string{}
}

func (res stateRes) Empty() bool {
	return false
}
