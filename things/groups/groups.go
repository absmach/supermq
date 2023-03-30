package groups

import (
	"context"
	"time"

	mfgroups "github.com/mainflux/mainflux/internal/mainflux/groups"
)

// MembershipsPage contains page related metadata as well as list of memberships that
// belong to this page.
type MembershipsPage struct {
	Page
	Memberships []mfgroups.Group
}

// GroupsPage contains page related metadata as well as list
// of Groups that belong to the page.
type GroupsPage struct {
	Page
	Path      string
	Level     uint64
	ID        string
	Direction int64 // ancestors (-1) or descendants (+1)
	Groups    []Group
}

// Group represents the group of Clients.
// Indicates a level in tree hierarchy. Root node is level 1.
// Path in a tree consisting of group IDs
// Paths are unique per owner.
type Group struct {
	ID          string    `json:"id"`
	Owner       string    `json:"owner_id"`
	Parent      string    `json:"parent_id,omitempty"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Metadata    Metadata  `json:"metadata,omitempty"`
	Level       int       `json:"level,omitempty"`
	Path        string    `json:"path,omitempty"`
	Children    []*Group  `json:"children,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	UpdatedBy   string    `json:"updated_by"`
	Status      Status    `json:"status"`
}

// Repository specifies a group persistence API.
type Repository interface {
	// Save group.
	Save(ctx context.Context, g mfgroups.Group) (mfgroups.Group, error)

	// Update a group.
	Update(ctx context.Context, g mfgroups.Group) (mfgroups.Group, error)

	// RetrieveByID retrieves group by its id.
	RetrieveByID(ctx context.Context, id string) (mfgroups.Group, error)

	// RetrieveAll retrieves all groups.
	RetrieveAll(ctx context.Context, gm GroupsPage) (GroupsPage, error)

	// Memberships retrieves everything that is assigned to a group identified by clientID.
	Memberships(ctx context.Context, clientID string, gm GroupsPage) (MembershipsPage, error)

	// ChangeStatus changes groups status to active or inactive
	ChangeStatus(ctx context.Context, group Group) (Group, error)
}

// Service specifies an API that must be fulfilled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	// CreateGroup creates new  group.
	CreateGroups(ctx context.Context, token string, gs ...mfgroups.Group) ([]mfgroups.Group, error)

	// UpdateGroup updates the group identified by the provided ID.
	UpdateGroup(ctx context.Context, token string, g mfgroups.Group) (mfgroups.Group, error)

	// ViewGroup retrieves data about the group identified by ID.
	ViewGroup(ctx context.Context, token, id string) (mfgroups.Group, error)

	// ListGroups retrieves groups.
	ListGroups(ctx context.Context, token string, gm GroupsPage) (GroupsPage, error)

	// ListMemberships retrieves everything that is assigned to a group identified by clientID.
	ListMemberships(ctx context.Context, token, clientID string, gm GroupsPage) (MembershipsPage, error)

	// EnableGroup logically enables the group identified with the provided ID.
	EnableGroup(ctx context.Context, token, id string) (mfgroups.Group, error)

	// DisableGroup logically disables the group identified with the provided ID.
	DisableGroup(ctx context.Context, token, id string) (mfgroups.Group, error)
}
