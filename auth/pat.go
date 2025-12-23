// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	apiutil "github.com/absmach/supermq/api/http/util"
	"github.com/absmach/supermq/pkg/errors"
)

const AnyIDs = "*"

type Operation uint32

// Clients operations.
const (
	ClientCreateOp Operation = iota + 100
	ClientListOp
	ClientViewOp
	ClientUpdateOp
	ClientUpdateTagsOp
	ClientUpdateSecretOp
	ClientEnableOp
	ClientDisableOp
	ClientDeleteOp
	ClientSetParentGroupOp
	ClientRemoveParentGroupOp
	ClientConnectToChannelOp
	ClientDisconnectFromChannelOp
)

// Channels operations.
const (
	ChannelCreateOp Operation = iota + 200
	ChannelListOp
	ChannelViewOp
	ChannelUpdateOp
	ChannelUpdateTagsOp
	ChannelEnableOp
	ChannelDisableOp
	ChannelDeleteOp
	ChannelSetParentGroupOp
	ChannelRemoveParentGroupOp
	ChannelConnectToClientOp
	ChannelDisconnectFromClientOp
)

// Groups operations.
const (
	GroupCreateOp Operation = iota + 300
	GroupListOp
	GroupViewOp
	GroupUpdateOp
	GroupUpdateTagsOp
	GroupEnableOp
	GroupDisableOp
	GroupDeleteOp
	GroupRetrieveHierarchyOp
	GroupAddParentGroupOp
	GroupRemoveParentGroupOp
	GroupAddChildrenGroupsOp
	GroupRemoveChildrenGroupsOp
	GroupRemoveAllChildrenGroupsOp
	GroupListChildrenGroupsOp
	GroupSetChildClientOp
	GroupRemoveChildClientOp
	GroupSetChildChannelOp
	GroupRemoveChildChannelOp
)

// Dashboard operations.
const (
	DashboardShareOp Operation = iota + 400
	DashboardUnshareOp
)

// Messages operations.
const (
	MessagePublishOp Operation = iota + 500
	MessageSubscribeOp
)

// Role operations - common for clients, channels, and groups.
const (
	RoleAddOp Operation = iota + 600
	RoleRemoveOp
	RoleUpdateOp
	RoleRetrieveOp
	RoleRetrieveAllOp
	RoleAddActionsOp
	RoleListActionsOp
	RoleCheckActionsExistsOp
	RoleRemoveActionsOp
	RoleRemoveAllActionsOp
	RoleAddMembersOp
	RoleListMembersOp
	RoleCheckMembersExistsOp
	RoleRemoveMembersOp
	RoleRemoveAllMembersOp
)

