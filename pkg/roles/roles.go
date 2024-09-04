// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package roles

import (
	"context"
	"time"
)

type Operation string

type Role struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	EntityID  string    `json:"entity_id"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedBy string    `json:"updated_by"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RoleProvision struct {
	Role
	OptionalOperations []Operation `json:"-"`
	OptionalMembers    []string    `json:"-"`
}

type RolePage struct {
	Total  uint64 `json:"total"`
	Offset uint64 `json:"offset"`
	Limit  uint64 `json:"limit"`
	Roles  []Role `json:"roles"`
}

type MembersPage struct {
	Total   uint64   `json:"total"`
	Offset  uint64   `json:"offset"`
	Limit   uint64   `json:"limit"`
	Members []string `json:"members"`
}

type OptionalPolicy struct {
	Namespace       string
	Subject         string
	SubjectType     string
	SubjectRelation string
	Object          string
	ObjectType      string
	Relation        string
	Permission      string
}

//go:generate mockery --name Roles --output=./mocks --filename roles.go --quiet --note "Copyright (c) Abstract Machines"
type Roles interface {

	// Get Allowed Operations

	// List inbuilt roles and its operations

	// AddNewEntityRoles adds roles to a newly creating entity.
	AddNewEntityRoles(ctx context.Context, entityID string, newEntityDefaultRoles map[string][]string, optionalPolicies []OptionalPolicy) ([]Role, error)

	// Add New role to entity
	Add(ctx context.Context, entityID, roleName string, optionalOperations []Operation, optionalMembers []string) (Role, error)

	// Remove removes the roles of entity.
	Remove(ctx context.Context, entityID, roleName string) error

	// UpdateName update the name of the entity role.
	UpdateName(ctx context.Context, entityID, oldRoleName, newRoleName string) (Role, error)

	Retrieve(ctx context.Context, entityID, roleName string) (Role, error)

	RetrieveAll(ctx context.Context, entityID string, limit, offset uint64) (RolePage, error)

	AddOperation(ctx context.Context, entityID, roleName string, operations []Operation) (ops []Operation, err error)

	ListOperations(ctx context.Context, entityID, roleName string) ([]Operation, error)

	CheckOperationsExists(ctx context.Context, entityID, roleName string, operations []Operation) (bool, error)

	RemoveOperations(ctx context.Context, entityID, roleName string, operations []Operation) (err error)

	RemoveAllOperations(ctx context.Context, entityID, roleName string) error

	AddMembers(ctx context.Context, entityID, roleName string, members []string) ([]string, error)

	ListMembers(ctx context.Context, entityID, roleName string, limit, offset uint64) (MembersPage, error)

	CheckMembersExists(ctx context.Context, entityID, roleName string, members []string) (bool, error)

	RemoveMembers(ctx context.Context, entityID, roleName string, members []string) (err error)

	RemoveAllMembers(ctx context.Context, entityID, roleName string) (err error)
}

//go:generate mockery --name Repository --output=./mocks --filename rolesRepo.go --quiet --note "Copyright (c) Abstract Machines"
type Repository interface {
	AddRoles(ctx context.Context, rps []RoleProvision) ([]Role, error)
	RemoveRoles(ctx context.Context, roleIDs []string) error
	UpdateRole(ctx context.Context, ro Role) (Role, error)
	RetrieveRole(ctx context.Context, roleID string) (Role, error)
	RetrieveRoleByEntityIDAndName(ctx context.Context, entityID, roleName string) (Role, error)
	RetrieveAllRoles(ctx context.Context, entityID string, limit, offset uint64) (RolePage, error)
	RoleAddOperation(ctx context.Context, role Role, operations []Operation) (ops []Operation, err error)
	RoleListOperations(ctx context.Context, roleID string) ([]Operation, error)
	RoleCheckOperationsExists(ctx context.Context, roleID string, operations []Operation) (bool, error)
	RoleRemoveOperations(ctx context.Context, role Role, operations []Operation) (err error)
	RoleRemoveAllOperations(ctx context.Context, role Role) error
	RoleAddMembers(ctx context.Context, role Role, members []string) ([]string, error)
	RoleListMembers(ctx context.Context, roleID string, limit, offset uint64) (MembersPage, error)
	RoleCheckMembersExists(ctx context.Context, roleID string, members []string) (bool, error)
	RoleRemoveMembers(ctx context.Context, role Role, members []string) (err error)
	RoleRemoveAllMembers(ctx context.Context, role Role) (err error)
}
