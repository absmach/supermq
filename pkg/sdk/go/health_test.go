package sdk_test

import (
	"fmt"
	"testing"

	sdk "github.com/mainflux/mainflux/pkg/sdk/go"
	"github.com/stretchr/testify/assert"
)

func TestHealth(t *testing.T) {
	svc := newThingsService(map[string]string{token: email})
	ts := newThingsServer(svc)
	defer ts.Close()

	sdkConf := sdk.Config{
		ThingsURL:       ts.URL,
		MsgContentType:  contentType,
		TLSVerification: false,
	}

	mainfluxSDK := sdk.NewSDK(sdkConf)
	cases := map[string]struct {
		empty bool
		err   error
	}{
		"get things service version": {
			empty: false,
			err:   nil,
		},
	}
	for desc, tc := range cases {
		h, err := mainfluxSDK.Health()
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", desc, tc.err, err))
		assert.Equal(t, tc.empty, h.Version == "", fmt.Sprintf("%s: expected non-empty result version, got %s", desc, h.Version))
	}
}
