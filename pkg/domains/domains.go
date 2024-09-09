// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package domains

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/absmach/magistrala/pkg/clients"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	"github.com/absmach/magistrala/pkg/roles"
)

// Status represents Domain status.
type Status uint8

// Possible Domain status values.
const (
	// EnabledStatus represents enabled Domain.
	EnabledStatus Status = iota
	// DisabledStatus represents disabled Domain.
	DisabledStatus
	// FreezeStatus represents domain is in freezed state.
	FreezeStatus

	// AllStatus is used for querying purposes to list Domains irrespective
	// of their status - enabled, disabled, freezed, deleting. It is never stored in the
	// database as the actual domain status and should always be the larger than freeze status
	// value in this enumeration.
	AllStatus
)

// String representation of the possible status values.
const (
	Disabled = "disabled"
	Enabled  = "enabled"
	Freezed  = "freezed"
	All      = "all"
	Unknown  = "unknown"
)

// String converts client/group status to string literal.
func (s Status) String() string {
	switch s {
	case DisabledStatus:
		return Disabled
	case EnabledStatus:
		return Enabled
	case AllStatus:
		return All
	case FreezeStatus:
		return Freezed
	default:
		return Unknown
	}
}

// ToStatus converts string value to a valid Domain status.
func ToStatus(status string) (Status, error) {
	switch status {
	case "", Enabled:
		return EnabledStatus, nil
	case Disabled:
		return DisabledStatus, nil
	case Freezed:
		return FreezeStatus, nil
	case All:
		return AllStatus, nil
	}
	return Status(0), svcerr.ErrInvalidStatus
}

// Custom Marshaller for Domains status.
func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// Custom Unmarshaler for Domains status.
func (s *Status) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), "\"")
	val, err := ToStatus(str)
	*s = val
	return err
}

const (
	Update                 roles.Operation = "update"
	Read                   roles.Operation = "read"
	Delete                 roles.Operation = "delete"
	Membership             roles.Operation = "membership"
	ManageRole             roles.Operation = "manage_role"
	AddRoleUsers           roles.Operation = "add_role_users"
	RemoveRoleUsers        roles.Operation = "remove_role_users"
	ViewRoleUsers          roles.Operation = "view_role_users"
	ThingCreate            roles.Operation = "thing_create"
	ThingUpdate            roles.Operation = "thing_update"
	ThingRead              roles.Operation = "thing_read"
	ThingDelete            roles.Operation = "thing_delete"
	ThingSetParentGroup    roles.Operation = "thing_set_parent_group"
	ThingConnectToChannel  roles.Operation = "thing_connect_to_channel"
	ThingManageRole        roles.Operation = "thing_manage_role"
	ThingAddRoleUsers      roles.Operation = "thing_add_role_users"
	ThingRemoveRoleUsers   roles.Operation = "thing_remove_role_users"
	ThingViewRoleUsers     roles.Operation = "thing_view_role_users"
	ChannelCreate          roles.Operation = "channel_create"
	ChannelUpdate          roles.Operation = "channel_update"
	ChannelRead            roles.Operation = "channel_read"
	ChannelDelete          roles.Operation = "channel_delete"
	ChannelSetParentGroup  roles.Operation = "channel_set_parent_group"
	ChannelConnectToThing  roles.Operation = "channel_connect_to_thing"
	ChannelPublish         roles.Operation = "channel_publish"
	ChannelSubscribe       roles.Operation = "channel_subscribe"
	ChannelManageRole      roles.Operation = "channel_manage_role"
	ChannelAddRoleUsers    roles.Operation = "channel_add_role_users"
	ChannelRemoveRoleUsers roles.Operation = "channel_remove_role_users"
	ChannelViewRoleUsers   roles.Operation = "channel_view_role_users"
	GroupCreate            roles.Operation = "group_create"
	GroupUpdate            roles.Operation = "group_update"
	GroupRead              roles.Operation = "group_read"
	GroupDelete            roles.Operation = "group_delete"
	GroupSetChild          roles.Operation = "group_set_child"
	GroupSetParent         roles.Operation = "group_set_parent"
	GroupManageRole        roles.Operation = "group_manage_role"
	GroupAddRoleUsers      roles.Operation = "group_add_role_users"
	GroupRemoveRoleUsers   roles.Operation = "group_remove_role_users"
	GroupViewRoleUsers     roles.Operation = "group_view_role_users"
)

const (
	BuiltInRoleAdmin      = "admin"
	BuiltInRoleMembership = "membership"
)

func AllowedOperations() []roles.Operation {
	return []roles.Operation{
		Update,
		Read,
		Delete,
		Membership,
		ManageRole,
		AddRoleUsers,
		RemoveRoleUsers,
		ViewRoleUsers,
		ThingCreate,
		ThingUpdate,
		ThingRead,
		ThingDelete,
		ThingSetParentGroup,
		ThingConnectToChannel,
		ThingManageRole,
		ThingAddRoleUsers,
		ThingRemoveRoleUsers,
		ThingViewRoleUsers,
		ChannelCreate,
		ChannelUpdate,
		ChannelRead,
		ChannelDelete,
		ChannelSetParentGroup,
		ChannelConnectToThing,
		ChannelPublish,
		ChannelSubscribe,
		ChannelManageRole,
		ChannelAddRoleUsers,
		ChannelRemoveRoleUsers,
		ChannelViewRoleUsers,
		GroupCreate,
		GroupUpdate,
		GroupRead,
		GroupDelete,
		GroupSetChild,
		GroupSetParent,
		GroupManageRole,
		GroupAddRoleUsers,
		GroupRemoveRoleUsers,
		GroupViewRoleUsers,
	}
}

