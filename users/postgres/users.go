//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package postgres

import (
	"context"
	"database/sql"
<<<<<<< HEAD
	"database/sql/driver"
=======
>>>>>>> add metadata to users
	"encoding/json"

	"github.com/jmoiron/sqlx"

	"github.com/lib/pq"
	"github.com/mainflux/mainflux/users"
)

var _ users.UserRepository = (*userRepository)(nil)

const errDuplicate = "unique_violation"

type userRepository struct {
	db *sqlx.DB
}

// New instantiates a PostgreSQL implementation of user
// repository.
func New(db *sqlx.DB) users.UserRepository {
	return &userRepository{db}
}

func (ur userRepository) Save(_ context.Context, user users.User) error {
	q := `INSERT INTO users (email, password, metadata) VALUES (:email, :password, :metadata)`

	dbu := toDBUser(user)

	if _, err := ur.db.NamedExec(q, dbu); err != nil {
		if pqErr, ok := err.(*pq.Error); ok && errDuplicate == pqErr.Code.Name() {
			return users.ErrConflict
		}
		return err
	}

	return nil
}

func (ur userRepository) RetrieveByID(_ context.Context, email string) (users.User, error) {
	q := `SELECT password, metadata FROM users WHERE email = $1`

	dbu := dbUser{
		Email: email,
	}
	if err := ur.db.QueryRowx(q, email).StructScan(&dbu); err != nil {
		if err == sql.ErrNoRows {
			return users.User{}, users.ErrNotFound
		}
		return users.User{}, err
	}

	user := toUser(dbu)

	return user, nil
}

// dbMetadata type for handling metadata properly in database/sql
type dbMetadata map[string]interface{}

// Scan - Implement the database/sql scanner interface
func (m *dbMetadata) Scan(value interface{}) error {
	if value == nil {
		m = nil
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		m = &dbMetadata{}
		return users.ErrScanMetadata
	}

	if err := json.Unmarshal(b, m); err != nil {
		m = &dbMetadata{}
		return err
	}

	return nil
}

// Value Implements valuer
func (m dbMetadata) Value() (driver.Value, error) {
	if len(m) == 0 {
		return nil, nil
	}

	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return b, err
}

type dbUser struct {
<<<<<<< HEAD
	Email    string     `db:"email"`
	Password string     `db:"password"`
	Metadata dbMetadata `db:"metadata"`
=======
	Email    string `json:"email"`
	Password string `json:"password"`
	Metadata string `db:"metadata"`
>>>>>>> add metadata to users
}

func toDBUser(u users.User) dbUser {
	data, err := json.Marshal(u.Metadata)
	if err != nil {
		return dbUser{}
	}
	return dbUser{
		Email:    u.Email,
		Password: u.Password,
		Metadata: string(data),
	}
}

func toUser(dbu dbUser) users.User {

	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(dbu.Metadata), &metadata); err != nil {
		return users.User{}
	}

	return users.User{
		Email:    dbu.Email,
		Password: dbu.Password,
		Metadata: metadata,
	}
}
