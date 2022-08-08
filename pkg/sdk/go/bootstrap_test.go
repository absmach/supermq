package sdk_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/bootstrap"
	bsapi "github.com/mainflux/mainflux/bootstrap/api"
	"github.com/mainflux/mainflux/bootstrap/mocks"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/errors"
	mfsdk "github.com/mainflux/mainflux/pkg/sdk/go"
	sdk "github.com/mainflux/mainflux/pkg/sdk/go"
	"github.com/mainflux/mainflux/things"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	unknown    = "unknown"
	wrongID    = "wrong_id"
	clientCert = "newCert"
	clientKey  = "newKey"
	caCert     = "newCert"
)

var (
	validToken  = "validToken"
	channelsNum = 3
	encKey      = []byte("1234567891011121")
	channel     = mfsdk.Channel{
		ID:       "1",
		Name:     "name",
		Metadata: map[string]interface{}{"name": "value"},
	}
	config = mfsdk.BootstrapConfig{
		ExternalID:  "external_id",
		ExternalKey: "external_key",
		MFChannels:  []mfsdk.Channel{channel},
		Content:     "config",
		MFKey:       "mfKey",
		State:       0,
	}
)

func generateChannels() map[string]things.Channel {
	channels := make(map[string]things.Channel, channelsNum)
	for i := 0; i < channelsNum; i++ {
		id := strconv.Itoa(i + 1)
		channels[id] = things.Channel{
			ID:       id,
			Owner:    email,
			Metadata: metadata,
		}
	}
	return channels
}

func newBootstrapThingsService(auth mainflux.AuthServiceClient) things.Service {
	return mocks.NewThingsService(map[string]things.Thing{}, generateChannels(), auth)
}

func newBootstrapService(auth mainflux.AuthServiceClient, url string) bootstrap.Service {
	things := mocks.NewConfigsRepository()
	config := mfsdk.Config{
		ThingsURL: url,
	}
	sdk := mfsdk.NewSDK(config)
	return bootstrap.New(auth, things, sdk, encKey)
}

func newBootstrapServer(svc bootstrap.Service) *httptest.Server {
	logger := logger.NewMock()
	mux := bsapi.MakeHandler(svc, bootstrap.NewConfigReader(encKey), logger)
	return httptest.NewServer(mux)
}
func TestAddBootstrap(t *testing.T) {
	auth := mocks.NewAuthClient(map[string]string{token: token})

	ts := newThingsServer(newBootstrapThingsService(auth))
	svc := newBootstrapService(auth, ts.URL)
	bs := newBootstrapServer(svc)
	defer bs.Close()

	sdkConf := mfsdk.Config{
		BootstrapURL:    bs.URL,
		MsgContentType:  contentType,
		TLSVerification: true,
	}
	mainfluxSDK := mfsdk.NewSDK(sdkConf)

	cases := []struct {
		desc   string
		config mfsdk.BootstrapConfig
		auth   string
		err    error
	}{

		{
			desc:   "add a config with invalid token",
			config: config,
			auth:   invalidToken,
			err:    createError(sdk.ErrFailedCreation, http.StatusUnauthorized),
		},
		{
			desc:   "add a config with empty token",
			config: config,
			auth:   "",
			err:    createError(sdk.ErrFailedCreation, http.StatusUnauthorized),
		},
		{
			desc:   "add a config with invalid config",
			config: mfsdk.BootstrapConfig{},
			auth:   token,
			err:    createError(sdk.ErrFailedCreation, http.StatusBadRequest),
		},
		{
			desc:   "add a valid config",
			config: config,
			auth:   token,
			err:    nil,
		},
		{
			desc:   "add an existing config",
			config: config,
			auth:   token,
			err:    createError(sdk.ErrFailedCreation, http.StatusConflict),
		},
	}
	for _, tc := range cases {
		_, err := mainfluxSDK.AddBootstrap(tc.auth, tc.config)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
	}
}

func TestWhitelist(t *testing.T) {
	auth := mocks.NewAuthClient(map[string]string{token: token})

	ts := newThingsServer(newBootstrapThingsService(auth))
	svc := newBootstrapService(auth, ts.URL)
	bs := newBootstrapServer(svc)
	defer bs.Close()

	sdkConf := mfsdk.Config{
		BootstrapURL:    bs.URL,
		MsgContentType:  contentType,
		TLSVerification: true,
	}
	mainfluxSDK := mfsdk.NewSDK(sdkConf)

	mfThingID, err := mainfluxSDK.AddBootstrap(token, config)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	updtConfig := config
	updtConfig.MFThing = mfThingID
	wrongConfig := config
	wrongConfig.MFThing = wrongID

	cases := []struct {
		desc   string
		config mfsdk.BootstrapConfig
		auth   string
		err    error
	}{
		{
			desc:   "change state with invalid token",
			config: updtConfig,
			auth:   invalidToken,
			err:    createError(sdk.ErrFailedWhitelist, http.StatusUnauthorized),
		},
		{
			desc:   "change state with empty token",
			config: updtConfig,
			auth:   "",
			err:    createError(sdk.ErrFailedWhitelist, http.StatusUnauthorized),
		},
		{
			desc:   "change state of non-existing config",
			config: wrongConfig,
			auth:   token,
			err:    createError(sdk.ErrFailedWhitelist, http.StatusNotFound),
		},
		{
			desc:   "change state to active",
			config: updtConfig,
			auth:   token,
			err:    nil,
		},
		{
			desc:   "change state to current state",
			config: updtConfig,
			auth:   token,
			err:    nil,
		},
	}
	for _, tc := range cases {
		err := mainfluxSDK.Whitelist(tc.auth, tc.config)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
	}
}

