package users

import (
	"context"
	"time"

	"github.com/mainflux/mainflux/internal/groups"

	"github.com/mainflux/mainflux/pkg/errors"
)

func (svc usersService) CreateGroup(ctx context.Context, token string, group groups.Group) (groups.Group, error) {
	user, err := svc.identify(ctx, token)
	if err != nil {
		return groups.Group{}, errors.Wrap(ErrUnauthorizedAccess, err)
	}

	ulid, err := svc.idProvider.ID()
	if err != nil {
		return groups.Group{}, err
	}

	timestamp := time.Now().UTC().Round(time.Millisecond)

	group.UpdatedAt = timestamp
	group.CreatedAt = timestamp

	group.ID = ulid
	group.OwnerID = user.id

	group, err = svc.groups.Save(ctx, group)
	if err != nil {
		return groups.Group{}, err
	}

	return group, nil
}

func (svc usersService) ListGroups(ctx context.Context, token string, pm groups.PageMetadata) (groups.GroupPage, error) {
	if _, err := svc.identify(ctx, token); err != nil {
		return groups.GroupPage{}, errors.Wrap(ErrUnauthorizedAccess, err)
	}
	return svc.groups.RetrieveAll(ctx, pm)
}

func (svc usersService) ListParents(ctx context.Context, token string, childID string, pm groups.PageMetadata) (groups.GroupPage, error) {
	if _, err := svc.identify(ctx, token); err != nil {
		return groups.GroupPage{}, errors.Wrap(ErrUnauthorizedAccess, err)
	}
	return svc.groups.RetrieveAllParents(ctx, childID, pm)
}

func (svc usersService) ListChildren(ctx context.Context, token string, parentID string, pm groups.PageMetadata) (groups.GroupPage, error) {
	if _, err := svc.identify(ctx, token); err != nil {
		return groups.GroupPage{}, errors.Wrap(ErrUnauthorizedAccess, err)
	}
	return svc.groups.RetrieveAllChildren(ctx, parentID, pm)
}

func (svc usersService) RemoveGroup(ctx context.Context, token, id string) error {
	if _, err := svc.identify(ctx, token); err != nil {
		return errors.Wrap(ErrUnauthorizedAccess, err)
	}
	return svc.groups.Delete(ctx, id)
}

func (svc usersService) UpdateGroup(ctx context.Context, token string, group groups.Group) (groups.Group, error) {
	if _, err := svc.identify(ctx, token); err != nil {
		return groups.Group{}, errors.Wrap(ErrUnauthorizedAccess, err)
	}

	group.UpdatedAt = time.Now().UTC().Round(time.Millisecond)

	return svc.groups.Update(ctx, group)
}

func (svc usersService) ViewGroup(ctx context.Context, token, id string) (groups.Group, error) {
	if _, err := svc.identify(ctx, token); err != nil {
		return groups.Group{}, errors.Wrap(ErrUnauthorizedAccess, err)
	}
	return svc.groups.RetrieveByID(ctx, id)
}

func (svc usersService) Assign(ctx context.Context, token, groupID string, memberIDs ...string) error {
	if _, err := svc.identify(ctx, token); err != nil {
		return errors.Wrap(ErrUnauthorizedAccess, err)
	}

	return svc.groups.Assign(ctx, groupID, memberIDs...)
}

func (svc usersService) Unassign(ctx context.Context, token string, groupID string, memberIDs ...string) error {
	if _, err := svc.identify(ctx, token); err != nil {
		return errors.Wrap(ErrUnauthorizedAccess, err)
	}

	return svc.groups.Unassign(ctx, groupID, memberIDs...)
}

func (svc usersService) ListMemberships(ctx context.Context, token string, memberID string, pm groups.PageMetadata) (groups.GroupPage, error) {
	if _, err := svc.identify(ctx, token); err != nil {
		return groups.GroupPage{}, errors.Wrap(ErrUnauthorizedAccess, err)
	}
	return svc.groups.Memberships(ctx, memberID, pm)
}

func (svc usersService) ListMembers(ctx context.Context, token, groupID string, m groups.PageMetadata) (UserPage, error) {
	if _, err := svc.identify(ctx, token); err != nil {
		return UserPage{}, err
	}

	userIDs, err := svc.members(ctx, token, groupID, m.Offset, m.Limit)
	if err != nil {
		return UserPage{}, err
	}

	if len(userIDs) == 0 {
		return UserPage{
			Users: []User{},
			PageMetadata: PageMetadata{
				Total:  0,
				Offset: m.Offset,
				Limit:  m.Limit,
			},
		}, nil
	}

	return svc.users.RetrieveAll(ctx, m.Offset, m.Limit, userIDs, "", nil)
}
