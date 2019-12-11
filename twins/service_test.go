//
// Copyright (c) 2019
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package twins_test

import (
	"context"
	"fmt"
	"testing"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/mainflux/mainflux/twins"
	"github.com/mainflux/mainflux/twins/mocks"
	broker "github.com/nats-io/go-nats"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mainflux/mainflux/twins/paho"
)

const (
	twinName   = "name"
	wrongKey   = ""
	wrongID    = ""
	token      = "token"
	wrongToken = "wrong-token"
	email      = "user@example.com"
	natsURL    = "nats://localhost:4222"
	mqttURL    = "tcp://localhost:1883"
	topic      = "topic"
)

func newTwin(name string) twins.Twin {
	return twins.Twin{
		Name:        name,
		Definitions: []twins.Definition{twins.Definition{}},
	}
}

func newService(tokens map[string]string) twins.Service {
	users := mocks.NewUsersService(tokens)
	twinsRepo := mocks.NewTwinRepository()
	statesRepo := mocks.NewStateRepository()
	idp := mocks.NewIdentityProvider()

	nc, _ := broker.Connect(natsURL)

	opts := mqtt.NewClientOptions()
	pc := mqtt.NewClient(opts)

	mc := paho.New(pc, topic)

	return twins.New(nc, mc, users, twinsRepo, statesRepo, idp)
}

func TestAddTwin(t *testing.T) {
	svc := newService(map[string]string{token: email})

	cases := []struct {
		desc  string
		twin  twins.Twin
		token string
		err   error
	}{
		{
			desc:  "add new twin",
			twin:  newTwin(twinName),
			token: token,
			err:   nil,
		},
		{
			desc:  "add twin with wrong credentials",
			twin:  newTwin(twinName),
			token: wrongToken,
			err:   twins.ErrUnauthorizedAccess,
		},
	}

	for _, tc := range cases {
		_, err := svc.AddTwin(context.Background(), tc.token, tc.twin, tc.twin.Definitions[0])
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestUpdateTwin(t *testing.T) {
	svc := newService(map[string]string{token: email})
	twin := newTwin(twinName)
	other := newTwin(twinName)
	other.ID = wrongID
	saved, err := svc.AddTwin(context.Background(), token, twin, twin.Definitions[0])
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s\n", err))

	cases := []struct {
		desc  string
		twin  twins.Twin
		token string
		err   error
	}{
		{
			desc:  "update existing twin",
			twin:  saved,
			token: token,
			err:   nil,
		},
		{
			desc:  "update twin with wrong credentials",
			twin:  saved,
			token: wrongToken,
			err:   twins.ErrUnauthorizedAccess,
		},
		{
			desc:  "update non-existing twin",
			twin:  other,
			token: token,
			err:   twins.ErrNotFound,
		},
	}

	for _, tc := range cases {
		err := svc.UpdateTwin(context.Background(), tc.token, tc.twin, tc.twin.Definitions[0])
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestViewTwin(t *testing.T) {
	svc := newService(map[string]string{token: email})
	twin := newTwin(twinName)
	saved, err := svc.AddTwin(context.Background(), token, twin, twin.Definitions[0])
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s\n", err))

	cases := map[string]struct {
		id    string
		token string
		err   error
	}{
		"view existing twin": {
			id:    saved.ID,
			token: token,
			err:   nil,
		},
		"view twin with wrong credentials": {
			id:    saved.ID,
			token: wrongToken,
			err:   twins.ErrUnauthorizedAccess,
		},
		"view non-existing twin": {
			id:    wrongID,
			token: token,
			err:   twins.ErrNotFound,
		},
	}

	for desc, tc := range cases {
		_, err := svc.ViewTwin(context.Background(), tc.token, tc.id)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
	}
}

func TestListTwins(t *testing.T) {
	svc := newService(map[string]string{token: email})

	twin := newTwin(twinName)
	m := make(map[string]interface{})
	m["serial"] = "123456"
	twin.Metadata = m

	n := uint64(10)
	for i := uint64(0); i < n; i++ {
		svc.AddTwin(context.Background(), token, twin, twin.Definitions[0])
	}

	cases := map[string]struct {
		token    string
		offset   uint64
		limit    uint64
		name     string
		size     uint64
		metadata map[string]interface{}
		err      error
	}{
		"list all twins": {
			token: token,
			limit: n + 1,
			size:  n,
			err:   nil,
		},
		"list with zero limit": {
			token: token,
			limit: 0,
			size:  0,
			err:   nil,
		},
		"list with wrong credentials": {
			token: wrongToken,
			limit: 0,
			size:  0,
			err:   twins.ErrUnauthorizedAccess,
		},
		"list with metadata": {
			token:    token,
			limit:    n + 1,
			size:     n,
			err:      nil,
			metadata: m,
		},
	}

	for desc, tc := range cases {
		page, err := svc.ListTwins(context.Background(), tc.token, tc.limit, 10, tc.name, tc.metadata)
		size := uint64(len(page.Twins))
		assert.Equal(t, tc.size, size, fmt.Sprintf("%s: expected %d got %d\n", desc, tc.size, size))
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
	}
}

func TestRemoveTwin(t *testing.T) {
	svc := newService(map[string]string{token: email})
	twin := newTwin(twinName)
	saved, err := svc.AddTwin(context.Background(), token, twin, twin.Definitions[0])
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s\n", err))

	cases := []struct {
		desc  string
		id    string
		token string
		err   error
	}{
		{
			desc:  "remove twin with wrong credentials",
			id:    saved.ID,
			token: wrongToken,
			err:   twins.ErrUnauthorizedAccess,
		},
		{
			desc:  "remove existing twin",
			id:    saved.ID,
			token: token,
			err:   nil,
		},
		{
			desc:  "remove removed twin",
			id:    saved.ID,
			token: token,
			err:   nil,
		},
		{
			desc:  "remove non-existing twin",
			id:    wrongID,
			token: token,
			err:   nil,
		},
	}

	for _, tc := range cases {
		err := svc.RemoveTwin(context.Background(), tc.token, tc.id)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}
