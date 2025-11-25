// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"

	"github.com/absmach/supermq/notifications"
	"github.com/absmach/supermq/pkg/errors"
	repoerr "github.com/absmach/supermq/pkg/errors/repository"
	"github.com/absmach/supermq/pkg/postgres"
)

var _ notifications.UserRepository = (*userRepository)(nil)

type userRepository struct {
	db postgres.Database
}

// NewUserRepository instantiates a PostgreSQL implementation of user repository.
func NewUserRepository(db postgres.Database) notifications.UserRepository {
	return &userRepository{
		db: db,
	}
}

func (repo *userRepository) RetrieveByID(ctx context.Context, id string) (notifications.User, error) {
	query := `SELECT id, email, COALESCE(username, '') as username FROM users WHERE id = $1`

	var user notifications.User
	if err := repo.db.QueryRowxContext(ctx, query, id).StructScan(&user); err != nil {
		return notifications.User{}, errors.Wrap(repoerr.ErrNotFound, err)
	}

	return user, nil
}
