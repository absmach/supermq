package groups

import (
	"context"
	"time"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/things/policies"
)

// Possible token types are access and refresh tokens.
const (
	usersObjectKey    = "users"
	authoritiesObject = "authorities"
	memberRelationKey = "member"
	readRelationKey   = "read"
	writeRelationKey  = "write"
	deleteRelationKey = "delete"
)

var (
	// ErrInvalidStatus indicates invalid status.
	ErrInvalidStatus = errors.New("invalid groups status")

	// ErrStatusAlreadyAssigned indicated that the client or group has already been assigned the status.
	ErrStatusAlreadyAssigned = errors.New("status already assigned")
)

// Service unites Clients and Group services.
type Service interface {
	GroupService
}

type service struct {
	auth       mainflux.AuthServiceClient
	groups     GroupRepository
	policies   policies.PolicyRepository
	idProvider mainflux.IDProvider
}

// NewService returns a new Clients service implementation.
func NewService(auth mainflux.AuthServiceClient, g GroupRepository, p policies.PolicyRepository, idp mainflux.IDProvider) Service {
	return service{
		auth:       auth,
		groups:     g,
		policies:   p,
		idProvider: idp,
	}
}

func (svc service) CreateGroups(ctx context.Context, token string, gs ...Group) ([]Group, error) {
	res, err := svc.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return []Group{}, errors.Wrap(errors.ErrAuthentication, err)
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
			g.Owner = res.GetEmail()
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
	res, err := svc.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Group{}, errors.Wrap(errors.ErrAuthentication, err)
	}

	if err := svc.authorize(ctx, res.GetId(), id, readRelationKey); err != nil {
		if err := svc.authorize(ctx, res.GetId(), authoritiesObject, memberRelationKey); err != nil {
			return Group{}, errors.Wrap(errors.ErrNotFound, err)
		}
	}
	return svc.groups.RetrieveByID(ctx, id)
}

func (svc service) ListGroups(ctx context.Context, token string, gm GroupsPage) (GroupsPage, error) {
	res, err := svc.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return GroupsPage{}, errors.Wrap(errors.ErrAuthentication, err)
	}

	// If the user is admin, fetch all channels from the database.
	if err := svc.authorize(ctx, res.GetId(), authoritiesObject, memberRelationKey); err == nil {
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
	res, err := svc.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return MembershipsPage{}, errors.Wrap(errors.ErrAuthentication, err)
	}

	gm.Subject = res.GetId()
	gm.Action = "g_list"
	return svc.groups.Memberships(ctx, clientID, gm)
}

func (svc service) UpdateGroup(ctx context.Context, token string, g Group) (Group, error) {
	res, err := svc.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Group{}, errors.Wrap(errors.ErrAuthentication, err)
	}

	if err := svc.authorize(ctx, res.GetId(), g.ID, writeRelationKey); err != nil {
		if err := svc.authorize(ctx, res.GetId(), authoritiesObject, memberRelationKey); err != nil {
			return Group{}, errors.Wrap(errors.ErrNotFound, err)
		}
	}
	g.Owner = res.GetId()
	g.UpdatedAt = time.Now()

	return svc.groups.Update(ctx, g)
}

func (svc service) EnableGroup(ctx context.Context, token, id string) (Group, error) {
	res, err := svc.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Group{}, errors.Wrap(errors.ErrAuthentication, err)
	}

	if err := svc.authorize(ctx, res.GetId(), id, deleteRelationKey); err != nil {
		if err := svc.authorize(ctx, res.GetId(), authoritiesObject, memberRelationKey); err != nil {
			return Group{}, errors.Wrap(errors.ErrNotFound, err)
		}
	}

	group, err := svc.changeGroupStatus(ctx, id, EnabledStatus)
	if err != nil {
		return Group{}, err
	}
	return group, nil
}

func (svc service) DisableGroup(ctx context.Context, token, id string) (Group, error) {
	res, err := svc.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Group{}, errors.Wrap(errors.ErrAuthentication, err)
	}

	if err := svc.authorize(ctx, res.GetId(), id, deleteRelationKey); err != nil {
		if err := svc.authorize(ctx, res.GetId(), authoritiesObject, memberRelationKey); err != nil {
			return Group{}, errors.Wrap(errors.ErrNotFound, err)
		}
	}
	group, err := svc.changeGroupStatus(ctx, id, DisabledStatus)
	if err != nil {
		return Group{}, err
	}
	return group, nil
}

func (svc service) IsChannelOwner(ctx context.Context, owner, chanID string) error {
	g, err := svc.groups.RetrieveByID(ctx, chanID)
	if err != nil {
		return err
	}
	if g.Owner != owner {
		return errors.New("not owner")
	}
	return nil
}

func (svc service) authorize(ctx context.Context, subject, object string, relation string) error {
	req := &mainflux.AuthorizeReq{
		Sub: subject,
		Obj: object,
		Act: relation,
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

func (svc service) changeGroupStatus(ctx context.Context, id string, status Status) (Group, error) {
	dbGroup, err := svc.groups.RetrieveByID(ctx, id)
	if err != nil {
		return Group{}, err
	}
	if dbGroup.Status == status {
		return Group{}, ErrStatusAlreadyAssigned
	}

	return svc.groups.ChangeStatus(ctx, id, status)
}

func (svc service) identify(ctx context.Context, tkn string) (string, error) {
	return "", nil
}