func TestViewBootstrap(t *testing.T) {
	auth := mocks.NewAuthClient(map[string]string{token: token})

	ts := newThingsServer(newBootstrapThingsService(auth))
	svc := newBootstrapService(auth, ts.URL)
	bs := newBootstrapServer(svc)
	defer bs.Close()

	sdkConf := mfsdk.Config{
		BootstrapURL:    bs.URL,
		MsgContentType:  contentType,
		TLSVerification: true,
	}
	mainfluxSDK := mfsdk.NewSDK(sdkConf)

	thingID, err := mainfluxSDK.AddBootstrap(token, config)
	require.Nil(t, err, fmt.Sprintf("Saving config expected to succeed: %s.\n", err))

	cases := []struct {
		desc string
		id   string
		auth string
		err  error
	}{
		{
			desc: "view a non-existing config",
			id:   unknown,
			auth: token,
			err:  createError(sdk.ErrFailedFetch, http.StatusNotFound),
		},
		{
			desc: "view a config with invalid token",
			id:   thingID,
			auth: invalidToken,
			err:  createError(sdk.ErrFailedFetch, http.StatusUnauthorized),
		},
		{
			desc: "view a config with empty token",
			id:   thingID,
			auth: "",
			err:  createError(sdk.ErrFailedFetch, http.StatusUnauthorized),
		},
		{
			desc: "view an existing config",
			id:   thingID,
			auth: token,
			err:  nil,
		},
	}
	for _, tc := range cases {
		_, err := mainfluxSDK.ViewBootstrap(tc.auth, tc.id)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
	}
}

func TestUpdateBootstrap(t *testing.T) {
	auth := mocks.NewAuthClient(map[string]string{token: email})

	ts := newThingsServer(newBootstrapThingsService(auth))
	svc := newBootstrapService(auth, ts.URL)
	bs := newBootstrapServer(svc)
	defer bs.Close()

	sdkConf := mfsdk.Config{
		BootstrapURL:    bs.URL,
		MsgContentType:  contentType,
		TLSVerification: true,
	}
	mainfluxSDK := mfsdk.NewSDK(sdkConf)

	thingID, err := mainfluxSDK.AddBootstrap(token, config)
	require.Nil(t, err, fmt.Sprintf("Saving config expected to succeed: %s.\n", err))

	updatedConfig := config
	updatedConfig.MFThing = thingID
	ch := channel
	ch.ID = "2"
	updatedConfig.MFChannels = append(updatedConfig.MFChannels, ch)
	nonExisting := config
	nonExisting.MFThing = unknown

	cases := []struct {
		desc   string
		auth   string
		config mfsdk.BootstrapConfig
		err    error
	}{
		{
			desc:   "update with invalid token",
			auth:   invalidToken,
			config: updatedConfig,
			err:    createError(sdk.ErrFailedUpdate, http.StatusUnauthorized),
		},
		{
			desc:   "update with empty token",
			auth:   "",
			config: updatedConfig,
			err:    createError(sdk.ErrFailedUpdate, http.StatusUnauthorized),
		},
		{
			desc:   "update a non-existing config",
			auth:   token,
			config: nonExisting,
			err:    createError(sdk.ErrFailedUpdate, http.StatusNotFound),
		},
		{
			desc:   "update a config with state created",
			auth:   token,
			config: updatedConfig,
			err:    nil,
		},
	}
	for _, tc := range cases {
		err := mainfluxSDK.UpdateBootstrap(tc.auth, tc.config)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
	}
}

