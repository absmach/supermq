package bootstrap_test

import (
	"fmt"
	"net/http/httptest"
	"nov/bootstrap"
	"nov/bootstrap/mocks"
	"testing"

	"github.com/mainflux/mainflux"

	"github.com/mainflux/mainflux/things"

	httpapi "github.com/mainflux/mainflux/things/api/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mfsdk "github.com/mainflux/mainflux/sdk/go"
)

const (
	validToken   = "validToken"
	invalidToken = "invalidToken"
	email        = "test@example.com"
)

var thing = bootstrap.Config{
	ExternalID:  "external-id",
	ExternalKey: "external-key",
	MFChannels:  []string{"1"},
	Content:     "config",
}

func newService(users mainflux.UsersServiceClient, url string) bootstrap.Service {
	things := mocks.NewThingsRepository()
	config := mfsdk.Config{
		BaseURL: url,
	}

	sdk := mfsdk.NewSDK(config)
	return bootstrap.New(users, things, sdk)
}

func newThingsService(t map[string]things.Thing, users mainflux.UsersServiceClient) things.Service {
	c := map[string]things.Channel{
		"1": things.Channel{
			ID:     "1",
			Owner:  validToken,
			Things: []things.Thing{t["1"]},
		},
	}

	return mocks.NewThingsService(t, c, users)
}

func newServer(svc things.Service) *httptest.Server {
	mux := httpapi.MakeHandler(svc)
	return httptest.NewServer(mux)
}

