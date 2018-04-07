package api_test

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mainflux/mainflux/manager"
	"github.com/mainflux/mainflux/manager/api"
	"github.com/mainflux/mainflux/manager/mocks"
	"github.com/stretchr/testify/assert"
)

const (
	contentType  = "application/json; charset=utf-8"
	invalidEmail = "userexample.com"
	wrongID      = "123e4567-e89b-12d3-a456-000000000042"
)

var (
	user    = manager.User{"user@example.com", "password"}
	client  = manager.Client{Type: "app", Name: "test_app", Payload: "test_payload"}
	channel = manager.Channel{Name: "test"}
)

type testRequest struct {
	method      string
	url         string
	contentType string
	token       string
	body        io.Reader
}

func newService() manager.Service {
	users := mocks.NewUserRepository()
	clients := mocks.NewClientRepository()
	channels := mocks.NewChannelRepository(clients)
	hasher := mocks.NewHasher()
	idp := mocks.NewIdentityProvider()

	return manager.New(users, clients, channels, hasher, idp)
}

func newServer(svc manager.Service) *httptest.Server {
	mux := api.MakeHandler(svc)
	return httptest.NewServer(mux)
}

func makeRequest(client *http.Client, testReq testRequest) (*http.Response, error) {
	req, err := http.NewRequest(testReq.method, testReq.url, testReq.body)
	if err != nil {
		return nil, err
	}
	if testReq.token != "" {
		req.Header.Set("Authorization", testReq.token)
	}
	if testReq.contentType != "" {
		req.Header.Set("Content-Type", testReq.contentType)
	}
	return client.Do(req)
}

