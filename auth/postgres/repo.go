// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/absmach/supermq/auth"
	"github.com/absmach/supermq/pkg/errors"
	repoerr "github.com/absmach/supermq/pkg/errors/repository"
	"github.com/absmach/supermq/pkg/postgres"
	"github.com/lib/pq"
)

var _ auth.PATSRepository = (*patRepo)(nil)

const (
	saveQuery = `
	INSERT INTO pats (
		id, user_id, name, description, secret, issued_at, expires_at, 
		updated_at, last_used_at, revoked, revoked_at
	) VALUES (
		:id, :user_id, :name, :description, :secret, :issued_at, :expires_at,
		:updated_at, :last_used_at, :revoked, :revoked_at
	)`

	retrieveQuery = `
		SELECT 
		id, user_id, name, description, secret, issued_at, expires_at,
		updated_at, last_used_at, revoked, revoked_at
		FROM pats WHERE user_id = $1 AND id = $2`

	saveScopeQuery = `
		INSERT INTO pat_scopes (pat_id, platform_type, domain_id, domain_type, operation_type, entity_ids)
		VALUES (:pat_id, :platform_type, :domain_id, :domain_type, :operation_type, :entity_ids)`

	updateScopeQuery = `
		UPDATE pat_scopes SET
			platform_type = :platform_type, 
			domain_id = :domain_id, 
			domain_type = :domain_type, 
			operation_type = :operation_type, 
			entity_ids = :entity_ids
		WHERE pat_id = :pat_id`

	retrieveScopesQuery = `
		SELECT platform_type, domain_id, domain_type, operation_type, entity_ids
		FROM pat_scopes WHERE pat_id = $1`

	deleteScopesQuery = `DELETE FROM pat_scopes WHERE pat_id = $1`
)

type patRepo struct {
	db    postgres.Database
	cache auth.Cache
}

func NewPatRepo(db postgres.Database, cache auth.Cache) auth.PATSRepository {
	return &patRepo{
		db:    db,
		cache: cache,
	}
}

func (pr *patRepo) Save(ctx context.Context, pat auth.PAT) error {
	record, scope, err := patToDBRecords(pat)
	if err != nil {
		return errors.Wrap(repoerr.ErrCreateEntity, err)
	}

	row, err := pr.db.NamedQueryContext(ctx, saveQuery, record)
	if err != nil {
		return postgres.HandleError(repoerr.ErrCreateEntity, err)
	}
	defer row.Close()

	scopeRow, err := pr.db.NamedQueryContext(ctx, saveScopeQuery, scope)
	if err != nil {
		return postgres.HandleError(repoerr.ErrCreateEntity, err)
	}
	defer scopeRow.Close()

	if err := pr.cache.Save(ctx, pat.ID, pat); err != nil {
		return errors.Wrap(repoerr.ErrCreateEntity, err)
	}

	return nil
}

