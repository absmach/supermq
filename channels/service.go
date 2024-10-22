package channels

import (
	"context"
	"time"

	"github.com/absmach/magistrala"
	"github.com/absmach/magistrala/pkg/authn"
	mgclients "github.com/absmach/magistrala/pkg/clients"
	"github.com/absmach/magistrala/pkg/errors"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	"github.com/absmach/magistrala/pkg/policies"
	"github.com/absmach/magistrala/pkg/roles"
	"golang.org/x/sync/errgroup"
)

var (
	errCreateChannelsPolicies = errors.New("failed to create channels policies")
	errRollbackRepo           = errors.New("failed to rollback repo")
)

type service struct {
	repo       Repository
	policy     policies.Service
	idProvider magistrala.IDProvider
	roles.ProvisionManageService
}

var _ Service = (*service)(nil)

func New(repo Repository, policy policies.Service, idProvider magistrala.IDProvider, sidProvider magistrala.IDProvider) (Service, error) {
	rpms, err := roles.NewProvisionManageService(policies.ChannelType, repo, policy, sidProvider, AvailableActions(), BuiltInRoles())
	if err != nil {
		return nil, err
	}

	return service{
		repo:                   repo,
		policy:                 policy,
		idProvider:             idProvider,
		ProvisionManageService: rpms,
	}, nil
}

func (svc service) CreateChannels(ctx context.Context, session authn.Session, chs ...Channel) ([]Channel, error) {
	var clients []Channel
	for _, c := range chs {
		if c.ID == "" {
			clientID, err := svc.idProvider.ID()
			if err != nil {
				return []Channel{}, err
			}
			c.ID = clientID
		}

		if c.Status != mgclients.DisabledStatus && c.Status != mgclients.EnabledStatus {
			return []Channel{}, svcerr.ErrInvalidStatus
		}
		c.Domain = session.DomainID
		c.CreatedAt = time.Now()
		clients = append(clients, c)
	}

	saved, err := svc.repo.Save(ctx, clients...)
	if err != nil {
		return nil, errors.Wrap(svcerr.ErrCreateEntity, err)
	}
	chIDs := []string{}
	for _, c := range saved {
		chIDs = append(chIDs, c.ID)
	}

	defer func() {
		if err != nil {
			if errRollBack := svc.repo.Remove(ctx, chIDs...); errRollBack != nil {
				err = errors.Wrap(err, errors.Wrap(errRollbackRepo, errRollBack))
			}
		}
	}()

	newBuiltInRoleMembers := map[roles.BuiltInRoleName][]roles.Member{
		BuiltInRoleAdmin: {roles.Member(session.UserID)},
	}

	optionalPolicies := []policies.Policy{}

	for _, chID := range chIDs {
		optionalPolicies = append(optionalPolicies,
			policies.Policy{
				Domain:      session.DomainID,
				SubjectType: policies.UserType,
				Subject:     session.DomainID,
				Relation:    policies.DomainRelation,
				ObjectType:  policies.ChannelType,
				Object:      chID,
			},
		)
	}
	if _, err := svc.AddNewEntitiesRoles(ctx, session.DomainID, session.UserID, chIDs, optionalPolicies, newBuiltInRoleMembers); err != nil {
		return []Channel{}, errors.Wrap(errCreateChannelsPolicies, err)
	}
	return saved, nil
}