func TestAdd(t *testing.T) {
	users := mocks.NewUsersService(map[string]string{validToken: email})

	c := map[string]things.Channel{
		"1": things.Channel{
			ID:     "1",
			Owner:  validToken,
			Things: []things.Thing{},
		},
	}
	server := newServer(mocks.NewThingsService(map[string]things.Thing{}, c, users))
	svc := newService(users, server.URL)

	wrongChannels := thing
	wrongChannels.MFChannels = append(wrongChannels.MFChannels, "2")

	cases := []struct {
		desc  string
		thing bootstrap.Config
		key   string
		err   error
	}{
		{
			desc:  "add a new thing",
			thing: thing,
			key:   validToken,
			err:   nil,
		},
		{
			desc:  "add a thing with wrong credentials",
			thing: thing,
			key:   invalidToken,
			err:   bootstrap.ErrUnauthorizedAccess,
		},
		{
			desc:  "add a thing with invalid list of channels",
			thing: wrongChannels,
			key:   validToken,
			err:   bootstrap.ErrMalformedEntity,
		},
	}

	for _, tc := range cases {
		_, err := svc.Add(tc.key, tc.thing)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestView(t *testing.T) {
	users := mocks.NewUsersService(map[string]string{validToken: email})

	c := map[string]things.Channel{
		"1": things.Channel{
			ID:     "1",
			Owner:  validToken,
			Things: []things.Thing{},
		},
	}
	server := newServer(mocks.NewThingsService(map[string]things.Thing{}, c, users))
	svc := newService(users, server.URL)

	saved, err := svc.Add(validToken, thing)
	require.Nil(t, err, fmt.Sprintf("Saving thing expected to succeed: %s.\n", err))

	cases := []struct {
		desc string
		id   string
		key  string
		err  error
	}{
		{
			desc: "view an existing thing",
			id:   saved.ID,
			key:  validToken,
			err:  nil,
		},
		{
			desc: "view a non-existing thing",
			id:   "non-existing",
			key:  validToken,
			err:  bootstrap.ErrNotFound,
		},
		{
			desc: "view a thing with wrong credentials",
			id:   thing.ID,
			key:  invalidToken,
			err:  bootstrap.ErrUnauthorizedAccess,
		},
	}

	for _, tc := range cases {
		_, err := svc.View(tc.key, tc.id)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestUpdate(t *testing.T) {
	users := mocks.NewUsersService(map[string]string{validToken: email})

	c := map[string]things.Channel{
		"1": things.Channel{
			ID:     "1",
			Owner:  validToken,
			Things: []things.Thing{},
		},
		"2": things.Channel{
			ID:     "2",
			Owner:  validToken,
			Things: []things.Thing{},
		},
	}

	server := newServer(mocks.NewThingsService(map[string]things.Thing{}, c, users))
	svc := newService(users, server.URL)

	saved, err := svc.Add(validToken, thing)
	require.Nil(t, err, fmt.Sprintf("Saving thing expected to succeed: %s.\n", err))

	saved.Content = "new-config"
	saved.MFChannels = []string{"2"}
	saved.State = bootstrap.Active

	nonExisting := thing
	nonExisting.ID = "non-existing"

	wrongChannels := saved
	wrongChannels.MFChannels = append(wrongChannels.MFChannels, "2")

	cases := []struct {
		desc  string
		thing bootstrap.Config
		key   string
		err   error
	}{
		{
			desc:  "update a thing",
			thing: saved,
			key:   validToken,
			err:   nil,
		},
		{
			desc:  "update a non-existing thing",
			thing: nonExisting,
			key:   validToken,
			err:   bootstrap.ErrNotFound,
		},
		{
			desc:  "update a thing with wrong credentials",
			thing: saved,
			key:   invalidToken,
			err:   bootstrap.ErrUnauthorizedAccess,
		},
		{
			desc:  "update a thing with invalid list of channels",
			thing: wrongChannels,
			key:   validToken,
			err:   bootstrap.ErrNotFound,
		},
	}

	for _, tc := range cases {
		err := svc.Update(tc.key, tc.thing)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestList(t *testing.T) {
	users := mocks.NewUsersService(map[string]string{validToken: email})

	c := map[string]things.Channel{
		"1": things.Channel{
			ID:     "1",
			Owner:  email,
			Things: []things.Thing{},
		},
	}
	server := newServer(mocks.NewThingsService(map[string]things.Thing{}, c, users))
	svc := newService(users, server.URL)

	numThings := 101
	var saved []bootstrap.Config
	for i := 0; i < numThings; i++ {
		s, err := svc.Add(validToken, thing)
		saved = append(saved, s)
		require.Nil(t, err, fmt.Sprintf("Saving thing expected to succeed: %s.\n", err))
	}
	// Set one Thing to the different state
	err := svc.ChangeState(validToken, "42", bootstrap.Active)
	require.Nil(t, err, fmt.Sprintf("Changing thing state expected to succeed: %s.\n", err))
	saved[41].State = bootstrap.Active

	cases := []struct {
		desc   string
		things []bootstrap.Config
		filter bootstrap.Filter
		offset uint64
		limit  uint64
		key    string
		err    error
	}{
		{
			desc:   "list things",
			things: saved[0:10],
			filter: bootstrap.Filter{},
			key:    validToken,
			offset: 0,
			limit:  10,
			err:    nil,
		},
		{
			desc:   "list things with wrong credentials",
			things: []bootstrap.Config{},
			filter: bootstrap.Filter{},
			key:    invalidToken,
			offset: 0,
			limit:  10,
			err:    bootstrap.ErrUnauthorizedAccess,
		},
		{
			desc:   "list last page",
			things: saved[95:],
			filter: bootstrap.Filter{},
			key:    validToken,
			offset: 95,
			limit:  10,
			err:    nil,
		},
		{
			desc:   "list Active things",
			things: []bootstrap.Config{saved[41]},
			filter: bootstrap.Filter{"state": bootstrap.Active.String()},
			key:    validToken,
			offset: 35,
			limit:  20,
			err:    nil,
		},
	}

	for _, tc := range cases {
		result, err := svc.List(tc.key, tc.filter, tc.offset, tc.limit)
		assert.ElementsMatch(t, tc.things, result, fmt.Sprintf("%s: expected %v got %v", tc.desc, tc.things, result))
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestRemove(t *testing.T) {
	users := mocks.NewUsersService(map[string]string{validToken: email})

	c := map[string]things.Channel{
		"1": things.Channel{
			ID:     "1",
			Owner:  validToken,
			Things: []things.Thing{},
		},
	}

	server := newServer(mocks.NewThingsService(map[string]things.Thing{}, c, users))
	svc := newService(users, server.URL)

	saved, err := svc.Add(validToken, thing)
	require.Nil(t, err, fmt.Sprintf("Saving thing expected to succeed: %s.\n", err))

	cases := []struct {
		desc string
		id   string
		key  string
		err  error
	}{
		{
			desc: "view a thing with wrong credentials",
			id:   thing.ID,
			key:  invalidToken,
			err:  bootstrap.ErrUnauthorizedAccess,
		},
		{
			desc: "remove an existing thing",
			id:   saved.ID,
			key:  validToken,
			err:  nil,
		},
		{
			desc: "remove removed thing",
			id:   saved.ID,
			key:  validToken,
			err:  nil,
		},
		{
			desc: "remove non-existing thing",
			id:   "non-existing",
			key:  validToken,
			err:  nil,
		},
	}

	for _, tc := range cases {
		err := svc.Remove(tc.key, tc.id)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestBootstrap(t *testing.T) {
	users := mocks.NewUsersService(map[string]string{validToken: email})

	c := map[string]things.Channel{
		"1": things.Channel{
			ID:     "1",
			Owner:  validToken,
			Things: []things.Thing{},
		},
	}

	server := newServer(mocks.NewThingsService(map[string]things.Thing{}, c, users))
	svc := newService(users, server.URL)

	saved, err := svc.Add(validToken, thing)
	require.Nil(t, err, fmt.Sprintf("Saving thing expected to succeed: %s.\n", err))

	cases := []struct {
		desc        string
		config      bootstrap.Config
		externalKey string
		externalID  string
		err         error
	}{
		{
			desc:        "bootstrap using invalid external id",
			config:      bootstrap.Config{},
			externalID:  "invalid",
			externalKey: thing.ExternalKey,
			err:         nil,
		},
		{
			desc:        "bootstrap using invalid external key",
			config:      bootstrap.Config{},
			externalID:  thing.ExternalID,
			externalKey: "invalid",
			err:         nil,
		},
		{
			desc:        "bootstrap an existing thing",
			config:      saved,
			externalID:  thing.ExternalID,
			externalKey: thing.ExternalKey,
			err:         nil,
		},
	}

	for _, tc := range cases {
		config, err := svc.Bootstrap(tc.externalKey, tc.externalID)
		assert.Equal(t, tc.config, config, fmt.Sprintf("%s: expected %v got %v\n", tc.desc, tc.config, config))
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}
func TestChangeState(t *testing.T) {
	users := mocks.NewUsersService(map[string]string{validToken: email})

	c := map[string]things.Channel{
		"1": things.Channel{
			ID:     "1",
			Owner:  email,
			Things: []things.Thing{},
		},
		"2": things.Channel{
			ID:     "2",
			Owner:  email,
			Things: []things.Thing{},
		},
	}

	server := newServer(mocks.NewThingsService(map[string]things.Thing{}, c, users))
	svc := newService(users, server.URL)

	saved, err := svc.Add(validToken, thing)
	require.Nil(t, err, fmt.Sprintf("Saving thing expected to succeed: %s.\n", err))

	cases := []struct {
		desc  string
		state bootstrap.State
		id    string
		key   string
		err   error
	}{
		{
			desc:  "change state with wrong credentials",
			state: bootstrap.Active,
			id:    saved.ID,
			key:   invalidToken,
			err:   bootstrap.ErrUnauthorizedAccess,
		},
		{
			desc:  "change state to Active",
			state: bootstrap.Active,
			id:    saved.ID,
			key:   validToken,
			err:   nil,
		},
		{
			desc:  "change state to Inactive",
			state: bootstrap.Inactive,
			id:    saved.ID,
			key:   validToken,
			err:   nil,
		},
	}

	for _, tc := range cases {
		err := svc.ChangeState(tc.key, tc.id, tc.state)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}
