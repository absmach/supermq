package postgres_test

import (
	"fmt"
	"testing"

	"github.com/mainflux/mainflux/manager"
	"github.com/mainflux/mainflux/manager/postgres"
	"github.com/stretchr/testify/assert"
)

func TestUserSave(t *testing.T) {
	cases := map[string]struct {
		user manager.User
		err  error
	}{
		"new user":       {manager.User{"foo@bar.com", "pass"}, nil},
		"duplicate user": {manager.User{"foo@bar.com", "pass"}, manager.ErrConflict},
	}

	repo := postgres.NewUserRepository(db)

	for desc, tc := range cases {
		err := repo.Save(tc.user)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
	}
}

func TestSingleUserRetrieval(t *testing.T) {
	cases := map[string]struct {
		email string
		err   error
	}{
		"existing user":     {"foo@bar.com", nil},
		"non-existing user": {"dummy@example.com", manager.ErrNotFound},
	}

	repo := postgres.NewUserRepository(db)

	for desc, tc := range cases {
		_, err := repo.One(tc.email)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
	}
}
