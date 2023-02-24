package groups

import (
	"context"
	"fmt"
	"time"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/things/policies"
	upolicies "github.com/mainflux/mainflux/users/policies"
)

// Possible token types are access and refresh tokens.
const (
	thingsObjectKey   = "things"
	createKey         = "g_add"
	updateRelationKey = "g_update"
	listRelationKey   = "g_list"
	deleteRelationKey = "g_delete"
	entityType        = "group"
)

var (
	// ErrInvalidStatus indicates invalid status.
	ErrInvalidStatus = errors.New("invalid groups status")

	// ErrEnableGroup indicates error in enabling group.
	ErrEnableGroup = errors.New("failed to enable group")

	// ErrDisableGroup indicates error in disabling group.
	ErrDisableGroup = errors.New("failed to disable group")

	// ErrStatusAlreadyAssigned indicated that the group has already been assigned the status.
	ErrStatusAlreadyAssigned = errors.New("status already assigned")
)

type service struct {
	auth       upolicies.AuthServiceClient
	groups     Repository
	idProvider mainflux.IDProvider
}

// NewService returns a new Clients service implementation.
func NewService(auth upolicies.AuthServiceClient, g Repository, p policies.Repository, idp mainflux.IDProvider) Service {
	return service{
		auth:       auth,
		groups:     g,
		idProvider: idp,
	}
}

func (svc service) CreateGroups(ctx context.Context, token string, gs ...Group) ([]Group, error) {
	res, err := svc.auth.Identify(ctx, &upolicies.Token{Value: token})
	if err != nil {
		return []Group{}, errors.Wrap(errors.ErrAuthentication, err)
	}
	if err := svc.authorize(ctx, token, thingsObjectKey, createKey); err != nil {
		return []Group{}, err
	}
	var grps []Group
	for _, g := range gs {
		if g.ID == "" {
			groupID, err := svc.idProvider.ID()
			if err != nil {
				return []Group{}, err
			}
			g.ID = groupID
		}
		if g.Owner == "" {
			g.Owner = res.GetId()
		}

		if g.Status != EnabledStatus && g.Status != DisabledStatus {
			return []Group{}, apiutil.ErrInvalidStatus
		}

		g.CreatedAt = time.Now()
		g.UpdatedAt = g.CreatedAt
		grp, err := svc.groups.Save(ctx, g)
		if err != nil {
			return []Group{}, err
		}
		grps = append(grps, grp)
	}
	return grps, nil
}

func (svc service) ViewGroup(ctx context.Context, token string, id string) (Group, error) {
	if err := svc.authorize(ctx, token, id, listRelationKey); err != nil {
		return Group{}, errors.Wrap(errors.ErrNotFound, err)
	}
	return svc.groups.RetrieveByID(ctx, id)
}

func (svc service) ListGroups(ctx context.Context, token string, gm GroupsPage) (GroupsPage, error) {
	res, err := svc.auth.Identify(ctx, &upolicies.Token{Value: token})
	if err != nil {
		return GroupsPage{}, errors.Wrap(errors.ErrAuthentication, err)
	}

	// If the user is admin, fetch all channels from the database.
	if err := svc.authorize(ctx, token, thingsObjectKey, listRelationKey); err == nil {
		page, err := svc.groups.RetrieveAll(ctx, gm)
		if err != nil {
			return GroupsPage{}, err
		}
		return page, err
	}

	gm.Subject = res.GetId()
	gm.OwnerID = res.GetId()
	gm.Action = "g_list"
	return svc.groups.RetrieveAll(ctx, gm)
}

func (svc service) ListMemberships(ctx context.Context, token, clientID string, gm GroupsPage) (MembershipsPage, error) {
	res, err := svc.auth.Identify(ctx, &upolicies.Token{Value: token})
	if err != nil {
		return MembershipsPage{}, errors.Wrap(errors.ErrAuthentication, err)
	}

	// If the user is admin, fetch all channels from the database.
	if err := svc.authorize(ctx, token, thingsObjectKey, listRelationKey); err == nil {
		return svc.groups.Memberships(ctx, clientID, gm)
	}

	gm.Subject = res.GetId()
	gm.OwnerID = res.GetId()
	gm.Action = "g_list"
	return svc.groups.Memberships(ctx, clientID, gm)
}

func (svc service) UpdateGroup(ctx context.Context, token string, g Group) (Group, error) {
	res, err := svc.auth.Identify(ctx, &upolicies.Token{Value: token})
	if err != nil {
		return Group{}, errors.Wrap(errors.ErrAuthentication, err)
	}

	if err := svc.authorize(ctx, token, g.ID, updateRelationKey); err != nil {
		return Group{}, errors.Wrap(errors.ErrNotFound, err)
	}

	g.Owner = res.GetId()
	g.UpdatedAt = time.Now()

	return svc.groups.Update(ctx, g)
}

func (svc service) EnableGroup(ctx context.Context, token, id string) (Group, error) {
	group, err := svc.changeGroupStatus(ctx, token, id, EnabledStatus)
	if err != nil {
		return Group{}, errors.Wrap(ErrEnableGroup, err)
	}
	return group, nil
}

func (svc service) DisableGroup(ctx context.Context, token, id string) (Group, error) {
	group, err := svc.changeGroupStatus(ctx, token, id, DisabledStatus)
	if err != nil {
		return Group{}, errors.Wrap(ErrDisableGroup, err)
	}
	return group, nil
}

func (svc service) changeGroupStatus(ctx context.Context, token, id string, status Status) (Group, error) {
	if err := svc.authorize(ctx, token, id, deleteRelationKey); err != nil {
		return Group{}, errors.Wrap(errors.ErrNotFound, err)
	}
	dbGroup, err := svc.groups.RetrieveByID(ctx, id)
	if err != nil {
		return Group{}, err
	}
	fmt.Println(dbGroup)
	if dbGroup.Status == status {
		return Group{}, ErrStatusAlreadyAssigned
	}

	return svc.groups.ChangeStatus(ctx, id, status)
}

func (svc service) authorize(ctx context.Context, subject, object string, relation string) error {
	req := &upolicies.AuthorizeReq{
		Sub:        subject,
		Obj:        object,
		Act:        relation,
		EntityType: entityType,
	}
	res, err := svc.auth.Authorize(ctx, req)
	if err != nil {
		return errors.Wrap(errors.ErrAuthorization, err)
	}
	if !res.GetAuthorized() {
		return errors.ErrAuthorization
	}
	return nil
}