func TestUpdateBootstrapCerts(t *testing.T) {
	auth := mocks.NewAuthClient(map[string]string{token: email})

	ts := newThingsServer(newBootstrapThingsService(auth))
	svc := newBootstrapService(auth, ts.URL)
	bs := newBootstrapServer(svc)
	defer bs.Close()

	sdkConf := mfsdk.Config{
		BootstrapURL:    bs.URL,
		MsgContentType:  contentType,
		TLSVerification: true,
	}
	mainfluxSDK := mfsdk.NewSDK(sdkConf)

	_, err := mainfluxSDK.AddBootstrap(token, config)
	require.Nil(t, err, fmt.Sprintf("Saving config expected to succeed: %s.\n", err))

	updatedConfig := config
	updatedConfig.MFKey = thingID
	ch := channel
	ch.ID = "2"
	updatedConfig.MFChannels = append(updatedConfig.MFChannels, ch)

	cases := []struct {
		desc       string
		auth       string
		id         string
		clientCert string
		clientKey  string
		caCert     string
		config     mfsdk.BootstrapConfig
		err        error
	}{
		{
			desc:       "update cert for a non-existing config",
			id:         wrongID,
			clientCert: clientCert,
			clientKey:  clientKey,
			caCert:     caCert,
			auth:       token,
			err:        createError(sdk.ErrFailedCertUpdate, http.StatusNotFound),
		},
		{
			desc:       "update cert with invalid token",
			id:         updatedConfig.MFKey,
			clientCert: clientCert,
			clientKey:  clientKey,
			caCert:     caCert,
			auth:       invalidToken,
			err:        createError(sdk.ErrFailedCertUpdate, http.StatusUnauthorized),
		},
		{
			desc:       "update certs with an empty token",
			id:         updatedConfig.MFKey,
			clientCert: clientCert,
			clientKey:  clientKey,
			caCert:     caCert,
			auth:       "",
			err:        createError(sdk.ErrFailedCertUpdate, http.StatusUnauthorized),
		},
		{
			desc:       "update certs for the valid config",
			id:         updatedConfig.MFKey,
			clientCert: clientCert,
			clientKey:  clientKey,
			caCert:     caCert,
			auth:       token,
			err:        nil,
		},
	}
	for _, tc := range cases {
		err := mainfluxSDK.UpdateBootstrapCerts(tc.auth, tc.id, tc.clientCert, tc.clientKey, tc.caCert)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
	}
}

func TestRemoveBootstrap(t *testing.T) {
	auth := mocks.NewAuthClient(map[string]string{token: email})

	ts := newThingsServer(newBootstrapThingsService(auth))
	svc := newBootstrapService(auth, ts.URL)
	bs := newBootstrapServer(svc)
	defer bs.Close()

	sdkConf := mfsdk.Config{
		BootstrapURL:    bs.URL,
		MsgContentType:  contentType,
		TLSVerification: true,
	}
	mainfluxSDK := mfsdk.NewSDK(sdkConf)

	mfThingID, err := mainfluxSDK.AddBootstrap(token, config)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	cases := []struct {
		desc  string
		id    string
		token string
		err   error
	}{
		{
			desc:  "remove with invalid token",
			id:    mfThingID,
			token: invalidToken,
			err:   createError(sdk.ErrFailedRemoval, http.StatusUnauthorized),
		},
		{
			desc:  "remove with empty token",
			id:    mfThingID,
			token: "",
			err:   createError(sdk.ErrFailedRemoval, http.StatusUnauthorized),
		},
		{
			desc:  "remove non-existing config",
			id:    unknown,
			token: token,
			err:   nil,
		},
		{
			desc:  "remove an existing config",
			id:    mfThingID,
			token: token,
			err:   nil,
		},
		{
			desc:  "remove removed config",
			id:    mfThingID,
			token: token,
			err:   nil,
		},
	}
	for _, tc := range cases {
		err := mainfluxSDK.RemoveBootstrap(tc.token, tc.id)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
	}
}

func testBootstrap(t *testing.T) {
	auth := mocks.NewAuthClient(map[string]string{token: email})

	ts := newThingsServer(newBootstrapThingsService(auth))
	svc := newBootstrapService(auth, ts.URL)
	bs := newBootstrapServer(svc)
	defer bs.Close()

	sdkConf := mfsdk.Config{
		BootstrapURL:    bs.URL,
		MsgContentType:  contentType,
		TLSVerification: true,
	}
	mainfluxSDK := mfsdk.NewSDK(sdkConf)

	mfThingID, err := mainfluxSDK.AddBootstrap(token, config)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	updtConfig := config
	updtConfig.MFThing = mfThingID

	cases := []struct {
		desc        string
		config      mfsdk.BootstrapConfig
		externalKey string
		externalID  string
		err         error
	}{
		{
			desc:        "bootstrap an existing config",
			config:      updtConfig,
			externalID:  config.ExternalID,
			externalKey: config.ExternalKey,
			err:         nil,
		},
	}
	for _, tc := range cases {
		fmt.Println(tc.config.ThingID)
		_, err := mainfluxSDK.Bootstrap(tc.externalKey, tc.externalID)
		//assert.Equal(t, tc.config, config, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.config, config))
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))

	}

}
