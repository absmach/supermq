// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/absmach/supermq/auth"
	"github.com/absmach/supermq/pkg/errors"
	repoerr "github.com/absmach/supermq/pkg/errors/repository"
	"github.com/absmach/supermq/pkg/postgres"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

var errScanJWKData = errors.New("cannot scan into jwkData")

var _ auth.PublicKeyRepository = (*publicKeyRepo)(nil)

type publicKeyRepo struct {
	db postgres.Database
}

// NewPublicKeyRepo creates a new instance of PublicKeyRepository.
func NewPublicKeyRepo(db postgres.Database) auth.PublicKeyRepository {
	return &publicKeyRepo{
		db: db,
	}
}

type dbPublicKey struct {
	Kid       string     `db:"kid"`
	JWKData   jwkData    `db:"jwk_data"`
	CreatedAt time.Time  `db:"created_at"`
	RetiredAt *time.Time `db:"retired_at"`
	Status    int        `db:"status"`
}

type jwkData struct {
	jwk.Key
}

func (j *jwkData) Scan(value any) error {
	if value == nil {
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errScanJWKData
	}

	key, err := jwk.ParseKey(bytes)
	if err != nil {
		return err
	}

	j.Key = key
	return nil
}

func (j jwkData) Value() (driver.Value, error) {
	if j.Key == nil {
		return nil, nil
	}

	return json.Marshal(j.Key)
}

func (pkr *publicKeyRepo) Save(ctx context.Context, key auth.PublicKey) error {
	q := `
		INSERT INTO public_keys (kid, jwk_data, created_at, retired_at, status)
		VALUES (:kid, :jwk_data, :created_at, :retired_at, :status)
		ON CONFLICT (kid) DO UPDATE SET
			jwk_data = EXCLUDED.jwk_data,
			retired_at = EXCLUDED.retired_at,
			status = EXCLUDED.status`

	dbKey := dbPublicKey{
		Kid:       key.Kid,
		JWKData:   jwkData{Key: key.JWKData},
		CreatedAt: key.CreatedAt,
		RetiredAt: key.RetiredAt,
		Status:    int(key.Status),
	}

	_, err := pkr.db.NamedExecContext(ctx, q, dbKey)
	if err != nil {
		return postgres.HandleError(repoerr.ErrCreateEntity, err)
	}

	return nil
}

func (pkr *publicKeyRepo) Retrieve(ctx context.Context, kid string) (auth.PublicKey, error) {
	q := `SELECT kid, jwk_data, created_at, retired_at, status FROM public_keys WHERE kid = $1`

	var dbKey dbPublicKey
	if err := pkr.db.QueryRowxContext(ctx, q, kid).StructScan(&dbKey); err != nil {
		return auth.PublicKey{}, postgres.HandleError(repoerr.ErrViewEntity, err)
	}

	return toAuthPublicKey(dbKey), nil
}

func (pkr *publicKeyRepo) RetrieveActive(ctx context.Context) ([]auth.PublicKey, error) {
	q := `SELECT kid, jwk_data, created_at, retired_at, status FROM public_keys WHERE status = $1 ORDER BY created_at DESC`

	rows, err := pkr.db.QueryxContext(ctx, q, int(auth.ActiveKeyStatus))
	if err != nil {
		return nil, postgres.HandleError(repoerr.ErrViewEntity, err)
	}
	defer rows.Close()

	var keys []auth.PublicKey
	for rows.Next() {
		var dbKey dbPublicKey
		if err := rows.StructScan(&dbKey); err != nil {
			return nil, postgres.HandleError(repoerr.ErrViewEntity, err)
		}
		keys = append(keys, toAuthPublicKey(dbKey))
	}

	return keys, nil
}

func (pkr *publicKeyRepo) Retire(ctx context.Context, kid string) error {
	now := time.Now().UTC()
	q := `UPDATE public_keys SET status = $1, retired_at = $2 WHERE kid = $3`

	result, err := pkr.db.ExecContext(ctx, q, int(auth.RetiredKeyStatus), now, kid)
	if err != nil {
		return postgres.HandleError(repoerr.ErrUpdateEntity, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return postgres.HandleError(repoerr.ErrUpdateEntity, err)
	}

	if rowsAffected == 0 {
		return repoerr.ErrNotFound
	}

	return nil
}

func (pkr *publicKeyRepo) PurgeExpired(ctx context.Context, expiredBefore time.Time) error {
	q := `DELETE FROM public_keys WHERE status = $1 AND retired_at IS NOT NULL AND retired_at < $2`

	_, err := pkr.db.ExecContext(ctx, q, int(auth.RetiredKeyStatus), expiredBefore)
	if err != nil {
		return postgres.HandleError(repoerr.ErrRemoveEntity, err)
	}

	return nil
}

func toAuthPublicKey(dbKey dbPublicKey) auth.PublicKey {
	return auth.PublicKey{
		Kid:       dbKey.Kid,
		JWKData:   dbKey.JWKData.Key,
		CreatedAt: dbKey.CreatedAt,
		RetiredAt: dbKey.RetiredAt,
		Status:    auth.PublicKeyStatus(dbKey.Status),
	}
}
