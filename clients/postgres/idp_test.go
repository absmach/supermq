package postgres_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mainflux/mainflux/clients"
	"github.com/mainflux/mainflux/clients/postgres"
)

func TestKey(t *testing.T) {
	idp := postgres.NewIdentityProvider(db, testLog)
	key := idp.Key()
	assert.NotEmpty(t, key, "generate client key: expected non-empty value got empty value")
}

func TestIdentity(t *testing.T) {
	email := "identity-client@example.com"

	idp := postgres.NewIdentityProvider(db, testLog)

	clientRepo := postgres.NewClientRepository(db, testLog)

	c := clients.Client{
		ID:    clientRepo.ID(),
		Owner: email,
		Key:   idp.Key(),
	}

	clientRepo.Save(c)

	cases := map[string]struct {
		key string
		id  string
		err error
	}{
		"identify existing user":     {c.Key, c.ID, nil},
		"identify non-existent user": {"invalid-key", "", clients.ErrUnauthorizedAccess},
	}

	for desc, tc := range cases {
		id, err := idp.Identity(tc.key)
		assert.Equal(t, tc.id, id, fmt.Sprintf("%s: expected %s got %s", desc, tc.id, id))
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s", desc, tc.err, err))
	}
}
