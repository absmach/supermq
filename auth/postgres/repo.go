// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"time"

	apiutil "github.com/absmach/supermq/api/http/util"
	"github.com/absmach/supermq/auth"
	"github.com/absmach/supermq/auth/cache"
	"github.com/absmach/supermq/pkg/errors"
	repoerr "github.com/absmach/supermq/pkg/errors/repository"
	"github.com/absmach/supermq/pkg/postgres"
	"github.com/jmoiron/sqlx"
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
		FROM pats WHERE user_id = :user_id AND id = :id`

	saveScopeQuery = `
		INSERT INTO pat_scopes (pat_id, entity_type, optional_domain_id, operation, entity_id)
		VALUES (:pat_id, :entity_type, :optional_domain_id, :operation, :entity_id)`

	retrieveScopesQuery = `
		SELECT entity_type, optional_domain_id, operation, entity_id
		FROM pat_scopes WHERE pat_id = :pat_id`

	deleteScopesQuery = `
		DELETE FROM pat_scopes 
		WHERE pat_id = :pat_id 
			AND entity_type = :entity_type 
			AND optional_domain_id = :optional_domain_id 
			AND operation = :operation 
			AND entity_id = :entity_id`

	deleteAllScopesQuery = `DELETE FROM pat_scopes WHERE pat_id = :pat_id`

	checkWildcardQuery = `
		SELECT EXISTS (
			SELECT 1 FROM pat_scopes 
			WHERE pat_id = :pat_id 
			AND COALESCE(optional_domain_id, '') = COALESCE(:optional_domain_id, '')
			AND entity_type = :entity_type 
			AND operation = :operation
			AND entity_id = '*'
		)`

	deleteSpecificEntriesQuery = `
		DELETE FROM pat_scopes 
		WHERE pat_id = :pat_id 
		AND COALESCE(optional_domain_id, '') = COALESCE(:optional_domain_id, '')
		AND entity_type = :entity_type 
		AND operation = :operation
		AND entity_id != '*'`
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
	dbPat, dbScope, err := patToDBRecords(pat)
	if err != nil {
		return errors.Wrap(repoerr.ErrCreateEntity, err)
	}

	tx, err := pr.db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.Wrap(repoerr.ErrCreateEntity, err)
	}

	defer func() {
		if err != nil {
			if errRollback := tx.Rollback(); errRollback != nil {
				err = errors.Wrap(errors.Wrap(apiutil.ErrRollbackTx, errRollback), err)
			}
		}
	}()

	_, err = tx.NamedExec(saveQuery, dbPat)
	if err != nil {
		return postgres.HandleError(repoerr.ErrCreateEntity, err)
	}

	exists, err := pr.hasWildcardInScope(tx, pat)
	if err != nil {
		return postgres.HandleError(repoerr.ErrCreateEntity, err)
	}

	if exists {
		return nil
	}

	_, err = tx.NamedExec(saveScopeQuery, dbScope)
	if err != nil {
		return postgres.HandleError(repoerr.ErrCreateEntity, err)
	}

	if err := tx.Commit(); err != nil {
		return postgres.HandleError(repoerr.ErrCreateEntity, err)
	}

	return nil
}

func (pr *patRepo) hasWildcardInScope(wrapper *sqlx.Tx, pat auth.PAT) (bool, error) {
	for _, sc := range pat.Scope {
		if sc.EntityId == auth.AnyIDs {
			wildcardScope := toDBScope(pat.ID, sc.EntityType, sc.OptionalDomainId, sc.Operation, auth.AnyIDs)
			_, err := wrapper.NamedExec(saveScopeQuery, wildcardScope)
			if err != nil {
				return false, postgres.HandleError(repoerr.ErrCreateEntity, err)
			}
		}
	}
	if err := wrapper.Commit(); err != nil {
		return false, postgres.HandleError(repoerr.ErrCreateEntity, err)
	}
	return true, nil
}

func (pr *patRepo) Retrieve(ctx context.Context, userID, patID string) (auth.PAT, error) {
	pat, err := pr.retrieveFromDB(ctx, userID, patID)
	if err != nil {
		return auth.PAT{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}
	if err := pr.cache.Save(ctx, pat); err != nil {
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
		SET name = :name, updated_at = :updated_at
		WHERE user_id = :user_id AND id = :id`

	upm := dbPatPagemeta{
		User:      userID,
		ID:        patID,
		Name:      name,
		UpdatedAt: time.Now(),
	}
	tx, err := pr.db.BeginTxx(ctx, nil)
	if err != nil {
		return auth.PAT{}, errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	defer func() {
		if err != nil {
			if errRollback := tx.Rollback(); errRollback != nil {
				err = errors.Wrap(errors.Wrap(apiutil.ErrRollbackTx, errRollback), err)
			}
		}
	}()

	res, err := tx.NamedExec(q, upm)
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

	if err := tx.Commit(); err != nil {
		return auth.PAT{}, errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	pat, err := pr.retrieveFromDB(ctx, userID, patID)
	if err != nil {
		return auth.PAT{}, err
	}

	if err := pr.cache.Save(ctx, pat); err != nil {
		return auth.PAT{}, errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	return pat, nil
}

func (pr *patRepo) UpdateDescription(ctx context.Context, userID, patID, description string) (auth.PAT, error) {
	q := `
		UPDATE pats 
		SET description = :description, updated_at = :updated_at
		WHERE user_id = :user_id AND id = :id`

	upm := dbPatPagemeta{
		User:        userID,
		ID:          patID,
		UpdatedAt:   time.Now(),
		Description: description,
	}

	tx, err := pr.db.BeginTxx(ctx, nil)
	if err != nil {
		return auth.PAT{}, errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	defer func() {
		if err != nil {
			if errRollback := tx.Rollback(); errRollback != nil {
				err = errors.Wrap(errors.Wrap(apiutil.ErrRollbackTx, errRollback), err)
			}
		}
	}()

	res, err := tx.NamedExec(q, upm)
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

	if err := pr.cache.Save(ctx, pat); err != nil {
		return auth.PAT{}, errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	return pat, nil
}

func (pr *patRepo) UpdateTokenHash(ctx context.Context, userID, patID, tokenHash string, expiryAt time.Time) (auth.PAT, error) {
	q := `
		UPDATE pats 
		SET secret = :secret, expires_at = :expires_at, updated_at = :updated_at
		WHERE user_id = :user_id AND id = :id`

	upm := dbPatPagemeta{
		User:      userID,
		ID:        patID,
		UpdatedAt: time.Now(),
		ExpiresAt: expiryAt,
		Secret:    tokenHash,
	}

	tx, err := pr.db.BeginTxx(ctx, nil)
	if err != nil {
		return auth.PAT{}, errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	defer func() {
		if err != nil {
			if errRollback := tx.Rollback(); errRollback != nil {
				err = errors.Wrap(errors.Wrap(apiutil.ErrRollbackTx, errRollback), err)
			}
		}
	}()
	res, err := tx.NamedExec(q, upm)
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

	if err := pr.cache.Save(ctx, pat); err != nil {
		return auth.PAT{}, errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	return pat, nil
}

func (pr *patRepo) Revoke(ctx context.Context, userID, patID string) error {
	q := `
		UPDATE pats 
		SET revoked = true, revoked_at = :revoked_at
		WHERE user_id = :user_id AND id = :id`

	upm := dbPatPagemeta{
		User:      userID,
		ID:        patID,
		ExpiresAt: time.Now(),
	}

	tx, err := pr.db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	defer func() {
		if err != nil {
			if errRollback := tx.Rollback(); errRollback != nil {
				err = errors.Wrap(errors.Wrap(apiutil.ErrRollbackTx, errRollback), err)
			}
		}
	}()

	res, err := tx.NamedExec(q, upm)
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
		WHERE user_id = :user_id AND id = :id`

	upm := dbPatPagemeta{
		User: userID,
		ID:   patID,
	}

	tx, err := pr.db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	defer func() {
		if err != nil {
			if errRollback := tx.Rollback(); errRollback != nil {
				err = errors.Wrap(errors.Wrap(apiutil.ErrRollbackTx, errRollback), err)
			}
		}
	}()
	res, err := tx.NamedExec(q, upm)
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

	pat, err := pr.retrieveFromDB(ctx, userID, patID)
	if err != nil {
		return err
	}

	if err := pr.cache.Save(ctx, pat); err != nil {
		return errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	return nil
}

func (pr *patRepo) Remove(ctx context.Context, userID, patID string) error {
	q := `DELETE FROM pats WHERE user_id = :user_id AND id = :id`
	upm := dbPatPagemeta{
		User: userID,
		ID:   patID,
	}

	tx, err := pr.db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}

	defer func() {
		if err != nil {
			if errRollback := tx.Rollback(); errRollback != nil {
				err = errors.Wrap(errors.Wrap(apiutil.ErrRollbackTx, errRollback), err)
			}
		}
	}()
	res, err := tx.NamedExec(q, upm)
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

func (pr *patRepo) AddScopeEntry(ctx context.Context, userID, patID string, entityType auth.EntityType, optionalDomainID string, operation auth.Operation, entityIDs ...string) ([]auth.Scope, error) {
	tx, err := pr.db.BeginTxx(ctx, nil)
	if err != nil {
		return []auth.Scope{}, errors.Wrap(repoerr.ErrRemoveEntity, err)
	}

	defer func() {
		if err != nil {
			if errRollback := tx.Rollback(); errRollback != nil {
				err = errors.Wrap(errors.Wrap(apiutil.ErrRollbackTx, errRollback), err)
			}
		}
	}()

	checkScope := dbScope{
		PatID:            patID,
		EntityType:       entityType.String(),
		OptionalDomainId: optionalDomainID,
		Operation:        operation.String(),
		EntityID:         auth.AnyIDs,
	}

	existsDB, err := pr.checkWildcardInDB(ctx, checkScope)
	if err != nil {
		return []auth.Scope{}, err
	}

	if existsDB {
		pat, err := pr.retrieveFromDB(ctx, userID, patID)
		if err != nil {
			return []auth.Scope{}, err
		}
		return pat.Scope, nil
	}

	wildcardScope := toDBScope(patID, entityType, optionalDomainID, operation, auth.AnyIDs)
	existsID, err := pr.hasWildcardInIDs(ctx, checkScope, wildcardScope, entityIDs...)
	if err != nil {
		return []auth.Scope{}, err
	}
	if existsID {
		pat, err := pr.retrieveFromDB(ctx, userID, patID)
		if err != nil {
			return []auth.Scope{}, err
		}
		return pat.Scope, nil
	}

	scopes := toDBScope(patID, entityType, optionalDomainID, operation, entityIDs...)
	_, err = tx.NamedQuery(saveScopeQuery, scopes)
	if err != nil {
		return []auth.Scope{}, postgres.HandleError(repoerr.ErrUpdateEntity, err)
	}

	if err := tx.Commit(); err != nil {
		return []auth.Scope{}, postgres.HandleError(repoerr.ErrUpdateEntity, err)
	}

	pat, err := pr.retrieveFromDB(ctx, userID, patID)
	if err != nil {
		return []auth.Scope{}, err
	}

	if err := pr.cache.Save(ctx, pat); err != nil {
		return []auth.Scope{}, errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	return pat.Scope, nil
}

// check for wildcard in the entity IDs.
func (pr *patRepo) hasWildcardInIDs(ctx context.Context, checkScope dbScope, wildcardScope []dbScope, entityIDs ...string) (bool, error) {
	tx, err := pr.db.BeginTxx(ctx, nil)
	if err != nil {
		return false, errors.Wrap(repoerr.ErrRemoveEntity, err)
	}

	defer func() {
		if err != nil {
			if errRollback := tx.Rollback(); errRollback != nil {
				err = errors.Wrap(errors.Wrap(apiutil.ErrRollbackTx, errRollback), err)
			}
		}
	}()
	for _, entityID := range entityIDs {
		if entityID == auth.AnyIDs {
			_, err = tx.NamedExec(deleteSpecificEntriesQuery, checkScope)
			if err != nil {
				return false, errors.Wrap(repoerr.ErrUpdateEntity, err)
			}

			_, err = tx.NamedQuery(saveScopeQuery, wildcardScope)
			if err != nil {
				return false, postgres.HandleError(repoerr.ErrUpdateEntity, err)
			}

			if err := tx.Commit(); err != nil {
				return false, postgres.HandleError(repoerr.ErrUpdateEntity, err)
			}

			return true, nil
		}
	}

	return false, nil
}

// check for wildcard in the database.
func (pr *patRepo) checkWildcardInDB(ctx context.Context, checkScope dbScope) (bool, error) {
	rows, err := pr.db.NamedQueryContext(ctx, checkWildcardQuery, checkScope)
	if err != nil {
		return false, errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	var exists bool
	if rows.Next() {
		if err := rows.Scan(&exists); err != nil {
			return false, errors.Wrap(repoerr.ErrUpdateEntity, err)
		}
	}

	return exists, nil
}

func (pr *patRepo) RemoveScopeEntry(ctx context.Context, userID, patID string, entityType auth.EntityType, optionalDomainID string, operation auth.Operation, entityIDs ...string) ([]auth.Scope, error) {
	scopes := toDBScope(patID, entityType, optionalDomainID, operation, entityIDs...)

	_, err := pr.db.NamedQueryContext(ctx, deleteScopesQuery, scopes)
	if err != nil {
		return []auth.Scope{}, errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	pat, err := pr.retrieveFromDB(ctx, userID, patID)
	if err != nil {
		return []auth.Scope{}, err
	}

	if err := pr.cache.Save(ctx, pat); err != nil {
		return []auth.Scope{}, errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	return pat.Scope, nil
}

func (pr *patRepo) CheckScopeEntry(ctx context.Context, userID, patID string, entityType auth.EntityType, optionalDomainID string, operation auth.Operation, entityIDs ...string) error {
	for _, entityID := range entityIDs {
		key := cache.GenerateKey(patID, optionalDomainID, entityType, operation, entityID)
		authorized, err := pr.cache.CheckScope(ctx, key)
		if err != nil {
			break
		}

		if authorized {
			return nil
		}
	}

	pat, err := pr.retrieveFromDB(ctx, userID, patID)
	if err != nil {
		return err
	}

	for _, entityID := range entityIDs {
		if !pat.CheckAccess(entityType, optionalDomainID, operation, entityID) {
			return repoerr.ErrNotFound
		}
	}
	return nil
}

func (pr *patRepo) RemoveAllScopeEntry(ctx context.Context, userID, patID string) error {
	pat, err := pr.retrieveFromDB(ctx, userID, patID)
	if err != nil {
		return err
	}

	pat.Scope = []auth.Scope{}

	scope := toDBPatScope(pat)

	_, err = pr.db.NamedQueryContext(ctx, deleteAllScopesQuery, scope)
	if err != nil {
		return postgres.HandleError(repoerr.ErrUpdateEntity, err)
	}

	if err := pr.cache.Save(ctx, pat); err != nil {
		return errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	return nil
}

func (pr *patRepo) RetrieveScope(ctx context.Context, pm auth.ScopesPageMeta) (scopes auth.ScopesPage, err error) {
	pat, err := pr.retrieveFromDB(ctx, pm.UserID, pm.PatID)
	if err != nil {
		return auth.ScopesPage{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}
	if err := pr.cache.Save(ctx, pat); err != nil {
		return auth.ScopesPage{}, err
	}

	return auth.ScopesPage{
		Total:  uint64(len(pat.Scope)),
		Scopes: pat.Scope,
	}, nil
}

func (pr *patRepo) retrieveFromDB(ctx context.Context, userID, patID string) (auth.PAT, error) {
	dbs := dbPatPagemeta{
		PatID: patID,
	}
	scopeRows, err := pr.db.NamedQueryContext(ctx, retrieveScopesQuery, dbs)
	if err != nil {
		return auth.PAT{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}
	defer scopeRows.Close()

	var scopes []dbScope
	for scopeRows.Next() {
		var scope dbScope
		if err := scopeRows.StructScan(&scope); err != nil {
			return auth.PAT{}, errors.Wrap(repoerr.ErrViewEntity, err)
		}
		scopes = append(scopes, scope)
	}

	dbp := dbPatPagemeta{
		ID:   patID,
		User: userID,
	}

	rows, err := pr.db.NamedQueryContext(ctx, retrieveQuery, dbp)
	if err != nil {
		return auth.PAT{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}
	defer rows.Close()

	var record dbPat
	if rows.Next() {
		if err := rows.StructScan(&record); err != nil {
			return auth.PAT{}, errors.Wrap(repoerr.ErrViewEntity, err)
		}
		return toAuthPat(record, scopes)
	}

	return auth.PAT{}, repoerr.ErrNotFound
}
