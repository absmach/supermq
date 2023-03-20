package homing

import (
	"testing"
)

func TestGetIp(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		if _, err := getIP(); err != nil {
			t.Errorf(err.Error())
		}
	})
}

func TestSendTelemetry(t *testing.T) {
	var data telemetryData
	data.Service = "hs.serviceName"
	data.Version = "hs.version"
	err := sendTelemetry(&data)
	if err != nil {
		t.Error(err)
	}

}
