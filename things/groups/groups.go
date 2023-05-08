package groups

import (
	"context"

	"github.com/mainflux/mainflux/internal/mainflux/groups"
)

// MembershipsPage contains page related metadata as well as list of memberships that
// belong to this page.
type MembershipsPage struct {
	Page
	Memberships []groups.Group
}

// GroupsPage contains page related metadata as well as list
// of Groups that belong to the page.
type GroupsPage struct {
	Page
	Path      string
	Level     uint64
	ID        string
	Direction int64 // ancestors (-1) or descendants (+1)
	Groups    []groups.Group
}

// Repository specifies a group persistence API.
type Repository interface {
	// Save group.
	Save(ctx context.Context, g groups.Group) (groups.Group, error)

	// Update a group.
	Update(ctx context.Context, g groups.Group) (groups.Group, error)

	// RetrieveByID retrieves group by its id.
	RetrieveByID(ctx context.Context, id string) (groups.Group, error)

	// RetrieveAll retrieves all groups.
	RetrieveAll(ctx context.Context, gm GroupsPage) (GroupsPage, error)

	// Memberships retrieves everything that is assigned to a group identified by clientID.
	Memberships(ctx context.Context, clientID string, gm GroupsPage) (MembershipsPage, error)

	// ChangeStatus changes groups status to active or inactive
	ChangeStatus(ctx context.Context, group groups.Group) (groups.Group, error)
}

// Service specifies an API that must be fulfilled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	// CreateGroup creates new  group.
	CreateGroups(ctx context.Context, token string, gs ...groups.Group) ([]groups.Group, error)

	// UpdateGroup updates the group identified by the provided ID.
	UpdateGroup(ctx context.Context, token string, g groups.Group) (groups.Group, error)

	// ViewGroup retrieves data about the group identified by ID.
	ViewGroup(ctx context.Context, token, id string) (groups.Group, error)

	// ListGroups retrieves groups.
	ListGroups(ctx context.Context, token string, gm GroupsPage) (GroupsPage, error)

	// ListMemberships retrieves everything that is assigned to a group identified by clientID.
	ListMemberships(ctx context.Context, token, clientID string, gm GroupsPage) (MembershipsPage, error)

	// EnableGroup logically enables the group identified with the provided ID.
	EnableGroup(ctx context.Context, token, id string) (groups.Group, error)

	// DisableGroup logically disables the group identified with the provided ID.
	DisableGroup(ctx context.Context, token, id string) (groups.Group, error)
}