// operationToString maps Operation values to their string representation.
var operationToString = map[Operation]string{
	// Client operations
	ClientCreateOp:                "client_create",
	ClientListOp:                  "client_list",
	ClientViewOp:                  "client_view",
	ClientUpdateOp:                "client_update",
	ClientUpdateTagsOp:            "client_update_tags",
	ClientUpdateSecretOp:          "client_update_secret",
	ClientEnableOp:                "client_enable",
	ClientDisableOp:               "client_disable",
	ClientDeleteOp:                "client_delete",
	ClientSetParentGroupOp:        "client_set_parent_group",
	ClientRemoveParentGroupOp:     "client_remove_parent_group",
	ClientConnectToChannelOp:      "client_connect_to_channel",
	ClientDisconnectFromChannelOp: "client_disconnect_from_channel",
	// Channel operations
	ChannelCreateOp:               "channel_create",
	ChannelListOp:                 "channel_list",
	ChannelViewOp:                 "channel_view",
	ChannelUpdateOp:               "channel_update",
	ChannelUpdateTagsOp:           "channel_update_tags",
	ChannelEnableOp:               "channel_enable",
	ChannelDisableOp:              "channel_disable",
	ChannelDeleteOp:               "channel_delete",
	ChannelSetParentGroupOp:       "channel_set_parent_group",
	ChannelRemoveParentGroupOp:    "channel_remove_parent_group",
	ChannelConnectToClientOp:      "channel_connect_to_client",
	ChannelDisconnectFromClientOp: "channel_disconnect_from_client",
	// Group operations
	GroupCreateOp:                  "group_create",
	GroupListOp:                    "group_list",
	GroupViewOp:                    "group_view",
	GroupUpdateOp:                  "group_update",
	GroupUpdateTagsOp:              "group_update_tags",
	GroupEnableOp:                  "group_enable",
	GroupDisableOp:                 "group_disable",
	GroupDeleteOp:                  "group_delete",
	GroupRetrieveHierarchyOp:       "group_retrieve_hierarchy",
	GroupAddParentGroupOp:          "group_add_parent_group",
	GroupRemoveParentGroupOp:       "group_remove_parent_group",
	GroupAddChildrenGroupsOp:       "group_add_children_groups",
	GroupRemoveChildrenGroupsOp:    "group_remove_children_groups",
	GroupRemoveAllChildrenGroupsOp: "group_remove_all_children_groups",
	GroupListChildrenGroupsOp:      "group_list_children_groups",
	GroupSetChildClientOp:          "group_set_child_client",
	GroupRemoveChildClientOp:       "group_remove_child_client",
	GroupSetChildChannelOp:         "group_set_child_channel",
	GroupRemoveChildChannelOp:      "group_remove_child_channel",
	// Dashboard operations
	DashboardShareOp:   "dashboard_share",
	DashboardUnshareOp: "dashboard_unshare",
	// Message operations
	MessagePublishOp:   "message_publish",
	MessageSubscribeOp: "message_subscribe",
	// Role operations - common for clients, channels, and groups
	RoleAddOp:                "role_add",
	RoleRemoveOp:             "role_remove",
	RoleUpdateOp:             "role_update",
	RoleRetrieveOp:           "role_retrieve",
	RoleRetrieveAllOp:        "role_retrieve_all",
	RoleAddActionsOp:         "role_add_actions",
	RoleListActionsOp:        "role_list_actions",
	RoleCheckActionsExistsOp: "role_check_actions_exists",
	RoleRemoveActionsOp:      "role_remove_actions",
	RoleRemoveAllActionsOp:   "role_remove_all_actions",
	RoleAddMembersOp:         "role_add_members",
	RoleListMembersOp:        "role_list_members",
	RoleCheckMembersExistsOp: "role_check_members_exists",
	RoleRemoveMembersOp:      "role_remove_members",
	RoleRemoveAllMembersOp:   "role_remove_all_members",
}

// stringToOperation is the reverse map, built from operationToString.
var stringToOperation = func() map[string]Operation {
	m := make(map[string]Operation)
	for op, str := range operationToString {
		m[str] = op
	}
	return m
}()

func (op Operation) String() string {
	if str, ok := operationToString[op]; ok {
		return str
	}
	return fmt.Sprintf("unknown operation type %d", op)
}

func (op Operation) ValidString() (string, error) {
	if str, ok := operationToString[op]; ok {
		return str, nil
	}
	return "", fmt.Errorf("unknown operation type %d", op)
}

func ParseOperation(op string) (Operation, error) {
	if operation, ok := stringToOperation[op]; ok {
		return operation, nil
	}
	return 0, fmt.Errorf("unknown operation type %s", op)
}

func (op Operation) MarshalJSON() ([]byte, error) {
	return json.Marshal(op.String())
}

func (op *Operation) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), "\"")
	val, err := ParseOperation(str)
	*op = val
	return err
}

func (op Operation) MarshalText() (text []byte, err error) {
	return []byte(op.String()), nil
}

func (op *Operation) UnmarshalText(data []byte) (err error) {
	str := strings.Trim(string(data), "\"")
	*op, err = ParseOperation(str)
	return err
}

type EntityType uint32

const (
	GroupsType EntityType = iota
	ChannelsType
	ClientsType
	DashboardType
	MessagesType
)

const (
	GroupsScopeStr   = "groups"
	ChannelsScopeStr = "channels"
	ClientsScopeStr  = "clients"
	DashboardsStr    = "dashboards"
	MessagesStr      = "messages"
)

func (et EntityType) String() string {
	switch et {
	case GroupsType:
		return GroupsScopeStr
	case ChannelsType:
		return ChannelsScopeStr
	case ClientsType:
		return ClientsScopeStr
	case DashboardType:
		return DashboardsStr
	case MessagesType:
		return MessagesStr
	default:
		return fmt.Sprintf("unknown domain entity type %d", et)
	}
}

func (et EntityType) ValidString() (string, error) {
	str := et.String()
	if str == fmt.Sprintf("unknown operation type %d", et) {
		return "", errors.New(str)
	}
	return str, nil
}

