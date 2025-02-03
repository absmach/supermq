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

	// Scopes data stored as JSON
	ScopesData string `db:"scopes_data,omitempty"`

	// Aggregated scope fields for querying
	AllowedOps  []string `db:"allowed_ops,omitempty"`  // Combined list of all allowed operations
	EntityIDs   []string `db:"entity_ids,omitempty"`   // Combined list of all entity IDs
	Domains     []string `db:"domains,omitempty"`      // List of all domain IDs
	EntityTypes []string `db:"entity_types,omitempty"` // List of all entity types

	// Metadata
	Metadata string `db:"metadata,omitempty"` // JSON string for additional metadata
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

	// Parse scopes data
	var scopeData struct {
		Users        auth.OperationScope        `json:"users,omitempty"`
		Dashboard    auth.OperationScope        `json:"dashboard,omitempty"`
		Messaging    auth.OperationScope        `json:"messaging,omitempty"`
		DomainScopes map[string]DomainScopeData `json:"domain_scopes,omitempty"`
	}

	if err := json.Unmarshal([]byte(db.ScopesData), &scopeData); err != nil {
		return auth.PAT{}, fmt.Errorf("failed to unmarshal scopes data: %w", err)
	}

	// Parse metadata if exists
	var metadata map[string]interface{}
	if db.Metadata != "" {
		if err := json.Unmarshal([]byte(db.Metadata), &metadata); err != nil {
			return auth.PAT{}, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	// Set platform scopes
	pat.Scope.Users = scopeData.Users
	pat.Scope.Dashboard = scopeData.Dashboard
	pat.Scope.Messaging = scopeData.Messaging

	// Set domain scopes
	for domainID, dsd := range scopeData.DomainScopes {
		domainScope := auth.DomainScope{
			DomainManagement: dsd.DomainManagement,
			Entities:         make(map[auth.DomainEntityType]auth.OperationScope),
		}

		// Process domain entity scopes
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
	// Initialize scope data structure
	scopeData := struct {
		Users        auth.OperationScope        `json:"users,omitempty"`
		Dashboard    auth.OperationScope        `json:"dashboard,omitempty"`
		Messaging    auth.OperationScope        `json:"messaging,omitempty"`
		DomainScopes map[string]DomainScopeData `json:"domain_scopes,omitempty"`
	}{
		DomainScopes: make(map[string]DomainScopeData),
	}

	// Collect all operations and entity IDs
	var allOps []string
	var allEntityIDs []string
	var allDomains []string
	var allEntityTypes []string

	// Process platform scopes
	if len(pat.Scope.Users) > 0 {
		scopeData.Users = pat.Scope.Users
		ops, ids := extractOpsAndIDs(pat.Scope.Users)
		allOps = append(allOps, ops...)
		allEntityIDs = append(allEntityIDs, ids...)
		allEntityTypes = append(allEntityTypes, "users")
	}

	if len(pat.Scope.Dashboard) > 0 {
		scopeData.Dashboard = pat.Scope.Dashboard
		ops, ids := extractOpsAndIDs(pat.Scope.Dashboard)
		allOps = append(allOps, ops...)
		allEntityIDs = append(allEntityIDs, ids...)
		allEntityTypes = append(allEntityTypes, "dashboard")
	}

	if len(pat.Scope.Messaging) > 0 {
		scopeData.Messaging = pat.Scope.Messaging
		ops, ids := extractOpsAndIDs(pat.Scope.Messaging)
		allOps = append(allOps, ops...)
		allEntityIDs = append(allEntityIDs, ids...)
		allEntityTypes = append(allEntityTypes, "messaging")
	}

	// Process domain scopes
	for domainID, domainScope := range pat.Scope.Domains {
		dsd := DomainScopeData{
			DomainManagement: domainScope.DomainManagement,
			Entities:         make(map[string]auth.OperationScope),
		}

		if len(domainScope.DomainManagement) > 0 {
			ops, ids := extractOpsAndIDs(domainScope.DomainManagement)
			allOps = append(allOps, ops...)
			allEntityIDs = append(allEntityIDs, ids...)
			allEntityTypes = append(allEntityTypes, "domain_management")
		}

		for entityType, ops := range domainScope.Entities {
			entityTypeStr, err := entityType.ValidString()
			if err != nil {
				return dbPat{}, fmt.Errorf("invalid entity type: %w", err)
			}
			dsd.Entities[entityTypeStr] = ops

			extractedOps, ids := extractOpsAndIDs(ops)
			allOps = append(allOps, extractedOps...)
			allEntityIDs = append(allEntityIDs, ids...)
			allEntityTypes = append(allEntityTypes, entityTypeStr)
		}

		scopeData.DomainScopes[domainID] = dsd
		allDomains = append(allDomains, domainID)
	}

	// Remove duplicates
	allOps = uniqueStrings(allOps)
	allEntityIDs = uniqueStrings(allEntityIDs)
	allDomains = uniqueStrings(allDomains)
	allEntityTypes = uniqueStrings(allEntityTypes)

	// Marshal scope data
	scopesJSON, err := json.Marshal(scopeData)
	if err != nil {
		return dbPat{}, fmt.Errorf("failed to marshal scopes data: %w", err)
	}

	// Create metadata
	metadata := map[string]interface{}{
		"created_at": time.Now(),
		"version":    "1.0",
	}
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return dbPat{}, fmt.Errorf("failed to marshal metadata: %w", err)
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
		AllowedOps:  allOps,
		EntityIDs:   allEntityIDs,
		Domains:     allDomains,
		EntityTypes: allEntityTypes,
		Metadata:    string(metadataJSON),
	}, nil
}

type DomainScopeData struct {
	DomainManagement auth.OperationScope            `json:"domain_management,omitempty"`
	Entities         map[string]auth.OperationScope `json:"entities,omitempty"`
}

func extractOpsAndIDs(ops auth.OperationScope) ([]string, []string) {
	var operations []string
	var entityIDs []string

	for op, scopeValue := range ops {
		opStr, err := op.ValidString()
		if err != nil {
			continue
		}
		operations = append(operations, opStr)
		entityIDs = append(entityIDs, scopeValue.Values()...)
	}

	return operations, entityIDs
}

func uniqueStrings(strs []string) []string {
	seen := make(map[string]struct{})
	var result []string
	for _, str := range strs {
		if _, exists := seen[str]; !exists {
			seen[str] = struct{}{}
			result = append(result, str)
		}
	}
	return result
}

func toDBAuthPage(user string, pm auth.PATSPageMeta) dbAuthPage {
	return dbAuthPage{
		Limit:  pm.Limit,
		Offset: pm.Offset,
		User:   user,
	}
}
