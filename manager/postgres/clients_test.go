package postgres_test

import (
	"fmt"
	"testing"

	"github.com/mainflux/mainflux/manager"
	"github.com/mainflux/mainflux/manager/postgres"
	"github.com/stretchr/testify/assert"
)

func TestClientSave(t *testing.T) {
	email := "client-save@example.com"

	userRepo := postgres.NewUserRepository(db)
	userRepo.Save(manager.User{email, "pass"})

	clientRepo := postgres.NewClientRepository(db)

	c1 := manager.Client{
		ID:    clientRepo.Id(),
		Owner: email,
	}

	c2 := manager.Client{
		ID:    clientRepo.Id(),
		Owner: "unknown@example.com",
	}

	cases := map[string]struct {
		client manager.Client
		hasErr bool
	}{
		"new client, existing user":     {c1, false},
		"new client, non-existing user": {c2, true},
		"duplicate client":              {c1, true},
	}

	for desc, tc := range cases {
		hasErr := clientRepo.Save(tc.client) != nil
		assert.Equal(t, tc.hasErr, hasErr, fmt.Sprintf("%s: expected %t got %t\n", desc, tc.hasErr, hasErr))
	}
}

func TestClientUpdate(t *testing.T) {
	email := "client-update@example.com"

	userRepo := postgres.NewUserRepository(db)
	userRepo.Save(manager.User{email, "pass"})

	clientRepo := postgres.NewClientRepository(db)

	c := manager.Client{
		ID:    clientRepo.Id(),
		Owner: email,
	}

	clientRepo.Save(c)

	cases := map[string]struct {
		client manager.Client
		err    error
	}{
		"existing client":                            {c, nil},
		"non-existing client with existing user":     {manager.Client{ID: wrong, Owner: email}, manager.ErrNotFound},
		"non-existing client with non-existing user": {manager.Client{ID: wrong, Owner: wrong}, manager.ErrNotFound},
	}

	for desc, tc := range cases {
		err := clientRepo.Update(tc.client)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
	}
}

func TestSingleClientRetrieval(t *testing.T) {
	email := "client-single-retrieval@example.com"

	userRepo := postgres.NewUserRepository(db)
	userRepo.Save(manager.User{email, "pass"})

	clientRepo := postgres.NewClientRepository(db)

	c := manager.Client{
		ID:    clientRepo.Id(),
		Owner: email,
	}

	clientRepo.Save(c)

	cases := map[string]struct {
		owner string
		ID    string
		err   error
	}{
		"existing user":                  {c.Owner, c.ID, nil},
		"non-existing user, wrong owner": {wrong, c.ID, manager.ErrNotFound},
		"non-existing user, wrong ID":    {c.Owner, wrong, manager.ErrNotFound},
	}

	for desc, tc := range cases {
		_, err := clientRepo.One(tc.owner, tc.ID)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
	}
}

func TestMultiClientRetrieval(t *testing.T) {
	email := "client-multi-retrieval@example.com"

	userRepo := postgres.NewUserRepository(db)
	userRepo.Save(manager.User{email, "pass"})

	clientRepo := postgres.NewClientRepository(db)

	n := 10

	for i := 0; i < n; i++ {
		c := manager.Client{
			ID:    clientRepo.Id(),
			Owner: email,
		}

		clientRepo.Save(c)
	}

	cases := map[string]struct {
		owner string
		len   int
	}{
		"existing user":                  {email, n},
		"non-existing user, wrong owner": {wrong, 0},
	}

	for desc, tc := range cases {
		n := len(clientRepo.All(tc.owner))
		assert.Equal(t, tc.len, n, fmt.Sprintf("%s: expected %d got %d\n", desc, tc.len, n))
	}
}