func ParseEntityType(et string) (EntityType, error) {
	switch et {
	case GroupsScopeStr:
		return GroupsType, nil
	case ChannelsScopeStr:
		return ChannelsType, nil
	case ClientsScopeStr:
		return ClientsType, nil
	case DashboardsStr:
		return DashboardType, nil
	case MessagesStr:
		return MessagesType, nil
	default:
		return 0, fmt.Errorf("unknown domain entity type %s", et)
	}
}

func (et EntityType) MarshalJSON() ([]byte, error) {
	return json.Marshal(et.String())
}

func (et *EntityType) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), "\"")
	val, err := ParseEntityType(str)
	*et = val
	return err
}

func (et EntityType) MarshalText() ([]byte, error) {
	return []byte(et.String()), nil
}

func (et *EntityType) UnmarshalText(data []byte) (err error) {
	str := strings.Trim(string(data), "\"")
	*et, err = ParseEntityType(str)
	return err
}

var ValidOperationsForEntity = map[EntityType][]Operation{
	ClientsType: {
		ClientCreateOp,
		ClientListOp,
		ClientViewOp,
		ClientUpdateOp,
		ClientUpdateTagsOp,
		ClientUpdateSecretOp,
		ClientEnableOp,
		ClientDisableOp,
		ClientDeleteOp,
		ClientSetParentGroupOp,
		ClientRemoveParentGroupOp,
		ClientConnectToChannelOp,
		ClientDisconnectFromChannelOp,
		RoleAddOp,
		RoleRemoveOp,
		RoleUpdateOp,
		RoleRetrieveOp,
		RoleRetrieveAllOp,
		RoleAddActionsOp,
		RoleListActionsOp,
		RoleCheckActionsExistsOp,
		RoleRemoveActionsOp,
		RoleRemoveAllActionsOp,
		RoleAddMembersOp,
		RoleListMembersOp,
		RoleCheckMembersExistsOp,
		RoleRemoveMembersOp,
		RoleRemoveAllMembersOp,
	},
	ChannelsType: {
		ChannelCreateOp,
		ChannelListOp,
		ChannelViewOp,
		ChannelUpdateOp,
		ChannelUpdateTagsOp,
		ChannelEnableOp,
		ChannelDisableOp,
		ChannelDeleteOp,
		ChannelSetParentGroupOp,
		ChannelRemoveParentGroupOp,
		ChannelConnectToClientOp,
		ChannelDisconnectFromClientOp,
		RoleAddOp,
		RoleRemoveOp,
		RoleUpdateOp,
		RoleRetrieveOp,
		RoleRetrieveAllOp,
		RoleAddActionsOp,
		RoleListActionsOp,
		RoleCheckActionsExistsOp,
		RoleRemoveActionsOp,
		RoleRemoveAllActionsOp,
		RoleAddMembersOp,
		RoleListMembersOp,
		RoleCheckMembersExistsOp,
		RoleRemoveMembersOp,
		RoleRemoveAllMembersOp,
	},
	GroupsType: {
		GroupCreateOp,
		GroupListOp,
		GroupViewOp,
		GroupUpdateOp,
		GroupUpdateTagsOp,
		GroupEnableOp,
		GroupDisableOp,
		GroupDeleteOp,
		GroupRetrieveHierarchyOp,
		GroupAddParentGroupOp,
		GroupRemoveParentGroupOp,
		GroupAddChildrenGroupsOp,
		GroupRemoveChildrenGroupsOp,
		GroupRemoveAllChildrenGroupsOp,
		GroupListChildrenGroupsOp,
		GroupSetChildClientOp,
		GroupRemoveChildClientOp,
		GroupSetChildChannelOp,
		GroupRemoveChildChannelOp,
		RoleAddOp,
		RoleRemoveOp,
		RoleUpdateOp,
		RoleRetrieveOp,
		RoleRetrieveAllOp,
		RoleAddActionsOp,
		RoleListActionsOp,
		RoleCheckActionsExistsOp,
		RoleRemoveActionsOp,
		RoleRemoveAllActionsOp,
		RoleAddMembersOp,
		RoleListMembersOp,
		RoleCheckMembersExistsOp,
		RoleRemoveMembersOp,
		RoleRemoveAllMembersOp,
	},
	DashboardType: {
		DashboardShareOp,
		DashboardUnshareOp,
	},
	MessagesType: {
		MessagePublishOp,
		MessageSubscribeOp,
	},
}

// IsValidOperationForEntity checks if the given operation is valid for the entity type.
func IsValidOperationForEntity(entityType EntityType, operation Operation) bool {
	validOps, exists := ValidOperationsForEntity[entityType]
	if !exists {
		return false
	}
	for _, op := range validOps {
		if op == operation {
			return true
		}
	}
	return false
}

