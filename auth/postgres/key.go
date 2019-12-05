package postgres

import (
	"context"
	"database/sql"
	"time"

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

// New instantiates a PostgreSQL implementation of key repository.
func New(db Database) auth.KeyRepository {
	return &keyRepository{
		db: db,
	}
}

func (kr keyRepository) Save(ctx context.Context, key auth.Key) (string, error) {
	q := `INSERT INTO keys (id, type, issuer, issued_at, expires_at)
	      VALUES (:id, :type, :issuer, :issued_at, :expires_at)`

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
	q := `SELECT id, type, issuer, issued_at, expires_at FROM keys WHERE issuer = $1 AND id = $2`
	key := dbKey{}
	if err := kr.db.QueryRowxContext(ctx, q, issuer, id).StructScan(&key); err != nil {
		pqErr, ok := err.(*pq.Error)
		if err == sql.ErrNoRows || ok && errInvalid == pqErr.Code.Name() {
			return auth.Key{}, auth.ErrNotFound
		}

		return auth.Key{}, err
	}

	return toKey(key), nil
}

func (kr keyRepository) Remove(ctx context.Context, issuer, id string) error {
	q := `DELETE FROM keys WHERE issuer = $1 AND id = $2`

	if _, err := kr.db.ExecContext(ctx, q, issuer, id); err != nil {
		return err
	}

	return nil
}

type dbKey struct {
	ID        string       `db:"id"`
	Type      uint32       `db:"type"`
	Issuer    string       `db:"issuer"`
	Revoked   bool         `db:"revoked"`
	IssuedAt  time.Time    `db:"issued_at"`
	ExpiresAt sql.NullTime `db:"expires_at"`
}

func toDBKey(key auth.Key) dbKey {
	ret := dbKey{
		ID:       key.ID,
		Type:     key.Type,
		Issuer:   key.Issuer,
		IssuedAt: key.IssuedAt,
	}
	if key.ExpiresAt != nil {
		ret.ExpiresAt = sql.NullTime{Time: *key.ExpiresAt, Valid: true}
	}

	return ret
}

func toKey(key dbKey) auth.Key {
	ret := auth.Key{
		ID:       key.ID,
		Type:     key.Type,
		Issuer:   key.Issuer,
		IssuedAt: key.IssuedAt,
	}
	if key.ExpiresAt.Valid {
		ret.ExpiresAt = &key.ExpiresAt.Time
	}

	return ret
}
