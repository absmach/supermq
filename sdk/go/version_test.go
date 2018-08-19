package sdk_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/sdk/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	version = "0.4.0"
	service = "testService"
)

func TestVersion(t *testing.T) {
	// Create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() == "/version" {
			vi := mainflux.VersionInfo{
				Service: service,
				Version: version,
			}

			data, err := json.Marshal(vi)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Write(data)
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	err := setSDKServer(ts.URL)
	assert.Nil(t, err)

	resp, err := sdk.Version()
	require.Nil(t, err)

	require.Equal(t, resp.StatusCode, http.StatusOK, "HTTP status code should be 200")

	vi := mainflux.VersionInfo{}
	viExp := mainflux.VersionInfo{
		Service: service,
		Version: version,
	}

	body, err := ioutil.ReadAll(resp.Body)
	require.Nil(t, err)

	err = json.Unmarshal(body, &vi)
	require.Nil(t, err)
	assert.Equal(t, vi, viExp, "VersionInfo structure not as expected")

}