// Example Scope as JSON
//
// [
//     {
//         "optional_domain_id": "domain_1",
//         "entity_type": "groups",
//         "operation": "group_create",
//         "entity_id": "*"
//     },
//     {
//         "optional_domain_id": "domain_1",
//         "entity_type": "channels",
//         "operation": "channel_delete",
//         "entity_id": "channel1"
//     },
//     {
//         "optional_domain_id": "domain_1",
//         "entity_type": "clients",
//         "operation": "client_update",
//         "entity_id": "*"
//     }
// ]

type Scope struct {
	ID               string     `json:"id"`
	PatID            string     `json:"pat_id"`
	OptionalDomainID string     `json:"optional_domain_id"`
	EntityType       EntityType `json:"entity_type"`
	EntityID         string     `json:"entity_id"`
	Operation        Operation  `json:"operation"`
}

func (s *Scope) Authorized(entityType EntityType, optionalDomainID string, operation Operation, entityID string) bool {
	if s == nil {
		return false
	}

	if s.EntityType != entityType {
		return false
	}

	if optionalDomainID != "" && s.OptionalDomainID != optionalDomainID {
		return false
	}

	if s.Operation != operation {
		return false
	}

	if s.EntityID == "*" {
		return true
	}

	if s.EntityID == entityID {
		return true
	}
	return false
}

func (s *Scope) Validate() error {
	if s == nil {
		return errInvalidScope
	}
	if s.EntityID == "" {
		return apiutil.ErrMissingEntityID
	}

	switch s.EntityType {
	case ChannelsType, GroupsType, ClientsType:
		if s.OptionalDomainID == "" {
			return apiutil.ErrMissingDomainID
		}
	}

	if !IsValidOperationForEntity(s.EntityType, s.Operation) {
		return errors.Wrap(apiutil.ErrInvalidQueryParams, errors.New("operation not valid for entity type"))
	}

	return nil
}