func membershipRoleOperations() []roles.Operation {
	return []roles.Operation{
		Membership,
	}
}

func BuiltInRoles() map[roles.BuiltInRoleName][]roles.Operation {
	return map[roles.BuiltInRoleName][]roles.Operation{
		"admin":      AllowedOperations(),
		"membership": membershipRoleOperations(),
	}
}

const (
	UpdatePermission          = "update_permission"
	ReadPermission            = "read_permission"
	MembershipPermission      = "membership_permission"
	DeletePermission          = "delete_permission"
	ManageRolePermission      = "manage_role_permission"
	AddRoleUsersPermission    = "add_role_users_permission"
	RemoveRoleUsersPermission = "remove_role_users_permission"
	ViewRoleUsersPermission   = "view_role_users_permission"
)

type DomainReq struct {
	Name     *string           `json:"name,omitempty"`
	Metadata *clients.Metadata `json:"metadata,omitempty"`
	Tags     *[]string         `json:"tags,omitempty"`
	Alias    *string           `json:"alias,omitempty"`
	Status   *Status           `json:"status,omitempty"`
}
type Domain struct {
	ID         string           `json:"id"`
	Name       string           `json:"name"`
	Metadata   clients.Metadata `json:"metadata,omitempty"`
	Tags       []string         `json:"tags,omitempty"`
	Alias      string           `json:"alias,omitempty"`
	Status     Status           `json:"status"`
	Permission string           `json:"permission,omitempty"`
	CreatedBy  string           `json:"created_by,omitempty"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedBy  string           `json:"updated_by,omitempty"`
	UpdatedAt  time.Time        `json:"updated_at,omitempty"`
}

type Page struct {
	Total      uint64           `json:"total"`
	Offset     uint64           `json:"offset"`
	Limit      uint64           `json:"limit"`
	Name       string           `json:"name,omitempty"`
	Order      string           `json:"-"`
	Dir        string           `json:"-"`
	Metadata   clients.Metadata `json:"metadata,omitempty"`
	Tag        string           `json:"tag,omitempty"`
	Permission string           `json:"permission,omitempty"`
	Status     Status           `json:"status,omitempty"`
	ID         string           `json:"id,omitempty"`
	IDs        []string         `json:"-"`
	Identity   string           `json:"identity,omitempty"`
	SubjectID  string           `json:"-"`
}

type DomainsPage struct {
	Total   uint64   `json:"total"`
	Offset  uint64   `json:"offset"`
	Limit   uint64   `json:"limit"`
	Domains []Domain `json:"domains"`
}

func (page DomainsPage) MarshalJSON() ([]byte, error) {
	type Alias DomainsPage
	a := struct {
		Alias
	}{
		Alias: Alias(page),
	}

	if a.Domains == nil {
		a.Domains = make([]Domain, 0)
	}

	return json.Marshal(a)
}

type Policy struct {
	SubjectType     string `json:"subject_type,omitempty"`
	SubjectID       string `json:"subject_id,omitempty"`
	SubjectRelation string `json:"subject_relation,omitempty"`
	Relation        string `json:"relation,omitempty"`
	ObjectType      string `json:"object_type,omitempty"`
	ObjectID        string `json:"object_id,omitempty"`
}

//go:generate mockery --name Service --output=./mocks --filename service.go --quiet --note "Copyright (c) Abstract Machines"
type Service interface {
	CreateDomain(ctx context.Context, token string, d Domain) (Domain, error)
	RetrieveDomain(ctx context.Context, token string, id string) (Domain, error)
	UpdateDomain(ctx context.Context, token string, id string, d DomainReq) (Domain, error)
	ChangeDomainStatus(ctx context.Context, token string, id string, d DomainReq) (Domain, error)
	ListDomains(ctx context.Context, token string, page Page) (DomainsPage, error)
	roles.Roles
}

// DomainsRepository specifies Domain persistence API.
//
//go:generate mockery --name DomainsRepository --output=./mocks --filename repository.go  --quiet --note "Copyright (c) Abstract Machines"
type DomainsRepository interface {
	// Save creates db insert transaction for the given domain.
	Save(ctx context.Context, d Domain) (Domain, error)

	// RetrieveByID retrieves Domain by its unique ID.
	RetrieveByID(ctx context.Context, id string) (Domain, error)

	// RetrieveAllByIDs retrieves for given Domain IDs.
	RetrieveAllByIDs(ctx context.Context, pm Page) (DomainsPage, error)

	// Update updates the client name and metadata.
	Update(ctx context.Context, id string, userID string, d DomainReq) (Domain, error)

	// Delete
	Delete(ctx context.Context, id string) error

	// ListDomains list all the domains
	ListDomains(ctx context.Context, pm Page) (DomainsPage, error)

	roles.Repository
}