func (pr *patRepo) Retrieve(ctx context.Context, userID, patID string) (auth.PAT, error) {
	if pat, err := pr.cache.ID(ctx, patID); err == nil {
		return pat, nil
	}

	pat, err := pr.retrieveFromDB(ctx, userID, patID)
	if err != nil {
		return auth.PAT{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}
	if err := pr.cache.Save(ctx, pat.ID, pat); err != nil {
		return auth.PAT{}, err
	}

	return pat, nil
}

func (pr *patRepo) RetrieveAll(ctx context.Context, userID string, pm auth.PATSPageMeta) (auth.PATSPage, error) {
	q := `
		SELECT 
		p.id, p.user_id, p.name, p.description, p.issued_at, p.expires_at,
		p.updated_at, p.revoked, p.revoked_at
		FROM pats p WHERE user_id = :user_id
		ORDER BY issued_at DESC
		LIMIT :limit OFFSET :offset`

	dbPage := toDBAuthPage(userID, pm)

	rows, err := pr.db.NamedQueryContext(ctx, q, dbPage)
	if err != nil {
		return auth.PATSPage{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}
	defer rows.Close()

	type data struct {
		ID          string    `db:"id,omitempty"`
		User        string    `db:"user_id,omitempty"`
		Name        string    `db:"name,omitempty"`
		Description string    `db:"description,omitempty"`
		IssuedAt    time.Time `db:"issued_at,omitempty"`
		ExpiresAt   time.Time `db:"expires_at,omitempty"`
		UpdatedAt   time.Time `db:"updated_at,omitempty"`
		Revoked     bool      `db:"revoked,omitempty"`
		RevokedAt   time.Time `db:"revoked_at,omitempty"`
	}

	var items []auth.PAT
	for rows.Next() {
		var pat data
		if err := rows.StructScan(&pat); err != nil {
			return auth.PATSPage{}, errors.Wrap(repoerr.ErrViewEntity, err)
		}
		items = append(items, auth.PAT{
			ID:          pat.ID,
			User:        pat.User,
			Name:        pat.Name,
			Description: pat.Description,
			IssuedAt:    pat.IssuedAt,
			ExpiresAt:   pat.ExpiresAt,
			UpdatedAt:   pat.UpdatedAt,
			Revoked:     pat.Revoked,
			RevokedAt:   pat.RevokedAt,
		})
	}

	cq := `SELECT COUNT(*) FROM pats p WHERE user_id = :user_id`

	total, err := postgres.Total(ctx, pr.db, cq, dbPage)
	if err != nil {
		return auth.PATSPage{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}

	page := auth.PATSPage{
		PATS:   items,
		Total:  total,
		Offset: pm.Offset,
		Limit:  pm.Limit,
	}
	return page, nil
}

func (pr *patRepo) RetrieveSecretAndRevokeStatus(ctx context.Context, userID, patID string) (string, bool, bool, error) {
	if pat, err := pr.cache.ID(ctx, patID); err == nil {
		expired := time.Now().After(pat.ExpiresAt)
		return pat.Secret, pat.Revoked, expired, nil
	}

	q := `
		SELECT p.secret, p.revoked, p.expires_at 
		FROM pats p
		WHERE user_id = $1 AND id = $2`

	rows, err := pr.db.QueryContext(ctx, q, userID, patID)
	if err != nil {
		return "", true, true, postgres.HandleError(repoerr.ErrNotFound, err)
	}
	defer rows.Close()

	var secret string
	var revoked bool
	var expiresAt time.Time

	if !rows.Next() {
		return "", true, true, repoerr.ErrNotFound
	}

	if err := rows.Scan(&secret, &revoked, &expiresAt); err != nil {
		return "", true, true, postgres.HandleError(repoerr.ErrNotFound, err)
	}

	expired := time.Now().After(expiresAt)
	return secret, revoked, expired, nil
}

func (pr *patRepo) UpdateName(ctx context.Context, userID, patID, name string) (auth.PAT, error) {
	q := `
		UPDATE pats p
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

	pat, err := pr.retrieveFromDB(ctx, userID, patID)
	if err != nil {
		return auth.PAT{}, err
	}

	if err := pr.cache.Save(ctx, patID, pat); err != nil {
		return auth.PAT{}, errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	return pat, nil
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

	pat, err := pr.retrieveFromDB(ctx, userID, patID)
	if err != nil {
		return auth.PAT{}, err
	}

	if err := pr.cache.Save(ctx, patID, pat); err != nil {
		return auth.PAT{}, errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	return pat, nil
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

	pat, err := pr.retrieveFromDB(ctx, userID, patID)
	if err != nil {
		return auth.PAT{}, err
	}

	if err := pr.cache.Save(ctx, patID, pat); err != nil {
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

	pat, err := pr.Retrieve(ctx, userID, patID)
	if err != nil {
		return err
	}

	if err := pr.cache.Save(ctx, patID, pat); err != nil {
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

	if err := pr.cache.Remove(ctx, patID); err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}

	return nil
}

func (pr *patRepo) AddScopeEntry(ctx context.Context, userID, patID string, platformEntityType auth.PlatformEntityType, optionalDomainID string, optionalDomainEntityType auth.DomainEntityType, operation auth.OperationType, entityIDs ...string) (auth.Scope, error) {
	pat, err := pr.retrieveFromDB(ctx, userID, patID)
	if err != nil {
		return auth.Scope{}, err
	}

	if err := pat.Scope.Add(platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...); err != nil {
		return auth.Scope{}, errors.Wrap(repoerr.ErrCreateEntity, err)
	}

	_, scope, err := patToDBRecords(pat)
	if err != nil {
		return auth.Scope{}, errors.Wrap(repoerr.ErrCreateEntity, err)
	}

	res, err := pr.db.NamedExecContext(ctx, updateScopeQuery, scope)
	if err != nil {
		return auth.Scope{}, postgres.HandleError(repoerr.ErrUpdateEntity, err)
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		return auth.Scope{}, errors.Wrap(repoerr.ErrUpdateEntity, err)
	}
	if cnt == 0 {
		return auth.Scope{}, repoerr.ErrNotFound
	}

	if err := pr.cache.Save(ctx, pat.ID, pat); err != nil {
		return auth.Scope{}, errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	return pat.Scope, nil
}

func (pr *patRepo) RemoveScopeEntry(ctx context.Context, userID, patID string, platformEntityType auth.PlatformEntityType, optionalDomainID string, optionalDomainEntityType auth.DomainEntityType, operation auth.OperationType, entityIDs ...string) (auth.Scope, error) {
	pat, err := pr.Retrieve(ctx, userID, patID)
	if err != nil {
		return auth.Scope{}, err
	}

	if err := pat.Scope.Delete(platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...); err != nil {
		return auth.Scope{}, errors.Wrap(repoerr.ErrRemoveEntity, err)
	}

	_, scope, err := patToDBRecords(pat)
	if err != nil {
		return auth.Scope{}, errors.Wrap(repoerr.ErrRemoveEntity, err)
	}

	res, err := pr.db.NamedExecContext(ctx, updateScopeQuery, scope)
	if err != nil {

		x := fmt.Sprintf("%+v \n %+v \n %+v \n", pat.Scope, scope, isEmptyScope(pat.Scope))
		return auth.Scope{}, errors.New(x)
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		return auth.Scope{}, errors.Wrap(repoerr.ErrUpdateEntity, err)
	}
	if cnt == 0 {
		return auth.Scope{}, repoerr.ErrNotFound
	}

	if err := pr.cache.Save(ctx, pat.ID, pat); err != nil {
		return auth.Scope{}, errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	return pat.Scope, nil
}

func (pr *patRepo) CheckScopeEntry(ctx context.Context, userID, patID string, platformEntityType auth.PlatformEntityType, optionalDomainID string, optionalDomainEntityType auth.DomainEntityType, operation auth.OperationType, entityIDs ...string) error {
	if pat, err := pr.cache.ID(ctx, patID); err == nil {
		if !pat.Scope.Check(platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...) {
			return repoerr.ErrNotFound
		}
		return nil
	}

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
	pat, err := pr.Retrieve(ctx, userID, patID)
	if err != nil {
		return err
	}

	pat.Scope = auth.Scope{Domains: make(map[string]auth.DomainScope)}

	_, scope, err := patToDBRecords(pat)
	if err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}

	res, err := pr.db.NamedExecContext(ctx, deleteScopesQuery, scope)
	if err != nil {
		return postgres.HandleError(repoerr.ErrUpdateEntity, err)
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(repoerr.ErrUpdateEntity, err)
	}
	if cnt == 0 {
		return repoerr.ErrNotFound
	}

	if err := pr.cache.Save(ctx, pat.ID, pat); err != nil {
		return errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	return nil
}

func (pr *patRepo) retrieveFromDB(ctx context.Context, userID, patID string) (auth.PAT, error) {
	scopeRows, err := pr.db.QueryContext(ctx, retrieveScopesQuery, patID)
	if err != nil {
		return auth.PAT{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}
	defer scopeRows.Close()

	var scopes []dbScope
	var entityIDs pq.StringArray
	for scopeRows.Next() {
		var scope dbScope
		if err := scopeRows.Scan(
			&scope.Platformtype,
			&scope.DomainID,
			&scope.DomainType,
			&scope.OperationType,
			&entityIDs,
		); err != nil {
			return auth.PAT{}, errors.Wrap(repoerr.ErrViewEntity, err)
		}
		scope.EntityIDs = []string(entityIDs)
		scopes = append(scopes, scope)
	}

	rows, err := pr.db.QueryContext(ctx, retrieveQuery, userID, patID)
	if err != nil {
		return auth.PAT{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}
	defer rows.Close()

	var record dbPat
	if rows.Next() {
		if err := rows.Scan(
			&record.ID,
			&record.User,
			&record.Name,
			&record.Description,
			&record.Secret,
			&record.IssuedAt,
			&record.ExpiresAt,
			&record.UpdatedAt,
			&record.LastUsedAt,
			&record.Revoked,
			&record.RevokedAt,
		); err != nil {
			return auth.PAT{}, errors.Wrap(repoerr.ErrViewEntity, err)
		}
		return toAuthPat(record, scopes)
	}

	return auth.PAT{}, repoerr.ErrNotFound
}

func isEmptyScope(scope auth.Scope) bool {
	return len(scope.Users) == 0 &&
		len(scope.Dashboard) == 0 &&
		len(scope.Messaging) == 0 &&
		len(scope.Domains) == 0
}
