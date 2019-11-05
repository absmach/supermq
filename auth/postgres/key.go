package postgres

import (
	"context"
	"database/sql"
	"time"

	// required for DB access
	"github.com/lib/pq"
	"github.com/mainflux/mainflux/auth"
)

var _ auth.KeyRepository = (*keyRepository)(nil)

const (
	errDuplicate = "unique_violation"
	errInvalid   = "invalid_text_representation"
)

type keyRepository struct {
	db Database
}

// New instantiates a PostgreSQL implementation of key
// repository.
func New(db Database) auth.KeyRepository {
	return &keyRepository{
		db: db,
	}
}

func (kr keyRepository) Save(ctx context.Context, key auth.Key) (string, error) {
	q := `INSERT INTO keys (id, type, issuer, secret, issued_at, expires_at)
	      VALUES (:id, :type, :issuer, :secret, :issued_at, :expires_at)`

	dbKey := toDBKey(key)
	if _, err := kr.db.NamedExecContext(ctx, q, dbKey); err != nil {

		pqErr, ok := err.(*pq.Error)
		if ok {
			if pqErr.Code.Name() == errDuplicate {
				return "", auth.ErrConflict
			}
		}

		return "", err
	}

	return dbKey.ID, nil
}

func (kr keyRepository) Retrieve(ctx context.Context, issuer, id string) (auth.Key, error) {
	key := auth.Key{
		Issuer: issuer,
		ID:     id,
	}
	q := `SELECT FROM keys VALUES (id, type, issuer, value, issued_at, expires_at) WHERE issuer = $1 AND id = $2`

	if err := kr.db.QueryRowxContext(ctx, q, issuer, id).StructScan(&key); err != nil {
		pqErr, ok := err.(*pq.Error)
		if err == sql.ErrNoRows || ok && errInvalid == pqErr.Code.Name() {
			return auth.Key{}, auth.ErrNotFound
		}

		return auth.Key{}, err
	}

	return auth.Key{}, nil
}

func (kr keyRepository) Remove(ctx context.Context, issuer, id string) error {
	q := `DELETE FROM keys WHERE issuer = $1 AND id = $2`

	if _, err := kr.db.ExecContext(ctx, q, issuer, id); err != nil {
		return err
	}

	return nil
}

type dbKey struct {
	ID        string     `db:"id"`
	Type      uint32     `db:"type"`
	Issuer    string     `db:"issuer"`
	Secret    string     `db:"secret"`
	Revoked   bool       `db:"revoked"`
	IssuedAt  time.Time  `db:"issued_at"`
	ExpiresAt *time.Time `db:"expires_at"`
}

func toDBKey(key auth.Key) dbKey {
	return dbKey{
		ID:        key.ID,
		Type:      key.Type,
		Issuer:    key.Issuer,
		Secret:    key.Secret,
		IssuedAt:  key.IssuedAt,
		ExpiresAt: key.ExpiresAt,
	}
}

func toKey(key dbKey) auth.Key {
	return auth.Key{
		ID:        key.ID,
		Type:      key.Type,
		Issuer:    key.Issuer,
		Secret:    key.Secret,
		IssuedAt:  key.IssuedAt,
		ExpiresAt: key.ExpiresAt,
	}
}