func (svc service) UpdateChannel(ctx context.Context, session authn.Session, ch Channel) (Channel, error) {
	channel := Channel{
		ID:        ch.ID,
		Name:      ch.Name,
		Metadata:  ch.Metadata,
		UpdatedAt: time.Now(),
		UpdatedBy: session.UserID,
	}
	channel, err := svc.repo.Update(ctx, channel)
	if err != nil {
		return Channel{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}
	return channel, nil
}

func (svc service) UpdateChannelTags(ctx context.Context, session authn.Session, ch Channel) (Channel, error) {

	channel := Channel{
		ID:        ch.ID,
		Tags:      ch.Tags,
		UpdatedAt: time.Now(),
		UpdatedBy: session.UserID,
	}
	channel, err := svc.repo.UpdateTags(ctx, channel)
	if err != nil {
		return Channel{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}
	return channel, nil
}

func (svc service) EnableChannel(ctx context.Context, session authn.Session, id string) (Channel, error) {
	channel := Channel{
		ID:        id,
		Status:    mgclients.EnabledStatus,
		UpdatedAt: time.Now(),
	}
	ch, err := svc.changeChannelStatus(ctx, session.UserID, channel)
	if err != nil {
		return Channel{}, errors.Wrap(mgclients.ErrEnableClient, err)
	}

	return ch, nil
}

func (svc service) DisableChannel(ctx context.Context, session authn.Session, id string) (Channel, error) {
	channel := Channel{
		ID:        id,
		Status:    mgclients.DisabledStatus,
		UpdatedAt: time.Now(),
	}
	ch, err := svc.changeChannelStatus(ctx, session.UserID, channel)
	if err != nil {
		return Channel{}, errors.Wrap(mgclients.ErrDisableClient, err)
	}

	return ch, nil
}

func (svc service) ViewChannel(ctx context.Context, session authn.Session, id string) (Channel, error) {
	channel, err := svc.repo.RetrieveByID(ctx, id)
	if err != nil {
		return Channel{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	return channel, nil
}

func (svc service) ListChannels(ctx context.Context, session authn.Session, pm PageMetadata) (Page, error) {
	var ids []string
	var err error
	if !session.SuperAdmin {
		ids, err = svc.listChannelIDs(ctx, session.DomainUserID, pm.Permission)
		if err != nil {
			return Page{}, errors.Wrap(svcerr.ErrNotFound, err)
		}
	}
	if len(ids) == 0 && pm.Domain == "" {
		return Page{}, nil
	}
	pm.IDs = ids

	cp, err := svc.repo.RetrieveAll(ctx, pm)
	if err != nil {
		return Page{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}

	if pm.ListPerms && len(cp.Channels) > 0 {
		g, ctx := errgroup.WithContext(ctx)

		for i := range cp.Channels {
			// Copying loop variable "i" to avoid "loop variable captured by func literal"
			iter := i
			g.Go(func() error {
				return svc.retrievePermissions(ctx, session.DomainUserID, &cp.Channels[iter])
			})
		}

		if err := g.Wait(); err != nil {
			return Page{}, err
		}
	}
	return cp, nil
}

func (svc service) ListChannelsByThing(ctx context.Context, session authn.Session, thID string, pm PageMetadata) (Page, error) {

	return Page{}, nil
}

func (svc service) RemoveChannel(ctx context.Context, session authn.Session, id string) error {
	if _, err := svc.repo.ChangeStatus(ctx, Channel{ID: id, Status: mgclients.DeletedStatus}); err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}

	filterDeletePolicies := []policies.Policy{
		{
			SubjectType: policies.ChannelType,
			Subject:     id,
		},
		{
			ObjectType: policies.ChannelType,
			Object:     id,
		},
	}
	deletePolicies := []policies.Policy{
		{
			SubjectType: policies.DomainType,
			Subject:     session.DomainID,
			Relation:    policies.DomainRelation,
			ObjectType:  policies.ChannelType,
			Object:      id,
		},
	}

	if err := svc.RemoveEntitiesRoles(ctx, session.DomainID, session.DomainUserID, []string{id}, filterDeletePolicies, deletePolicies); err != nil {
		return errors.Wrap(svcerr.ErrDeletePolicies, err)
	}

	if err := svc.repo.Remove(ctx, id); err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}

	return nil
}

func (svc service) Connect(ctx context.Context, session authn.Session, chIDs, thIDs []string) (retErr error) {

	prs := []policies.Policy{}
	for _, chID := range chIDs {
		for _, thID := range thIDs {
			prs = append(prs, policies.Policy{
				SubjectType: policies.ThingType,
				Subject:     thID,
				Relation:    "connect",
				Object:      chID,
				ObjectType:  policies.ChannelType,
			})
		}
	}
	if err := svc.policy.AddPolicies(ctx, prs); err != nil {
		return errors.Wrap(svcerr.ErrAddPolicies, err)
	}
	defer func() {
		if retErr != nil {
			if errRollback := svc.policy.DeletePolicies(ctx, prs); errRollback != nil {
				retErr = errors.Wrap(retErr, errRollback)
			}
		}
	}()

	if err := svc.repo.Connect(ctx, chIDs, thIDs); err != nil {
		return errors.Wrap(svcerr.ErrCreateEntity, err)
	}

	return nil
}

func (svc service) Disconnect(ctx context.Context, session authn.Session, chIDs, thIDs []string) (retErr error) {

	prs := []policies.Policy{}
	for _, chID := range chIDs {
		for _, thID := range thIDs {

			prs = append(prs, policies.Policy{
				SubjectType: policies.ThingType,
				Subject:     thID,
				Relation:    "connect",
				Object:      chID,
				ObjectType:  policies.ChannelType,
			})
		}
	}
	if err := svc.policy.DeletePolicies(ctx, prs); err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}
	defer func() {
		if retErr != nil {
			if errRollback := svc.policy.AddPolicies(ctx, prs); errRollback != nil {
				retErr = errors.Wrap(retErr, errRollback)
			}
		}
	}()
	if err := svc.repo.Disconnect(ctx, chIDs, thIDs); err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}

	return nil
}

type identity struct {
	ID       string
	DomainID string
	UserID   string
}

func (svc service) listChannelIDs(ctx context.Context, userID, permission string) ([]string, error) {
	tids, err := svc.policy.ListAllObjects(ctx, policies.Policy{
		SubjectType: policies.UserType,
		Subject:     userID,
		Permission:  permission,
		ObjectType:  policies.ChannelType,
	})
	if err != nil {
		return nil, errors.Wrap(svcerr.ErrNotFound, err)
	}
	return tids.Policies, nil
}

func (svc service) retrievePermissions(ctx context.Context, userID string, channel *Channel) error {
	permissions, err := svc.listUserThingPermission(ctx, userID, channel.ID)
	if err != nil {
		return err
	}
	channel.Permissions = permissions
	return nil
}

func (svc service) listUserThingPermission(ctx context.Context, userID, thingID string) ([]string, error) {
	lp, err := svc.policy.ListPermissions(ctx, policies.Policy{
		SubjectType: policies.UserType,
		Subject:     userID,
		Object:      thingID,
		ObjectType:  policies.ChannelType,
	}, []string{})
	if err != nil {
		return []string{}, errors.Wrap(svcerr.ErrAuthorization, err)
	}
	return lp, nil
}

func (svc service) changeChannelStatus(ctx context.Context, userID string, channel Channel) (Channel, error) {

	dbchannel, err := svc.repo.RetrieveByID(ctx, channel.ID)
	if err != nil {
		return Channel{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	if dbchannel.Status == channel.Status {
		return Channel{}, errors.ErrStatusAlreadyAssigned
	}

	channel.UpdatedBy = userID

	channel, err = svc.repo.ChangeStatus(ctx, channel)
	if err != nil {
		return Channel{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}
	return channel, nil
}
