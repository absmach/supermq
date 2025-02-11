// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"time"

	"github.com/absmach/supermq/auth"
)

type dbPat struct {
	ID          string    `db:"id,omitempty"`
	User        string    `db:"user_id,omitempty"`
	Name        string    `db:"name,omitempty"`
	Description string    `db:"description,omitempty"`
	Secret      string    `db:"secret,omitempty"`
	IssuedAt    time.Time `db:"issued_at,omitempty"`
	ExpiresAt   time.Time `db:"expires_at,omitempty"`
	UpdatedAt   time.Time `db:"updated_at,omitempty"`
	LastUsedAt  time.Time `db:"last_used_at,omitempty"`
	Revoked     bool      `db:"revoked,omitempty"`
	RevokedAt   time.Time `db:"revoked_at,omitempty"`
}

type dbScope struct {
	PatID            string `db:"pat_id,omitempty"`
	OptionalDomainId string `db:"optional_domain_id,omitempty"`
	EntityType       string `db:"entity_type,omitempty"`
	EntityID         string `db:"entity_id,omitempty"`
	Operation        string `db:"operation,omitempty"`
}

type dbPatPagemeta struct {
	Limit       uint64    `db:"limit"`
	Offset      uint64    `db:"offset"`
	User        string    `db:"user_id"`
	PatID       string    `db:"pat_id"`
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	UpdatedAt   time.Time `db:"updated_at"`
	ExpiresAt   time.Time `db:"expires_at"`
	Description string    `db:"description"`
	Secret      string    `db:"secret"`
}

func toAuthPat(db dbPat, sc []dbScope) (auth.PAT, error) {
	pat := auth.PAT{
		ID:          db.ID,
		User:        db.User,
		Name:        db.Name,
		Description: db.Description,
		Secret:      db.Secret,
		IssuedAt:    db.IssuedAt,
		ExpiresAt:   db.ExpiresAt,
		UpdatedAt:   db.UpdatedAt,
		LastUsedAt:  db.LastUsedAt,
		Revoked:     db.Revoked,
		RevokedAt:   db.RevokedAt,
	}
	scope, err := toAuthScope(sc)
	if err != nil {
		return auth.PAT{}, err
	}
	pat.Scope = scope

	return pat, nil
}

func toAuthScope(dsc []dbScope) ([]auth.Scope, error) {
	scope := []auth.Scope{}

	for _, s := range dsc {
		entityType, err := auth.ParseEntityType(s.EntityType)
		if err != nil {
			return []auth.Scope{}, err
		}
		operation, err := auth.ParseOperation(s.Operation)
		if err != nil {
			return []auth.Scope{}, err
		}
		scope = append(scope, auth.Scope{
			PatId:            s.PatID,
			OptionalDomainId: s.OptionalDomainId,
			EntityType:       entityType,
			EntityId:         s.EntityID,
			Operation:        operation,
		})
	}

	return scope, nil
}

func toDBPatScope(pat auth.PAT) []dbScope {
	var dbScopes []dbScope

	if isEmptyScope(pat.Scope) {
		sc := dbScope{
			PatID: pat.ID,
		}
		dbScopes = append(dbScopes, sc)
		return dbScopes
	}

	for _, p := range pat.Scope {
		dbScopes = append(dbScopes, dbScope{
			PatID:            pat.ID,
			OptionalDomainId: p.OptionalDomainId,
			EntityType:       p.EntityType.String(),
			Operation:        p.Operation.String(),
			EntityID:         p.EntityId,
		})
	}

	return dbScopes
}

func patToDBRecords(pat auth.PAT) (dbPat, []dbScope, error) {
	scopes := toDBPatScope(pat)
	return dbPat{
		ID:          pat.ID,
		User:        pat.User,
		Name:        pat.Name,
		Description: pat.Description,
		Secret:      pat.Secret,
		IssuedAt:    pat.IssuedAt,
		ExpiresAt:   pat.ExpiresAt,
		UpdatedAt:   pat.UpdatedAt,
		LastUsedAt:  pat.LastUsedAt,
		Revoked:     pat.Revoked,
		RevokedAt:   pat.RevokedAt,
	}, scopes, nil
}

func toDBAuthPage(user string, pm auth.PATSPageMeta) dbPatPagemeta {
	return dbPatPagemeta{
		Limit:  pm.Limit,
		Offset: pm.Offset,
		User:   user,
	}
}

func isEmptyScope(scope []auth.Scope) bool {
	return len(scope) == 0
}

func toDBScope(patID string, entityType auth.EntityType, optionalDomainID string, operation auth.Operation, entityIDs ...string) []dbScope {
	var scopes []dbScope
	for _, entityID := range entityIDs {
		scopes = append(scopes, dbScope{
			PatID:            patID,
			OptionalDomainId: optionalDomainID,
			EntityType:       entityType.String(),
			EntityID:         entityID,
			Operation:        operation.String(),
		})
	}
	return scopes
}
