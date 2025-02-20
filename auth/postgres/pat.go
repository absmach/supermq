// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"time"

	"github.com/absmach/supermq/auth"
	repoerr "github.com/absmach/supermq/pkg/errors/repository"
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
	ID               string `db:"id,omitempty"`
	PatID            string `db:"pat_id,omitempty"`
	OptionalDomainID string `db:"optional_domain_id,omitempty"`
	EntityType       string `db:"entity_type,omitempty"`
	EntityID         string `db:"entity_id,omitempty"`
	Operation        string `db:"operation,omitempty"`
}

type dbPagemeta struct {
	Limit       uint64    `db:"limit"`
	Offset      uint64    `db:"offset"`
	User        string    `db:"user_id"`
	PatID       string    `db:"pat_id"`
	ScopesID    []string  `db:"scopes_id"`
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	UpdatedAt   time.Time `db:"updated_at"`
	ExpiresAt   time.Time `db:"expires_at"`
	Description string    `db:"description"`
	Secret      string    `db:"secret"`
}

func toAuthPat(db dbPat) (auth.PAT, error) {
	if db.ID == "" {
		return auth.PAT{}, repoerr.ErrNotFound
	}

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
			ID:               s.ID,
			PatID:            s.PatID,
			OptionalDomainID: s.OptionalDomainID,
			EntityType:       entityType,
			EntityID:         s.EntityID,
			Operation:        operation,
		})
	}

	return scope, nil
}

func toDBPats(pat auth.PAT) (dbPat, error) {
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
	}, nil
}

func toDBScope(sc []auth.Scope) []dbScope {
	var scopes []dbScope
	for _, s := range sc {
		scopes = append(scopes, dbScope{
			ID:               s.ID,
			PatID:            s.PatID,
			OptionalDomainID: s.OptionalDomainID,
			EntityType:       s.EntityType.String(),
			EntityID:         s.EntityID,
			Operation:        s.Operation.String(),
		})
	}
	return scopes
}
