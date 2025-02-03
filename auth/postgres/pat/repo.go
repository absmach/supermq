// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package pat

import (
	"context"
	"database/sql"
	"time"

	"github.com/absmach/supermq/auth"
	"github.com/absmach/supermq/pkg/errors"
	repoerr "github.com/absmach/supermq/pkg/errors/repository"
	"github.com/absmach/supermq/pkg/postgres"
	"github.com/lib/pq"
)


var _ auth.PATSRepository = (*patRepo)(nil)

type patRepo struct {
	db    postgres.Database
	cache auth.Cache
}

func NewPATSRepository(db postgres.Database, cache auth.Cache) auth.PATSRepository {
	return &patRepo{
		db:    db,
		cache: cache,
	}
}

func (pr *patRepo) Save(ctx context.Context, pat auth.PAT) error {
	record, err := patToDBRecords(pat)
	if err != nil {
		return errors.Wrap(repoerr.ErrCreateEntity, err)
	}

	q := `
		INSERT INTO pats (
			id, user_id, name, description, secret, issued_at, expires_at, 
			updated_at, last_used_at, revoked, revoked_at,
			scopes_data, allowed_operations, entity_ids, domains, entity_types, metadata
		) VALUES (
			:id, :user_id, :name, :description, :secret, :issued_at, :expires_at,
			:updated_at, :last_used_at, :revoked, :revoked_at,
			:scopes_data, :allowed_operations, :entity_ids, :domains, :entity_types, :metadata
		)`

	row, err := pr.db.NamedQueryContext(ctx, q, record)
	if err != nil {
		return postgres.HandleError(repoerr.ErrCreateEntity, err)
	}
	defer row.Close()

	if err := pr.cache.Save(ctx, pat.Secret, pat.ID, pat.Scope); err != nil {
		return errors.Wrap(repoerr.ErrCreateEntity, err)
	}

	return nil
}

