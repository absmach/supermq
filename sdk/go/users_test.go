package sdk_test

import (
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/mainflux/mainflux/sdk/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	email    = "user@example.com"
	password = "password"
	token    = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MjMzODg0NzcsImlhdCI6MTUyMzM1MjQ3NywiaXNzIjoibWFpbmZsdXgiLCJzdWIiOiJqb2huLmRvZUBlbWFpbC5jb20ifQ.cygz9zoqD7Rd8f88hpQNilTCAS1DrLLgLg4PRcH-iAI"
)

func setSDKServer(tsURL string) error {
	u, err := url.Parse(tsURL)
	if err != nil {
		return err
	}

	proto := u.Scheme
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		return err
	}

	sdk.SetServerAddr(proto, host, port)

	if proto == "https" {
		sdk.SetCerts()
	}

	return nil
}
func TestCreateUser(t *testing.T) {
	// Create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() == "/users" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	err := setSDKServer(ts.URL)
	assert.Nil(t, err)

	resp, err := sdk.CreateUser(email, password)
	assert.Nil(t, err)

	assert.Equal(t, resp.StatusCode, http.StatusOK, "HTTP status code should be 200")
}

func TestCreateToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() == "/tokens" {
			w.Header().Set("Location", token)
			w.WriteHeader(http.StatusCreated)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	err := setSDKServer(ts.URL)
	assert.Nil(t, err)

	resp, err := sdk.CreateToken(email, password)
	assert.Nil(t, err)

	assert.Equal(t, resp.StatusCode, http.StatusCreated, "HTTP status code should be 201")

	val, ok := resp.Header["Location"]
	require.True(t, ok)
	assert.Equal(t, val[0], token, "Wrong token")
}
