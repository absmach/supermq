package homing

import (
	"testing"
)

func TestGetIp(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		for _, endpoint := range ipEndpoints {
			if _, err := getIP(endpoint); err != nil {
				t.Errorf(err.Error())
			}
		}
	})
}