func (pr *patRepo) Retrieve(ctx context.Context, userID, patID string) (auth.PAT, error) {
	q := `
		SELECT id, user_id, name, description, secret, issued_at, expires_at,
		updated_at, last_used_at, revoked, revoked_at, entity_type, domain_id,
		allowed_operations, entity_ids, is_any_id, scope_hash, constraints, metadata
		FROM pats WHERE user_id = $1 AND id = $2`

	rows, err := pr.db.QueryContext(ctx, q, userID, patID)
	if err != nil {
		return auth.PAT{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}
	defer rows.Close()

	var record dbPat
	if rows.Next() {
		var r dbPat
		if err := rows.StructScan(&record); err != nil {
			return auth.PAT{}, errors.Wrap(repoerr.ErrViewEntity, err)
		}
		return toAuthPat(record)
	}

	return auth.PAT{}, repoerr.ErrNotFound
}

func (pr *patRepo) RetrieveAll(ctx context.Context, userID string, pm auth.PATSPageMeta) (auth.PATSPage, error) {
	q := `
		SELECT DISTINCT id FROM pats WHERE user_id = $1
		ORDER BY issued_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := pr.db.QueryContext(ctx, q, userID, pm.Limit, pm.Offset)
	if err != nil {
		return auth.PATSPage{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}
	defer rows.Close()

	var patIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return auth.PATSPage{}, errors.Wrap(repoerr.ErrViewEntity, err)
		}
		patIDs = append(patIDs, id)
	}

	q = `SELECT COUNT(*) FROM (SELECT DISTINCT id FROM pats WHERE user_id = $1) AS t`
	var total uint64
	if err := pr.db.QueryRowContext(ctx, q, userID).Scan(&total); err != nil {
		return auth.PATSPage{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}

	var pats []auth.PAT
	for _, id := range patIDs {
		pat, err := pr.Retrieve(ctx, userID, id)
		if err != nil {
			return auth.PATSPage{}, err
		}
		pats = append(pats, pat)
	}

	return auth.PATSPage{
		PATS:   pats,
		Total:  total,
		Offset: pm.Offset,
		Limit:  pm.Limit,
	}, nil
}

func (pr *patRepo) RetrieveSecretAndRevokeStatus(ctx context.Context, userID, patID string) (string, bool, bool, error) {
	q := `
		SELECT secret, revoked, expires_at 
		FROM pats 
		WHERE user_id = $1 AND id = $2 
		LIMIT 1`

	var secret string
	var revoked bool
	var expiresAt time.Time

	err := pr.db.QueryRowContext(ctx, q, userID, patID).Scan(&secret, &revoked, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", true, true, repoerr.ErrNotFound
		}
		return "", true, true, errors.Wrap(repoerr.ErrViewEntity, err)
	}

	expired := time.Now().After(expiresAt)
	return secret, revoked, expired, nil
}

func (pr *patRepo) UpdateName(ctx context.Context, userID, patID, name string) (auth.PAT, error) {
	q := `
		UPDATE pats 
		SET name = $1, updated_at = $2
		WHERE user_id = $3 AND id = $4`

	res, err := pr.db.ExecContext(ctx, q, name, time.Now(), userID, patID)
	if err != nil {
		return auth.PAT{}, errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		return auth.PAT{}, errors.Wrap(repoerr.ErrUpdateEntity, err)
	}
	if cnt == 0 {
		return auth.PAT{}, repoerr.ErrNotFound
	}

	return pr.Retrieve(ctx, userID, patID)
}

func (pr *patRepo) UpdateDescription(ctx context.Context, userID, patID, description string) (auth.PAT, error) {
	q := `
		UPDATE pats 
		SET description = $1, updated_at = $2
		WHERE user_id = $3 AND id = $4`

	res, err := pr.db.ExecContext(ctx, q, description, time.Now(), userID, patID)
	if err != nil {
		return auth.PAT{}, errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		return auth.PAT{}, errors.Wrap(repoerr.ErrUpdateEntity, err)
	}
	if cnt == 0 {
		return auth.PAT{}, repoerr.ErrNotFound
	}

	return pr.Retrieve(ctx, userID, patID)
}

func (pr *patRepo) UpdateTokenHash(ctx context.Context, userID, patID, tokenHash string, expiryAt time.Time) (auth.PAT, error) {
	q := `
		UPDATE pats 
		SET secret = $1, expires_at = $2, updated_at = $3
		WHERE user_id = $4 AND id = $5`

	res, err := pr.db.ExecContext(ctx, q, tokenHash, expiryAt, time.Now(), userID, patID)
	if err != nil {
		return auth.PAT{}, errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		return auth.PAT{}, errors.Wrap(repoerr.ErrUpdateEntity, err)
	}
	if cnt == 0 {
		return auth.PAT{}, repoerr.ErrNotFound
	}

	pat, err := pr.Retrieve(ctx, userID, patID)
	if err != nil {
		return auth.PAT{}, err
	}

	// Update cache with new token hash
	if err := pr.cache.Save(ctx, tokenHash, patID, pat.Scope); err != nil {
		return auth.PAT{}, errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	return pat, nil
}

func (pr *patRepo) Revoke(ctx context.Context, userID, patID string) error {
	q := `
		UPDATE pats 
		SET revoked = true, revoked_at = $1
		WHERE user_id = $2 AND id = $3`

	res, err := pr.db.ExecContext(ctx, q, time.Now(), userID, patID)
	if err != nil {
		return errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(repoerr.ErrUpdateEntity, err)
	}
	if cnt == 0 {
		return repoerr.ErrNotFound
	}

	// Remove from cache
	if err := pr.cache.Remove(ctx, patID); err != nil {
		return errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	return nil
}

func (pr *patRepo) Reactivate(ctx context.Context, userID, patID string) error {
	q := `
		UPDATE pats 
		SET revoked = false, revoked_at = NULL
		WHERE user_id = $1 AND id = $2`

	res, err := pr.db.ExecContext(ctx, q, userID, patID)
	if err != nil {
		return errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(repoerr.ErrUpdateEntity, err)
	}
	if cnt == 0 {
		return repoerr.ErrNotFound
	}

	// Re-cache the PAT
	pat, err := pr.Retrieve(ctx, userID, patID)
	if err != nil {
		return err
	}

	if err := pr.cache.Save(ctx, pat.Secret, patID, pat.Scope); err != nil {
		return errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	return nil
}

func (pr *patRepo) Remove(ctx context.Context, userID, patID string) error {
	q := `DELETE FROM pats WHERE user_id = $1 AND id = $2`

	res, err := pr.db.ExecContext(ctx, q, userID, patID)
	if err != nil {
		return postgres.HandleError(repoerr.ErrRemoveEntity, err)
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}
	if cnt == 0 {
		return repoerr.ErrNotFound
	}

	// Remove from cache
	if err := pr.cache.Remove(ctx, patID); err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}

	return nil
}

func (pr *patRepo) AddScopeEntry(ctx context.Context, userID, patID string, platformEntityType auth.PlatformEntityType, optionalDomainID string, optionalDomainEntityType auth.DomainEntityType, operation auth.OperationType, entityIDs ...string) (auth.Scope, error) {
	// First retrieve existing PAT
	pat, err := pr.Retrieve(ctx, userID, patID)
	if err != nil {
		return auth.Scope{}, err
	}

	// Add new scope entry
	if err := pat.Scope.Add(platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...); err != nil {
		return auth.Scope{}, errors.Wrap(repoerr.ErrCreateEntity, err)
	}

	// Convert to DB records
	records, err := patToDBRecords(pat)
	if err != nil {
		return auth.Scope{}, errors.Wrap(repoerr.ErrCreateEntity, err)
	}

	// Start transaction
	tx, err := pr.db.BeginTx(ctx, nil)
	if err != nil {
		return auth.Scope{}, errors.Wrap(repoerr.ErrCreateEntity, err)
	}

	// Delete existing scopes
	if _, err := tx.ExecContext(ctx, `DELETE FROM pats WHERE user_id = $1 AND id = $2`, userID, patID); err != nil {
		tx.Rollback()
		return auth.Scope{}, postgres.HandleError(repoerr.ErrRemoveEntity, err)
	}

	// Insert new records
	q := `
		INSERT INTO pats (
			id, user_id, name, description, secret, issued_at, expires_at,
			updated_at, last_used_at, revoked, revoked_at, entity_type, domain_id,
			allowed_operations, entity_ids, is_any_id, scope_hash, constraints, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
		)`

	for _, record := range records {
		_, err = tx.ExecContext(ctx, q,
			record.ID, record.User, record.Name, record.Description,
			record.Secret, record.IssuedAt, record.ExpiresAt,
			record.UpdatedAt, record.LastUsedAt, record.Revoked,
			record.RevokedAt, record.EntityType, record.Domain,
			pq.Array(record.AllowedOps), pq.Array(record.EntityIDs),
			record.IsAnyID, record.ScopeHash, record.Constraints,
			record.Metadata,
		)
		if err != nil {
			tx.Rollback()
			return auth.Scope{}, postgres.HandleError(repoerr.ErrCreateEntity, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return auth.Scope{}, errors.Wrap(repoerr.ErrCreateEntity, err)
	}

	// Update cache
	if err := pr.cache.Save(ctx, pat.Secret, pat.ID, pat.Scope); err != nil {
		return auth.Scope{}, errors.Wrap(repoerr.ErrCreateEntity, err)
	}

	return pat.Scope, nil
}

func (pr *patRepo) RemoveScopeEntry(ctx context.Context, userID, patID string, platformEntityType auth.PlatformEntityType, optionalDomainID string, optionalDomainEntityType auth.DomainEntityType, operation auth.OperationType, entityIDs ...string) (auth.Scope, error) {
	// First retrieve existing PAT
	pat, err := pr.Retrieve(ctx, userID, patID)
	if err != nil {
		return auth.Scope{}, err
	}

	// Remove scope entry
	if err := pat.Scope.Delete(platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...); err != nil {
		return auth.Scope{}, errors.Wrap(repoerr.ErrRemoveEntity, err)
	}

	// Convert to DB records
	records, err := patToDBRecords(pat)
	if err != nil {
		return auth.Scope{}, errors.Wrap(repoerr.ErrRemoveEntity, err)
	}

	// Start transaction
	tx, err := pr.db.BeginTx(ctx, nil)
	if err != nil {
		return auth.Scope{}, errors.Wrap(repoerr.ErrRemoveEntity, err)
	}

	// Delete existing scopes
	if _, err := tx.ExecContext(ctx, `DELETE FROM pats WHERE user_id = $1 AND id = $2`, userID, patID); err != nil {
		tx.Rollback()
		return auth.Scope{}, postgres.HandleError(repoerr.ErrRemoveEntity, err)
	}

	// Insert new records
	q := `
		INSERT INTO pats (
			id, user_id, name, description, secret, issued_at, expires_at,
			updated_at, last_used_at, revoked, revoked_at, entity_type, domain_id,
			allowed_operations, entity_ids, is_any_id, scope_hash, constraints, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
		)`

	for _, record := range records {
		_, err = tx.ExecContext(ctx, q,
			record.ID, record.User, record.Name, record.Description,
			record.Secret, record.IssuedAt, record.ExpiresAt,
			record.UpdatedAt, record.LastUsedAt, record.Revoked,
			record.RevokedAt, record.EntityType, record.Domain,
			pq.Array(record.AllowedOps), pq.Array(record.EntityIDs),
			record.IsAnyID, record.ScopeHash, record.Constraints,
			record.Metadata,
		)
		if err != nil {
			tx.Rollback()
			return auth.Scope{}, postgres.HandleError(repoerr.ErrRemoveEntity, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return auth.Scope{}, errors.Wrap(repoerr.ErrRemoveEntity, err)
	}

	// Update cache
	if err := pr.cache.Save(ctx, pat.Secret, pat.ID, pat.Scope); err != nil {
		return auth.Scope{}, errors.Wrap(repoerr.ErrRemoveEntity, err)
	}

	return pat.Scope, nil
}

func (pr *patRepo) CheckScopeEntry(ctx context.Context, userID, patID string, platformEntityType auth.PlatformEntityType, optionalDomainID string, optionalDomainEntityType auth.DomainEntityType, operation auth.OperationType, entityIDs ...string) error {
	// First try to get from cache
	if scope, err := pr.cache.ID(ctx, patID); err == nil {
		if !scope.Check(platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...) {
			return repoerr.ErrNotFound
		}
		return nil
	}

	// If not in cache, get from DB
	pat, err := pr.Retrieve(ctx, userID, patID)
	if err != nil {
		return err
	}

	if !pat.Scope.Check(platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...) {
		return repoerr.ErrNotFound
	}
	return nil
}

func (pr *patRepo) RemoveAllScopeEntry(ctx context.Context, userID, patID string) error {
	// First retrieve existing PAT to verify it exists
	pat, err := pr.Retrieve(ctx, userID, patID)
	if err != nil {
		return err
	}

	// Clear all scopes
	pat.Scope = auth.Scope{Domains: make(map[string]auth.DomainScope)}

	// Convert to DB records (should be minimal since scope is empty)
	records, err := patToDBRecords(pat)
	if err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}

	// Start transaction
	tx, err := pr.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}

	// Delete existing scopes
	if _, err := tx.ExecContext(ctx, `DELETE FROM pats WHERE user_id = $1 AND id = $2`, userID, patID); err != nil {
		tx.Rollback()
		return postgres.HandleError(repoerr.ErrRemoveEntity, err)
	}

	// Insert base record without scopes
	q := `
		INSERT INTO pats (
			id, user_id, name, description, secret, issued_at, expires_at,
			updated_at, last_used_at, revoked, revoked_at, entity_type, domain_id,
			allowed_operations, entity_ids, is_any_id, scope_hash, constraints, metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
		)`

	for _, record := range records {
		_, err = tx.ExecContext(ctx, q,
			record.ID, record.User, record.Name, record.Description,
			record.Secret, record.IssuedAt, record.ExpiresAt,
			record.UpdatedAt, record.LastUsedAt, record.Revoked,
			record.RevokedAt, record.EntityType, record.Domain,
			pq.Array(record.AllowedOps), pq.Array(record.EntityIDs),
			record.IsAnyID, record.ScopeHash, record.Constraints,
			record.Metadata,
		)
		if err != nil {
			tx.Rollback()
			return postgres.HandleError(repoerr.ErrRemoveEntity, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}

	// Update cache with empty scope
	if err := pr.cache.Save(ctx, pat.Secret, pat.ID, pat.Scope); err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}

	return nil
}
