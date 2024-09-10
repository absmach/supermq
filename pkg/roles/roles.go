// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package roles

import (
	"context"
	"time"
)

type Capability string

func (op Capability) String() string {
	return string(op)
}

type Member string

func (mem Member) String() string {
	return string(mem)
}

type RoleName string

func (r RoleName) String() string {
	return string(r)
}

type BuiltInRoleName RoleName

func (b BuiltInRoleName) ToRoleName() RoleName {
	return RoleName(b)
}

func (b BuiltInRoleName) String() string {
	return string(b)
}

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
	OptionalCapabilities []string `json:"-"`
	OptionalMembers      []string `json:"-"`
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
	SubjectKind     string
	Object          string
	ObjectType      string
	ObjectKind      string
	Relation        string
	Permission      string
}

//go:generate mockery --name Roles --output=./mocks --filename roles.go --quiet --note "Copyright (c) Abstract Machines"
type Roles interface {

	// Add New role to entity
	AddRole(ctx context.Context, token, entityID, roleName string, optionalCapabilities []string, optionalMembers []string) (Role, error)

	// Remove removes the roles of entity.
	RemoveRole(ctx context.Context, token, entityID, roleName string) error

	// UpdateName update the name of the entity role.
	UpdateRoleName(ctx context.Context, token, entityID, oldRoleName, newRoleName string) (Role, error)

	RetrieveRole(ctx context.Context, token, entityID, roleName string) (Role, error)

	RetrieveAllRoles(ctx context.Context, token, entityID string, limit, offset uint64) (RolePage, error)

	RoleAddCapabilities(ctx context.Context, token, entityID, roleName string, capabilities []string) (ops []string, err error)

	RoleListCapabilities(ctx context.Context, token, entityID, roleName string) ([]string, error)

	RoleCheckCapabilitiesExists(ctx context.Context, token, entityID, roleName string, capabilities []string) (bool, error)

	RoleRemoveCapabilities(ctx context.Context, token, entityID, roleName string, capabilities []string) (err error)

	RoleRemoveAllCapabilities(ctx context.Context, token, entityID, roleName string) error

	RoleAddMembers(ctx context.Context, token, entityID, roleName string, members []string) ([]string, error)

	RoleListMembers(ctx context.Context, token, entityID, roleName string, limit, offset uint64) (MembersPage, error)

	RoleCheckMembersExists(ctx context.Context, token, entityID, roleName string, members []string) (bool, error)

	RoleRemoveMembers(ctx context.Context, token, entityID, roleName string, members []string) (err error)

	RoleRemoveAllMembers(ctx context.Context, token, entityID, roleName string) (err error)
}

//go:generate mockery --name Repository --output=./mocks --filename rolesRepo.go --quiet --note "Copyright (c) Abstract Machines"
type Repository interface {
	AddRoles(ctx context.Context, rps []RoleProvision) ([]Role, error)
	RemoveRoles(ctx context.Context, roleIDs []string) error
	UpdateRole(ctx context.Context, ro Role) (Role, error)
	RetrieveRole(ctx context.Context, roleID string) (Role, error)
	RetrieveRoleByEntityIDAndName(ctx context.Context, entityID, roleName string) (Role, error)
	RetrieveAllRoles(ctx context.Context, entityID string, limit, offset uint64) (RolePage, error)
	RoleAddCapabilities(ctx context.Context, role Role, capabilities []string) (ops []string, err error)
	RoleListCapabilities(ctx context.Context, roleID string) ([]string, error)
	RoleCheckCapabilitiesExists(ctx context.Context, roleID string, capabilities []string) (bool, error)
	RoleRemoveCapabilities(ctx context.Context, role Role, capabilities []string) (err error)
	RoleRemoveAllCapabilities(ctx context.Context, role Role) error
	RoleAddMembers(ctx context.Context, role Role, members []string) ([]string, error)
	RoleListMembers(ctx context.Context, roleID string, limit, offset uint64) (MembersPage, error)
	RoleCheckMembersExists(ctx context.Context, roleID string, members []string) (bool, error)
	RoleRemoveMembers(ctx context.Context, role Role, members []string) (err error)
	RoleRemoveAllMembers(ctx context.Context, role Role) (err error)
}
