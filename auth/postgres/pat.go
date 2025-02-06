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
	PatID         string   `db:"pat_id,omitempty"`
	Platformtype  string   `db:"platform_type,omitempty"`
	DomainID      string   `db:"domain_id,omitempty"`
	DomainType    string   `db:"domain_type,omitempty"`
	OperationType string   `db:"operation_type,omitempty"`
	EntityIDs     []string `db:"entity_ids,omitempty"`
}

type dbAuthPage struct {
	Limit  uint64 `db:"limit"`
	Offset uint64 `db:"offset"`
	User   string `db:"user_id"`
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
		Scope:       auth.Scope{Domains: make(map[string]auth.DomainScope)},
	}
	scope, err := toAuthScope(sc)
	if err != nil {
		return auth.PAT{}, err
	}
	pat.Scope = scope

	return pat, nil
}

func toAuthScope(sc []dbScope) (auth.Scope, error) {
	scope := auth.Scope{
		Domains:   make(map[string]auth.DomainScope),
		Users:     auth.OperationScope{},
		Dashboard: auth.OperationScope{},
		Messaging: auth.OperationScope{},
	}

	for _, t := range sc {
		var platformType auth.PlatformEntityType
		var operation auth.OperationType
		var err error

		if t.Platformtype == "" {
			return scope, nil
		}

		platformType, err = auth.ParsePlatformEntityType(t.Platformtype)
		if err != nil {
			return auth.Scope{}, err
		}

		operation, err = auth.ParseOperationType(t.OperationType)
		if err != nil {
			return auth.Scope{}, err
		}

		switch platformType {
		case auth.PlatformUsersScope:
			if err := scope.Users.Add(operation, t.EntityIDs...); err != nil {
				return auth.Scope{}, err
			}
		case auth.PlatformDashBoardScope:
			if err := scope.Dashboard.Add(operation, t.EntityIDs...); err != nil {
				return auth.Scope{}, err
			}
		case auth.PlatformMesagingScope:
			if err := scope.Messaging.Add(operation, t.EntityIDs...); err != nil {
				return auth.Scope{}, err
			}
		case auth.PlatformDomainsScope:
			var domainEntityType auth.DomainEntityType
			if t.DomainType != "" {
				domainEntityType, err = auth.ParseDomainEntityType(t.DomainType)
				if err != nil {
					return auth.Scope{}, err
				}
			}

			if err := scope.Add(platformType, t.DomainID, domainEntityType, operation, t.EntityIDs...); err != nil {
				return auth.Scope{}, err
			}
		}
	}

	return scope, nil
}

func fromAuthScope(patID string, scope auth.Scope) []dbScope {
	var dbScopes []dbScope

	if isEmptyScope(scope) {
		sc := dbScope{
			PatID: patID,
		}
		dbScopes = append(dbScopes, sc)
		return dbScopes
	}

	for op, ids := range scope.Users {
		dbScopes = append(dbScopes, dbScope{
			PatID:         patID,
			Platformtype:  auth.PlatformUsersScope.String(),
			OperationType: op.String(),
			EntityIDs:     ids.Values(),
		})
	}

	for op, ids := range scope.Dashboard {
		dbScopes = append(dbScopes, dbScope{
			PatID:         patID,
			Platformtype:  auth.PlatformDashBoardScope.String(),
			OperationType: op.String(),
			EntityIDs:     ids.Values(),
		})
	}

	for op, ids := range scope.Messaging {
		dbScopes = append(dbScopes, dbScope{
			PatID:         patID,
			Platformtype:  auth.PlatformMesagingScope.String(),
			OperationType: op.String(),
			EntityIDs:     ids.Values(),
		})
	}

	for domainID, domainScope := range scope.Domains {
		for op, ids := range domainScope.DomainManagement {
			dbScopes = append(dbScopes, dbScope{
				PatID:         patID,
				Platformtype:  auth.PlatformDomainsScope.String(),
				DomainID:      domainID,
				DomainType:    auth.DomainManagementScope.String(),
				OperationType: op.String(),
				EntityIDs:     ids.Values(),
			})
		}

		for entityType, entityScope := range domainScope.Entities {
			for op, ids := range entityScope {
				dbScopes = append(dbScopes, dbScope{
					PatID:         patID,
					Platformtype:  auth.PlatformDomainsScope.String(),
					DomainID:      domainID,
					DomainType:    entityType.String(),
					OperationType: op.String(),
					EntityIDs:     ids.Values(),
				})
			}
		}
	}

	return dbScopes
}

func patToDBRecords(pat auth.PAT) (dbPat, []dbScope, error) {
	scopes := fromAuthScope(pat.ID, pat.Scope)
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

func toDBAuthPage(user string, pm auth.PATSPageMeta) dbAuthPage {
	return dbAuthPage{
		Limit:  pm.Limit,
		Offset: pm.Offset,
		User:   user,
	}
}