func TestRegister(t *testing.T) {
	svc := newService()
	ts := newServer(svc)
	defer ts.Close()
	client := ts.Client()
	url := ts.URL

	body, _ := json.Marshal(user)
	data := string(body)
	invalidBody, _ := json.Marshal(manager.User{Email: invalidEmail, Password: "password"})
	invalidData := string(invalidBody)

	cases := []struct {
		desc   string
		req    string
		status int
	}{
		{"register new user", data, http.StatusCreated},
		{"register existing user", data, http.StatusConflict},
		{"register user with invalid email address", invalidData, http.StatusBadRequest},
		{"register user with invalid request format", "{", http.StatusBadRequest},
		{"register user with empty JSON request", "{}", http.StatusBadRequest},
		{"register user with empty request", "", http.StatusBadRequest},
	}

	for _, tc := range cases {
		res, err := makeRequest(client, testRequest{
			method:      http.MethodPost,
			url:         fmt.Sprintf("%s/users", url),
			contentType: contentType,
			body:        strings.NewReader(tc.req),
		})
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}

func TestLogin(t *testing.T) {
	svc := newService()
	ts := newServer(svc)
	defer ts.Close()
	client := ts.Client()
	url := ts.URL

	token, _ := json.Marshal(map[string]string{"token": user.Email})
	tokenData := string(token)
	credentials, _ := json.Marshal(user)
	data := string(credentials)
	invalidEmailCredentials, _ := json.Marshal(manager.User{Email: invalidEmail, Password: "password"})
	invalidEmailData := string(invalidEmailCredentials)
	invalidCredentials, _ := json.Marshal(manager.User{"user@example.com", "invalid_password"})
	invalidData := string(invalidCredentials)
	nonexistentCredentials, _ := json.Marshal(manager.User{"non-existentuser@example.com", "pass"})
	nonexistentData := string(nonexistentCredentials)
	svc.Register(user)

	cases := []struct {
		desc   string
		req    string
		status int
		res    string
	}{
		{"login with valid credentials", data, http.StatusCreated, tokenData},
		{"login with invalid credentials", invalidData, http.StatusForbidden, ""},
		{"login with invalid email address", invalidEmailData, http.StatusBadRequest, ""},
		{"login non-existent user", nonexistentData, http.StatusForbidden, ""},
		{"login with invalid request format", "{", http.StatusBadRequest, ""},
		{"login with empty JSON request", "{}", http.StatusBadRequest, ""},
		{"login with empty request", "", http.StatusBadRequest, ""},
	}

	for _, tc := range cases {
		res, err := makeRequest(client, testRequest{
			method:      http.MethodPost,
			url:         fmt.Sprintf("%s/tokens", url),
			contentType: contentType,
			body:        strings.NewReader(tc.req),
		})
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		body, err := ioutil.ReadAll(res.Body)
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		token := strings.Trim(string(body), "\n")

		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
		assert.Equal(t, tc.res, token, fmt.Sprintf("%s: expected body %s got %s", tc.desc, tc.res, token))
	}
}

func TestAddClient(t *testing.T) {
	svc := newService()
	ts := newServer(svc)
	defer ts.Close()
	cli := ts.Client()
	url := ts.URL

	client, _ := json.Marshal(client)
	data := string(client)
	invalidClient, _ := json.Marshal(manager.Client{
		Type:    "foo",
		Name:    "invalid_client",
		Payload: "some_payload",
	})
	invalidData := string(invalidClient)
	svc.Register(user)

	cases := []struct {
		desc   string
		req    string
		auth   string
		status int
	}{
		{"add valid client", data, user.Email, http.StatusCreated},
		{"add client with invalid data", invalidData, user.Email, http.StatusBadRequest},
		{"add client with invalid auth token", data, "invalid_token", http.StatusForbidden},
		{"add client with invalid request format", "}", user.Email, http.StatusBadRequest},
		{"add client with empty JSON request", "{}", user.Email, http.StatusBadRequest},
		{"add client with empty request", "", user.Email, http.StatusBadRequest},
	}

	for _, tc := range cases {
		res, err := makeRequest(cli, testRequest{
			method:      http.MethodPost,
			url:         fmt.Sprintf("%s/clients", url),
			contentType: contentType,
			token:       tc.auth,
			body:        strings.NewReader(tc.req),
		})
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}

func TestUpdateClient(t *testing.T) {
	svc := newService()
	ts := newServer(svc)
	defer ts.Close()
	cli := ts.Client()
	url := ts.URL

	update, _ := json.Marshal(client)
	data := string(update)
	invalidUpdate, _ := json.Marshal(manager.Client{
		Type:    "foo",
		Name:    client.Name,
		Payload: client.Payload,
	})
	invalidData := string(invalidUpdate)
	svc.Register(user)
	id, _ := svc.AddClient(user.Email, client)

	cases := []struct {
		desc   string
		req    string
		id     string
		auth   string
		status int
	}{
		{"update existing client", data, id, user.Email, http.StatusOK},
		{"update non-existent client", data, wrongID, user.Email, http.StatusNotFound},
		{"update client with invalid id", data, "1", user.Email, http.StatusNotFound},
		{"update client with invalid data", invalidData, id, user.Email, http.StatusBadRequest},
		{"update client with invalid user token", data, id, invalidEmail, http.StatusForbidden},
		{"update client with invalid data format", "{", id, user.Email, http.StatusBadRequest},
		{"update client with empty JSON request", "{}", id, user.Email, http.StatusBadRequest},
		{"update client with empty request", "", id, user.Email, http.StatusBadRequest},
	}

	for _, tc := range cases {
		res, err := makeRequest(cli, testRequest{
			method:      http.MethodPut,
			url:         fmt.Sprintf("%s/clients/%s", url, tc.id),
			contentType: contentType,
			token:       tc.auth,
			body:        strings.NewReader(tc.req),
		})
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}

func TestViewClient(t *testing.T) {
	svc := newService()
	ts := newServer(svc)
	defer ts.Close()
	cli := ts.Client()
	url := ts.URL

	svc.Register(user)
	id, _ := svc.AddClient(user.Email, client)

	client.ID = id
	client.Key = id
	client, _ := json.Marshal(client)
	data := string(client)

	cases := []struct {
		desc   string
		id     string
		auth   string
		status int
		res    string
	}{
		{"view existing client", id, user.Email, http.StatusOK, data},
		{"view non-existent client", wrongID, user.Email, http.StatusNotFound, ""},
		{"view client by passing invalid id", "1", user.Email, http.StatusNotFound, ""},
		{"view client by passing invalid token", id, invalidEmail, http.StatusForbidden, ""},
	}

	for _, tc := range cases {
		res, err := makeRequest(cli, testRequest{
			method: http.MethodGet,
			url:    fmt.Sprintf("%s/clients/%s", url, tc.id),
			token:  tc.auth,
		})
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		body, err := ioutil.ReadAll(res.Body)
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		data := strings.Trim(string(body), "\n")
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
		assert.Equal(t, tc.res, data, fmt.Sprintf("%s: expected body %s got %s", tc.desc, tc.res, data))
	}
}

func TestListClients(t *testing.T) {
	svc := newService()
	ts := newServer(svc)
	defer ts.Close()
	cli := ts.Client()
	url := ts.URL

	svc.Register(user)
	noClientsUser := manager.User{Email: "no_clients_user@example.com", Password: user.Password}
	svc.Register(noClientsUser)
	clients := []manager.Client{}
	for i := 0; i < 10; i++ {
		id, _ := svc.AddClient(user.Email, client)
		client.ID = id
		client.Key = id
		clients = append(clients, client)
	}

	cases := []struct {
		desc   string
		auth   string
		status int
		res    []manager.Client
	}{
		{"fetch list of clients", user.Email, http.StatusOK, clients},
		{"fetch empty list of clients", noClientsUser.Email, http.StatusOK, []manager.Client{}},
		{"fetch list of clients with invalid token", invalidEmail, http.StatusForbidden, nil},
	}

	for _, tc := range cases {
		res, err := makeRequest(cli, testRequest{
			method: http.MethodGet,
			url:    fmt.Sprintf("%s/clients", url),
			token:  tc.auth,
		})
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		var data map[string][]manager.Client
		json.NewDecoder(res.Body).Decode(&data)
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
		assert.ElementsMatch(t, tc.res, data["clients"], fmt.Sprintf("%s: expected body %s got %s", tc.desc, tc.res, data["clients"]))
	}
}

func TestRemoveClient(t *testing.T) {
	svc := newService()
	ts := newServer(svc)
	defer ts.Close()
	cli := ts.Client()
	url := ts.URL

	svc.Register(user)
	id, _ := svc.AddClient(user.Email, client)

	cases := []struct {
		desc   string
		id     string
		auth   string
		status int
	}{
		{"delete existing client", id, user.Email, http.StatusNoContent},
		{"delete non-existent client", wrongID, user.Email, http.StatusNoContent},
		{"delete client with invalid id", "1", user.Email, http.StatusNoContent},
		{"delete client with invalid token", id, invalidEmail, http.StatusForbidden},
	}

	for _, tc := range cases {
		res, err := makeRequest(cli, testRequest{
			method: http.MethodDelete,
			url:    fmt.Sprintf("%s/clients/%s", url, tc.id),
			token:  tc.auth,
		})
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}

func TestCreateChannel(t *testing.T) {
	svc := newService()
	ts := newServer(svc)
	defer ts.Close()
	client := ts.Client()
	url := ts.URL

	channel, _ := json.Marshal(channel)
	data := string(channel)
	svc.Register(user)

	cases := []struct {
		desc   string
		req    string
		auth   string
		status int
	}{
		{"create new channel", data, user.Email, http.StatusCreated},
		{"create new channel with invalid token", data, invalidEmail, http.StatusForbidden},
		{"create new channel with invalid data format", "{", user.Email, http.StatusBadRequest},
		{"create new channel with empty JSON request", "{}", user.Email, http.StatusCreated},
		{"create new channel with empty request", "", user.Email, http.StatusBadRequest},
	}

	for _, tc := range cases {
		res, err := makeRequest(client, testRequest{
			method:      http.MethodPost,
			url:         fmt.Sprintf("%s/channels", url),
			contentType: contentType,
			token:       tc.auth,
			body:        strings.NewReader(tc.req),
		})
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}

func TestUpdateChannel(t *testing.T) {
	svc := newService()
	ts := newServer(svc)
	defer ts.Close()
	client := ts.Client()
	url := ts.URL

	update, _ := json.Marshal(map[string]string{
		"name": "updated_channel",
	})
	updateData := string(update)
	svc.Register(user)
	id, _ := svc.CreateChannel(user.Email, channel)

	cases := []struct {
		desc   string
		req    string
		id     string
		auth   string
		status int
	}{
		{"update existing channel", updateData, id, user.Email, http.StatusOK},
		{"update non-existing channel", updateData, wrongID, user.Email, http.StatusNotFound},
		{"update channel with invalid token", updateData, id, invalidEmail, http.StatusForbidden},
		{"update channel with invalid id", updateData, "1", user.Email, http.StatusNotFound},
		{"update channel with invalid data format", "}", id, user.Email, http.StatusBadRequest},
		{"update channel with empty JSON object", "{}", id, user.Email, http.StatusOK},
		{"update channel with empty request", "", id, user.Email, http.StatusBadRequest},
	}

	for _, tc := range cases {
		res, err := makeRequest(client, testRequest{
			method:      http.MethodPut,
			url:         fmt.Sprintf("%s/channels/%s", url, tc.id),
			contentType: contentType,
			token:       tc.auth,
			body:        strings.NewReader(tc.req),
		})
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}

func TestViewChannel(t *testing.T) {
	svc := newService()
	ts := newServer(svc)
	defer ts.Close()
	client := ts.Client()
	url := ts.URL

	svc.Register(user)
	id, _ := svc.CreateChannel(user.Email, channel)
	channel.ID = id
	channel, _ := json.Marshal(channel)
	data := string(channel)

	cases := []struct {
		desc   string
		id     string
		auth   string
		status int
		res    string
	}{
		{"view existing channel", id, user.Email, http.StatusOK, data},
		{"view non-existent channel", wrongID, user.Email, http.StatusNotFound, ""},
		{"view channel with invalid id", "1", user.Email, http.StatusNotFound, ""},
		{"view channel with invalid token", id, invalidEmail, http.StatusForbidden, ""},
	}

	for _, tc := range cases {
		res, err := makeRequest(client, testRequest{
			method: http.MethodGet,
			url:    fmt.Sprintf("%s/channels/%s", url, tc.id),
			token:  tc.auth,
		})
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		data, err := ioutil.ReadAll(res.Body)
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		body := strings.Trim(string(data), "\n")
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
		assert.Equal(t, tc.res, body, fmt.Sprintf("%s: expected body %s got %s", tc.desc, tc.res, body))
	}
}

func TestListChannels(t *testing.T) {
	svc := newService()
	ts := newServer(svc)
	defer ts.Close()
	client := ts.Client()
	url := ts.URL

	svc.Register(user)
	channels := []manager.Channel{}
	for i := 0; i < 10; i++ {
		id, _ := svc.CreateChannel(user.Email, channel)
		channel.ID = id
		channels = append(channels, channel)
	}

	cases := []struct {
		desc   string
		auth   string
		status int
		res    []manager.Channel
	}{
		{"get a list of channels", user.Email, http.StatusOK, channels},
		{"get a list of channels with invalid token", invalidEmail, http.StatusForbidden, nil},
	}

	for _, tc := range cases {
		res, err := makeRequest(client, testRequest{
			method: http.MethodGet,
			url:    fmt.Sprintf("%s/channels", url),
			token:  tc.auth,
		})
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		var body map[string][]manager.Channel
		json.NewDecoder(res.Body).Decode(&body)
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
		assert.ElementsMatch(t, tc.res, body["channels"], fmt.Sprintf("%s: expected body %s got %s", tc.desc, tc.res, body["channels"]))
	}
}

func TestRemoveChannel(t *testing.T) {
	svc := newService()
	ts := newServer(svc)
	defer ts.Close()
	client := ts.Client()
	url := ts.URL

	svc.Register(user)
	id, _ := svc.CreateChannel(user.Email, channel)
	channel.ID = id

	cases := []struct {
		desc   string
		id     string
		auth   string
		status int
	}{
		{"remove existing channel", channel.ID, user.Email, http.StatusNoContent},
		{"remove non-existent channel", channel.ID, user.Email, http.StatusNoContent},
		{"remove channel with invalid id", wrongID, user.Email, http.StatusNoContent},
		{"remove channel with invalid token", channel.ID, invalidEmail, http.StatusForbidden},
	}

	for _, tc := range cases {
		res, err := makeRequest(client, testRequest{
			method: http.MethodDelete,
			url:    fmt.Sprintf("%s/channels/%s", url, tc.id),
			token:  tc.auth,
		})
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}

func TestConnect(t *testing.T) {
	svc := newService()
	ts := newServer(svc)
	defer ts.Close()
	cli := ts.Client()
	url := ts.URL

	svc.Register(user)
	clientID, _ := svc.AddClient(user.Email, client)
	chanID, _ := svc.CreateChannel(user.Email, channel)

	otherUser := manager.User{Email: "other_user@example.com", Password: "password"}
	svc.Register(otherUser)
	otherClientID, _ := svc.AddClient(otherUser.Email, client)
	otherChanID, _ := svc.CreateChannel(otherUser.Email, channel)

	cases := []struct {
		desc     string
		chanID   string
		clientID string
		auth     string
		status   int
	}{
		{"connect existing client to existing channel", chanID, clientID, user.Email, http.StatusOK},
		{"connect existing client to non-existent channel", wrongID, clientID, user.Email, http.StatusNotFound},
		{"connect client with invalid id to channel", chanID, "1", user.Email, http.StatusNotFound},
		{"connect client to channel with invalid id", "1", clientID, user.Email, http.StatusNotFound},
		{"connect existing client to existing channel with invalid token", chanID, clientID, invalidEmail, http.StatusForbidden},
		{"connect client from owner to channel of other user", otherChanID, clientID, user.Email, http.StatusNotFound},
		{"connect client from other user to owner's channel", chanID, otherClientID, user.Email, http.StatusNotFound},
	}

	for _, tc := range cases {
		res, err := makeRequest(cli, testRequest{
			method: http.MethodPut,
			url:    fmt.Sprintf("%s/channels/%s/clients/%s", url, tc.chanID, tc.clientID),
			token:  tc.auth,
		})
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}

func TestDisconnnect(t *testing.T) {
	svc := newService()
	ts := newServer(svc)
	defer ts.Close()
	cli := ts.Client()
	url := ts.URL

	svc.Register(user)
	clientID, _ := svc.AddClient(user.Email, client)
	chanID, _ := svc.CreateChannel(user.Email, channel)
	svc.Connect(user.Email, chanID, clientID)
	otherUser := manager.User{Email: "other_user@example.com", Password: "password"}
	svc.Register(otherUser)
	otherClientID, _ := svc.AddClient(otherUser.Email, client)
	otherChanID, _ := svc.CreateChannel(otherUser.Email, channel)
	svc.Connect(otherUser.Email, otherChanID, otherClientID)

	cases := []struct {
		desc     string
		chanID   string
		clientID string
		auth     string
		status   int
	}{
		{"disconnect connected client from channel", chanID, clientID, user.Email, http.StatusNoContent},
		{"disconnect non-connected client from channel", chanID, clientID, user.Email, http.StatusNotFound},
		{"disconnect non-existent client from channel", chanID, "1", user.Email, http.StatusNotFound},
		{"disconnect client from non-existent channel", "1", clientID, user.Email, http.StatusNotFound},
		{"disconnect client from channel with invalid token", chanID, clientID, invalidEmail, http.StatusForbidden},
		{"disconnect owner's client from someone elses channel", otherChanID, clientID, user.Email, http.StatusNotFound},
		{"disconnect other's client from owner's channel", chanID, otherClientID, user.Email, http.StatusNotFound},
	}

	for _, tc := range cases {
		res, err := makeRequest(cli, testRequest{
			method: http.MethodDelete,
			url:    fmt.Sprintf("%s/channels/%s/clients/%s", url, tc.chanID, tc.clientID),
			token:  tc.auth,
		})
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}

func TestIdentity(t *testing.T) {
	svc := newService()
	ts := newServer(svc)
	defer ts.Close()
	cli := ts.Client()
	url := ts.URL

	svc.Register(user)
	clientID, _ := svc.AddClient(user.Email, client)

	cases := []struct {
		desc     string
		key      string
		status   int
		clientID string
	}{
		{"get client id using existing client key", clientID, http.StatusOK, clientID},
		{"get client id using non-existent client key", "", http.StatusForbidden, ""},
	}

	for _, tc := range cases {
		res, err := makeRequest(cli, testRequest{
			method: http.MethodGet,
			url:    fmt.Sprintf("%s/access-grant", url),
			token:  tc.key,
		})
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		clientID := res.Header.Get("X-client-id")
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
		assert.Equal(t, tc.clientID, clientID, fmt.Sprintf("%s: expected %s got %s", tc.desc, tc.clientID, clientID))
	}
}

func TestCanAccess(t *testing.T) {
	svc := newService()
	ts := newServer(svc)
	defer ts.Close()
	cli := ts.Client()
	url := ts.URL

	svc.Register(user)
	clientID, _ := svc.AddClient(user.Email, client)
	notConnectedClientID, _ := svc.AddClient(user.Email, client)
	chanID, _ := svc.CreateChannel(user.Email, channel)
	svc.Connect(user.Email, chanID, clientID)

	cases := []struct {
		desc      string
		chanID    string
		clientKey string
		status    int
		clientID  string
	}{
		{"check access to existing channel given connected client", chanID, clientID, http.StatusOK, clientID},
		{"check access to existing channel given not connected client", chanID, notConnectedClientID, http.StatusForbidden, ""},
		{"check access to existing channel given non-existent client", chanID, "invalid_token", http.StatusForbidden, ""},
		{"check access to non-existent channel given existing client", "invalid_token", clientID, http.StatusForbidden, ""},
	}

	for _, tc := range cases {
		res, err := makeRequest(cli, testRequest{
			method: http.MethodGet,
			url:    fmt.Sprintf("%s/channels/%s/access-grant", url, tc.chanID),
			token:  tc.clientKey,
		})
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		clientID := res.Header.Get("X-client-id")
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
		assert.Equal(t, tc.clientID, clientID, fmt.Sprintf("%s: expected %s got %s", tc.desc, tc.clientID, clientID))
	}
}
