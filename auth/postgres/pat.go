// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"encoding/json"
	"fmt"
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
	ScopesData  string    `db:"scopes_data,omitempty"`
}

type dbAuthPage struct {
	Limit  uint64 `db:"limit"`
	Offset uint64 `db:"offset"`
	User   string `db:"user_id"`
}

func toAuthPat(db dbPat) (auth.PAT, error) {
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
		Scope:       auth.Scope{Domains: make(map[string]auth.DomainScope)},
	}

	var scopeData struct {
		Users        auth.OperationScope        `json:"users,omitempty"`
		Dashboard    auth.OperationScope        `json:"dashboard,omitempty"`
		Messaging    auth.OperationScope        `json:"messaging,omitempty"`
		DomainScopes map[string]DomainScopeData `json:"domain_scopes,omitempty"`
	}

	if err := json.Unmarshal([]byte(db.ScopesData), &scopeData); err != nil {
		return auth.PAT{}, fmt.Errorf("failed to unmarshal scopes data: %w", err)
	}

	pat.Scope.Users = scopeData.Users
	pat.Scope.Dashboard = scopeData.Dashboard
	pat.Scope.Messaging = scopeData.Messaging

	for domainID, dsd := range scopeData.DomainScopes {
		domainScope := auth.DomainScope{
			DomainManagement: dsd.DomainManagement,
			Entities:         make(map[auth.DomainEntityType]auth.OperationScope),
		}

		for entityTypeStr, ops := range dsd.Entities {
			entityType, err := auth.ParseDomainEntityType(entityTypeStr)
			if err != nil {
				return auth.PAT{}, fmt.Errorf("invalid domain entity type %s: %w", entityTypeStr, err)
			}
			domainScope.Entities[entityType] = ops
		}

		pat.Scope.Domains[domainID] = domainScope
	}

	return pat, nil
}

func patToDBRecords(pat auth.PAT) (dbPat, error) {
	scopeData := struct {
		Users        auth.OperationScope        `json:"users,omitempty"`
		Dashboard    auth.OperationScope        `json:"dashboard,omitempty"`
		Messaging    auth.OperationScope        `json:"messaging,omitempty"`
		DomainScopes map[string]DomainScopeData `json:"domain_scopes,omitempty"`
	}{
		DomainScopes: make(map[string]DomainScopeData),
	}

	if len(pat.Scope.Users) > 0 {
		scopeData.Users = pat.Scope.Users
	}

	if len(pat.Scope.Dashboard) > 0 {
		scopeData.Dashboard = pat.Scope.Dashboard
	}

	if len(pat.Scope.Messaging) > 0 {
		scopeData.Messaging = pat.Scope.Messaging
	}

	for domainID, domainScope := range pat.Scope.Domains {
		dsd := DomainScopeData{
			DomainManagement: domainScope.DomainManagement,
			Entities:         make(map[string]auth.OperationScope),
		}

		for entityType, ops := range domainScope.Entities {
			entityTypeStr, err := entityType.ValidString()
			if err != nil {
				return dbPat{}, fmt.Errorf("invalid entity type: %w", err)
			}
			dsd.Entities[entityTypeStr] = ops
		}

		scopeData.DomainScopes[domainID] = dsd
	}

	scopesJSON, err := json.Marshal(scopeData)
	if err != nil {
		return dbPat{}, fmt.Errorf("failed to marshal scopes data: %w", err)
	}

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
		ScopesData:  string(scopesJSON),
	}, nil
}

type DomainScopeData struct {
	DomainManagement auth.OperationScope            `json:"domain_management,omitempty"`
	Entities         map[string]auth.OperationScope `json:"entities,omitempty"`
}

func toDBAuthPage(user string, pm auth.PATSPageMeta) dbAuthPage {
	return dbAuthPage{
		Limit:  pm.Limit,
		Offset: pm.Offset,
		User:   user,
	}
}
