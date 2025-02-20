// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"time"

	"github.com/absmach/supermq/auth"
	"github.com/absmach/supermq/pkg/errors"
	repoerr "github.com/absmach/supermq/pkg/errors/repository"
	"github.com/absmach/supermq/pkg/postgres"
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
		INSERT INTO pat_scopes (id, pat_id, entity_type, optional_domain_id, operation, entity_id)
		VALUES (:id, :pat_id, :entity_type, :optional_domain_id, :operation, :entity_id)`

	checkScopeQuery = `
        SELECT id, pat_id, entity_type, optional_domain_id, operation, entity_id
        FROM pat_scopes 
        WHERE pat_id = :pat_id 
          AND entity_type = :entity_type
          AND optional_domain_id = :optional_domain_id
          AND operation = :operation
          AND (entity_id = :entity_id OR entity_id = '*')
        LIMIT 1`

	retrieveScopesQuery = `
		SELECT id, pat_id, entity_type, optional_domain_id, operation, entity_id
		FROM pat_scopes WHERE pat_id = :pat_id AND id = :id OFFSET :offset LIMIT :limit`

	deleteScopesQuery = `
		DELETE FROM pat_scopes 
		WHERE id = :scopes_id`

	deleteAllScopesQuery = `DELETE FROM pat_scopes WHERE pat_id = :pat_id`
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
	dbPat, err := toDBPats(pat)
	if err != nil {
		return errors.Wrap(repoerr.ErrCreateEntity, err)
	}

	_, err = pr.db.NamedQueryContext(ctx, saveQuery, dbPat)
	if err != nil {
		return postgres.HandleError(repoerr.ErrCreateEntity, err)
	}

	return nil
}

func (pr *patRepo) Retrieve(ctx context.Context, userID, patID string) (auth.PAT, error) {
	pat, err := pr.retrievePATFromDB(ctx, userID, patID)
	if err != nil {
		return auth.PAT{}, errors.Wrap(repoerr.ErrViewEntity, err)
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

	dbPage := dbPagemeta{
		Limit:  pm.Limit,
		Offset: pm.Offset,
		User:   userID,
	}

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
		WHERE user_id = :user_id AND id = :id
		RETURNING id, user_id, name, description, secret, issued_at, updated_at, expires_at, revoked, revoked_at, last_used_at`

	upm := dbPagemeta{
		User:      userID,
		ID:        patID,
		Name:      name,
		UpdatedAt: time.Now(),
	}

	rows, err := pr.db.NamedQueryContext(ctx, q, upm)
	if err != nil {
		return auth.PAT{}, errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	var pat auth.PAT
	if err := rows.StructScan(&pat); err != nil {
		return auth.PAT{}, errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	return pat, nil
}

func (pr *patRepo) UpdateDescription(ctx context.Context, userID, patID, description string) (auth.PAT, error) {
	q := `
		UPDATE pats 
		SET description = :description, updated_at = :updated_at
		WHERE user_id = :user_id AND id = :id
		RETURNING id, user_id, name, description, secret, issued_at, updated_at, expires_at, revoked, revoked_at, last_used_at`

	upm := dbPagemeta{
		User:        userID,
		ID:          patID,
		UpdatedAt:   time.Now(),
		Description: description,
	}

	rows, err := pr.db.NamedQueryContext(ctx, q, upm)
	if err != nil {
		return auth.PAT{}, errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	var pat auth.PAT
	if err := rows.StructScan(&pat); err != nil {
		return auth.PAT{}, errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	return pat, nil
}

func (pr *patRepo) UpdateTokenHash(ctx context.Context, userID, patID, tokenHash string, expiryAt time.Time) (auth.PAT, error) {
	q := `
		UPDATE pats 
		SET secret = :secret, expires_at = :expires_at, updated_at = :updated_at
		WHERE user_id = :user_id AND id = :id
		RETURNING id, user_id, name, description, secret, issued_at, updated_at, expires_at, revoked, revoked_at, last_used_at`

	upm := dbPagemeta{
		User:      userID,
		ID:        patID,
		UpdatedAt: time.Now(),
		ExpiresAt: expiryAt,
		Secret:    tokenHash,
	}

	rows, err := pr.db.NamedQueryContext(ctx, q, upm)
	if err != nil {
		return auth.PAT{}, errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	var pat auth.PAT
	if err := rows.StructScan(&pat); err != nil {
		return auth.PAT{}, errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	return pat, nil
}

func (pr *patRepo) Revoke(ctx context.Context, userID, patID string) error {
	q := `
		UPDATE pats 
		SET revoked = true, revoked_at = :revoked_at
		WHERE user_id = :user_id AND id = :id`

	upm := dbPagemeta{
		User:      userID,
		ID:        patID,
		ExpiresAt: time.Now(),
	}

	_, err := pr.db.NamedQueryContext(ctx, q, upm)
	if err != nil {
		return errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	return nil
}

func (pr *patRepo) Reactivate(ctx context.Context, userID, patID string) error {
	q := `
		UPDATE pats 
		SET revoked = false, revoked_at = NULL
		WHERE user_id = :user_id AND id = :id`

	upm := dbPagemeta{
		User: userID,
		ID:   patID,
	}

	_, err := pr.db.NamedQueryContext(ctx, q, upm)
	if err != nil {
		return errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	return nil
}

func (pr *patRepo) Remove(ctx context.Context, userID, patID string) error {
	q := `DELETE FROM pats WHERE user_id = :user_id AND id = :id`
	upm := dbPagemeta{
		User: userID,
		ID:   patID,
	}

	_, err := pr.db.NamedQueryContext(ctx, q, upm)
	if err != nil {
		return postgres.HandleError(repoerr.ErrRemoveEntity, err)
	}

	return nil
}

func (pr *patRepo) AddScopeEntry(ctx context.Context, userID string, scopes []auth.Scope) error {
	dbscopes := toDBScope(scopes)

	_, err := pr.db.NamedQueryContext(ctx, saveScopeQuery, dbscopes)
	if err != nil {
		return postgres.HandleError(repoerr.ErrUpdateEntity, err)
	}

	if err := pr.cache.Save(ctx, userID, scopes); err != nil {
		return errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	return nil
}

func (pr *patRepo) RemoveScopeEntry(ctx context.Context, userID string, scopesID ...string) error {
	params := dbPagemeta{
		ScopesID: scopesID,
	}
	_, err := pr.db.NamedQueryContext(ctx, deleteScopesQuery, params)
	if err != nil {
		return errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	if err := pr.cache.Remove(ctx, userID, scopesID...); err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}

	return nil
}

func (pr *patRepo) CheckScopeEntry(ctx context.Context, userID, patID string, entityType auth.EntityType, optionalDomainID string, operation auth.Operation, entityID string) error {
	authorized := pr.cache.CheckScope(ctx, userID, patID, optionalDomainID, entityType, operation, entityID)
	if authorized {
		return nil
	}

	scope := dbScope{
		PatID:            patID,
		EntityType:       entityType.String(),
		OptionalDomainID: optionalDomainID,
		Operation:        operation.String(),
		EntityID:         entityID,
	}

	rows, err := pr.db.NamedQueryContext(ctx, checkScopeQuery, scope)
	if err != nil {
		return errors.Wrap(repoerr.ErrViewEntity, err)
	}
	defer rows.Close()

	if rows.Next() {
		var sc dbScope
		if err := rows.StructScan(&sc); err != nil {
			return errors.Wrap(repoerr.ErrViewEntity, err)
		}

		entityType, err := auth.ParseEntityType(sc.EntityType)
		if err != nil {
			return errors.Wrap(repoerr.ErrViewEntity, err)
		}
		operation, err := auth.ParseOperation(sc.Operation)
		if err != nil {
			return errors.Wrap(repoerr.ErrViewEntity, err)
		}
		authScope := auth.Scope{
			ID:               sc.ID,
			PatID:            sc.PatID,
			OptionalDomainID: sc.OptionalDomainID,
			EntityType:       entityType,
			EntityID:         sc.EntityID,
			Operation:        operation,
		}

		if err := pr.cache.Save(ctx, userID, []auth.Scope{authScope}); err != nil {
			return err
		}

		if authScope.Authorized(entityType, optionalDomainID, operation, entityID) {
			return nil
		}
	}

	return repoerr.ErrNotFound
}

func (pr *patRepo) RemoveAllScopeEntry(ctx context.Context, patID string) error {
	pm := dbPagemeta{
		PatID: patID,
	}

	_, err := pr.db.NamedQueryContext(ctx, deleteAllScopesQuery, pm)
	if err != nil {
		return postgres.HandleError(repoerr.ErrUpdateEntity, err)
	}

	if err := pr.cache.RemoveAllScope(ctx, pm.User, patID); err != nil {
		return errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	return nil
}

func (pr *patRepo) RetrieveScope(ctx context.Context, pm auth.ScopesPageMeta) (auth.ScopesPage, error) {
	dbs := dbPagemeta{
		PatID:  pm.PatID,
		Offset: pm.Offset,
		Limit:  pm.Limit,
		ID:     pm.ID,
	}

	scopes, err := pr.retrieveScopeFromDB(ctx, dbs)
	if err != nil {
		return auth.ScopesPage{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}

	cq := `SELECT COUNT(*) FROM pat_scopes WHERE pat_id = :pat_id`

	total, err := postgres.Total(ctx, pr.db, cq, dbs)
	if err != nil {
		return auth.ScopesPage{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}

	return auth.ScopesPage{
		Total:  total,
		Scopes: scopes,
		Offset: pm.Offset,
		Limit:  pm.Limit,
	}, nil
}

func (pr *patRepo) retrieveScopeFromDB(ctx context.Context, pm dbPagemeta) ([]auth.Scope, error) {
	scopeRows, err := pr.db.NamedQueryContext(ctx, retrieveScopesQuery, pm)
	if err != nil {
		return []auth.Scope{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}
	defer scopeRows.Close()

	var scopes []dbScope
	for scopeRows.Next() {
		var scope dbScope
		if err := scopeRows.StructScan(&scope); err != nil {
			return []auth.Scope{}, errors.Wrap(repoerr.ErrViewEntity, err)
		}
		scopes = append(scopes, scope)
	}

	sc, err := toAuthScope(scopes)
	if err != nil {
		return []auth.Scope{}, err
	}

	return sc, nil
}

func (pr *patRepo) retrievePATFromDB(ctx context.Context, userID, patID string) (auth.PAT, error) {
	dbp := dbPagemeta{
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
	}

	pat, err := toAuthPat(record)
	if err != nil {
		return auth.PAT{}, err
	}

	return pat, nil
}