// PAT represents Personal Access Token.
type PAT struct {
	ID          string    `json:"id,omitempty"`
	User        string    `json:"user_id,omitempty"`
	Name        string    `json:"name,omitempty"`
	Description string    `json:"description,omitempty"`
	Secret      string    `json:"secret,omitempty"`
	Role        Role      `json:"role,omitempty"`
	IssuedAt    time.Time `json:"issued_at,omitempty"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
	LastUsedAt  time.Time `json:"last_used_at,omitempty"`
	Revoked     bool      `json:"revoked,omitempty"`
	RevokedAt   time.Time `json:"revoked_at,omitempty"`
	Status      Status    `json:"status,omitempty"`
}

type PATSPageMeta struct {
	Offset uint64 `json:"offset"`
	Limit  uint64 `json:"limit"`
	Name   string `json:"name"`
	ID     string `json:"id"`
	Status Status `json:"status"`
}
type PATSPage struct {
	Total  uint64 `json:"total"`
	Offset uint64 `json:"offset"`
	Limit  uint64 `json:"limit"`
	PATS   []PAT  `json:"pats"`
}

type ScopesPageMeta struct {
	Offset uint64 `json:"offset"`
	Limit  uint64 `json:"limit"`
	PatID  string `json:"pat_id"`
	ID     string `json:"id"`
}

type ScopesPage struct {
	Total  uint64  `json:"total"`
	Offset uint64  `json:"offset"`
	Limit  uint64  `json:"limit"`
	Scopes []Scope `json:"scopes"`
}

func (pat PAT) MarshalBinary() ([]byte, error) {
	return json.Marshal(pat)
}

func (pat *PAT) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, pat)
}

// Validate checks if the PAT has valid fields.
func (pat *PAT) Validate() error {
	if pat == nil {
		return errors.New("PAT cannot be nil")
	}
	if pat.Name == "" {
		return errors.New("PAT name cannot be empty")
	}
	if pat.User == "" {
		return errors.New("PAT user cannot be empty")
	}
	return nil
}

func (pat *PAT) String() string {
	str, err := json.MarshalIndent(pat, "", "  ")
	if err != nil {
		return fmt.Sprintf("failed to convert PAT to string: json marshal error :%s", err.Error())
	}
	return string(str)
}

// PATS specifies function which are required for Personal access Token implementation.
type PATS interface {
	// Create function creates new PAT for given valid inputs.
	CreatePAT(ctx context.Context, token, name, description string, duration time.Duration) (PAT, error)

	// UpdateName function updates the name for the given PAT ID.
	UpdatePATName(ctx context.Context, token, patID, name string) (PAT, error)

	// UpdateDescription function updates the description for the given PAT ID.
	UpdatePATDescription(ctx context.Context, token, patID, description string) (PAT, error)

	// Retrieve function retrieves the PAT for given ID.
	RetrievePAT(ctx context.Context, userID string, patID string) (PAT, error)

	// RemoveAllPAT function removes all PATs of user.
	RemoveAllPAT(ctx context.Context, token string) error

	// ListPATS function lists all the PATs for the user.
	ListPATS(ctx context.Context, token string, pm PATSPageMeta) (PATSPage, error)

	// Delete function deletes the PAT for given ID.
	DeletePAT(ctx context.Context, token, patID string) error

	// ResetSecret function reset the secret and creates new secret for the given ID.
	ResetPATSecret(ctx context.Context, token, patID string, duration time.Duration) (PAT, error)

	// RevokeSecret function revokes the secret for the given ID.
	RevokePATSecret(ctx context.Context, token, patID string) error

	// AddScope function adds a new scope.
	AddScope(ctx context.Context, token, patID string, scopes []Scope) error

	// RemoveScope function removes a scope.
	RemoveScope(ctx context.Context, token string, patID string, scopeIDs ...string) error

	// RemovePATAllScope function removes all scope.
	RemovePATAllScope(ctx context.Context, token, patID string) error

	// List function lists all the Scopes for the patID.
	ListScopes(ctx context.Context, token string, pm ScopesPageMeta) (ScopesPage, error)

	// IdentifyPAT function will valid the secret.
	IdentifyPAT(ctx context.Context, paToken string) (PAT, error)

	// AuthorizePAT function will valid the secret and check the given scope exists.
	AuthorizePAT(ctx context.Context, userID, patID string, entityType EntityType, optionalDomainID string, operation Operation, entityID string) error
}

// PATSRepository specifies PATS persistence API.
type PATSRepository interface {
	// Save persists the PAT
	Save(ctx context.Context, pat PAT) (err error)

	// Retrieve retrieves users PAT by its unique identifier.
	Retrieve(ctx context.Context, userID, patID string) (pat PAT, err error)

	// RetrieveScope retrieves PAT scopes by its unique identifier.
	RetrieveScope(ctx context.Context, pm ScopesPageMeta) (scopes ScopesPage, err error)

	// RetrieveSecretAndRevokeStatus retrieves secret and revoke status of PAT by its unique identifier.
	RetrieveSecretAndRevokeStatus(ctx context.Context, userID, patID string) (string, bool, bool, error)

	// UpdateName updates the name of a PAT.
	UpdateName(ctx context.Context, userID, patID, name string) (PAT, error)

	// UpdateDescription updates the description of a PAT.
	UpdateDescription(ctx context.Context, userID, patID, description string) (PAT, error)

	// UpdateTokenHash updates the token hash of a PAT.
	UpdateTokenHash(ctx context.Context, userID, patID, tokenHash string, expiryAt time.Time) (PAT, error)

	// RetrieveAll retrieves all PATs belongs to userID.
	RetrieveAll(ctx context.Context, userID string, pm PATSPageMeta) (pats PATSPage, err error)

	// Revoke PAT with provided ID.
	Revoke(ctx context.Context, userID, patID string) error

	// Reactivate PAT with provided ID.
	Reactivate(ctx context.Context, userID, patID string) error

	// Remove removes Key with provided ID.
	Remove(ctx context.Context, userID, patID string) error

	// RemoveAllPAT removes all PAT for a given user.
	RemoveAllPAT(ctx context.Context, userID string) error

	AddScope(ctx context.Context, userID string, scopes []Scope) error

	RemoveScope(ctx context.Context, userID string, scopesIDs ...string) error

	CheckScope(ctx context.Context, userID, patID string, entityType EntityType, optionalDomainID string, operation Operation, entityID string) error

	RemoveAllScope(ctx context.Context, patID string) error
}

type Cache interface {
	Save(ctx context.Context, userID string, scopes []Scope) error

	CheckScope(ctx context.Context, userID, patID, optionalDomainID string, entityType EntityType, operation Operation, entityID string) bool

	Remove(ctx context.Context, userID string, scopesID []string) error

	RemoveUserAllScope(ctx context.Context, userID string) error

	RemoveAllScope(ctx context.Context, userID, patID string) error
}
