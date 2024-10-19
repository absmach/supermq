package channels

import (
	"context"
	"time"

	"github.com/absmach/magistrala"
	"github.com/absmach/magistrala/pkg/authn"
	"github.com/absmach/magistrala/pkg/channels"
	mgclients "github.com/absmach/magistrala/pkg/clients"
	"github.com/absmach/magistrala/pkg/entityroles"
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

type channelsService struct {
	repo        channels.Repository
	policy      policies.Service
	idProvider  magistrala.IDProvider
	sidProvider magistrala.IDProvider
	entityroles.RolesSvc
}

var _ channels.Service = (*channelsService)(nil)

func New(repo channels.Repository, policy policies.Service, idProvider magistrala.IDProvider, sidProvider magistrala.IDProvider) (channels.Service, error) {
	rolesSvc, err := entityroles.NewRolesSvc(policies.ChannelType, repo, sidProvider, policy, channels.AvailableActions(), channels.BuiltInRoles())
	if err != nil {
		return nil, err
	}

	opp := channels.NewOperationPerm()
	if err := opp.AddOperationPermissionMap(channels.NewOperationPermissionMap()); err != nil {
		return channelsService{}, err
	}
	if err := opp.Validate(); err != nil {
		return channelsService{}, err
	}
	return channelsService{
		repo:        repo,
		policy:      policy,
		idProvider:  idProvider,
		sidProvider: sidProvider,
		RolesSvc:    rolesSvc,
	}, nil
}

func (cs channelsService) CreateChannels(ctx context.Context, session authn.Session, chs ...channels.Channel) ([]channels.Channel, error) {
	var clients []channels.Channel
	for _, c := range chs {
		if c.ID == "" {
			clientID, err := cs.idProvider.ID()
			if err != nil {
				return []channels.Channel{}, err
			}
			c.ID = clientID
		}

		if c.Status != mgclients.DisabledStatus && c.Status != mgclients.EnabledStatus {
			return []channels.Channel{}, svcerr.ErrInvalidStatus
		}
		c.Domain = session.DomainID
		c.CreatedAt = time.Now()
		clients = append(clients, c)
	}

	saved, err := cs.repo.Save(ctx, clients...)
	if err != nil {
		return nil, errors.Wrap(svcerr.ErrCreateEntity, err)
	}
	chIDs := []string{}
	for _, c := range saved {
		chIDs = append(chIDs, c.ID)
	}

	defer func() {
		if err != nil {
			if errRollBack := cs.repo.Remove(ctx, chIDs...); errRollBack != nil {
				err = errors.Wrap(err, errors.Wrap(errRollbackRepo, errRollBack))
			}
		}
	}()

	newBuiltInRoleMembers := map[roles.BuiltInRoleName][]roles.Member{
		channels.BuiltInRoleAdmin: {roles.Member(session.UserID)},
	}

	optionalPolicies := []roles.OptionalPolicy{}

	for _, chID := range chIDs {
		optionalPolicies = append(optionalPolicies,
			roles.OptionalPolicy{
				Namespace:   session.DomainID,
				SubjectType: policies.UserType,
				Subject:     session.DomainUserID,
				Relation:    policies.AdministratorRelation,
				ObjectType:  policies.ChannelType,
				Object:      chID,
			},
			roles.OptionalPolicy{

				Namespace:   session.DomainID,
				SubjectType: policies.UserType,
				Subject:     session.DomainUserID,
				Relation:    policies.DomainRelation,
				ObjectType:  policies.ChannelType,
				Object:      chID,
			},
		)
	}
	if _, err := cs.AddNewEntityRoles(ctx, session.UserID, session.DomainID, session.DomainID, newBuiltInRoleMembers, optionalPolicies); err != nil {
		return []channels.Channel{}, errors.Wrap(errCreateChannelsPolicies, err)
	}
	return saved, nil
}

func (cs channelsService) UpdateChannel(ctx context.Context, session authn.Session, ch channels.Channel) (channels.Channel, error) {
	channel := channels.Channel{
		ID:        ch.ID,
		Name:      ch.Name,
		Metadata:  ch.Metadata,
		UpdatedAt: time.Now(),
		UpdatedBy: session.UserID,
	}
	channel, err := cs.repo.Update(ctx, channel)
	if err != nil {
		return channels.Channel{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}
	return channel, nil
}

func (cs channelsService) UpdateChannelTags(ctx context.Context, session authn.Session, ch channels.Channel) (channels.Channel, error) {

	channel := channels.Channel{
		ID:        ch.ID,
		Tags:      ch.Tags,
		UpdatedAt: time.Now(),
		UpdatedBy: session.UserID,
	}
	channel, err := cs.repo.UpdateTags(ctx, channel)
	if err != nil {
		return channels.Channel{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}
	return channel, nil
}

func (cs channelsService) EnableChannel(ctx context.Context, session authn.Session, id string) (channels.Channel, error) {
	channel := channels.Channel{
		ID:        id,
		Status:    mgclients.EnabledStatus,
		UpdatedAt: time.Now(),
	}
	ch, err := cs.changeChannelStatus(ctx, session.UserID, channel)
	if err != nil {
		return channels.Channel{}, errors.Wrap(mgclients.ErrEnableClient, err)
	}

	return ch, nil
}

func (cs channelsService) DisableChannel(ctx context.Context, session authn.Session, id string) (channels.Channel, error) {
	channel := channels.Channel{
		ID:        id,
		Status:    mgclients.DisabledStatus,
		UpdatedAt: time.Now(),
	}
	ch, err := cs.changeChannelStatus(ctx, session.UserID, channel)
	if err != nil {
		return channels.Channel{}, errors.Wrap(mgclients.ErrDisableClient, err)
	}

	return ch, nil
}

func (cs channelsService) ViewChannel(ctx context.Context, session authn.Session, id string) (channels.Channel, error) {
	channel, err := cs.repo.RetrieveByID(ctx, id)
	if err != nil {
		return channels.Channel{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	return channel, nil
}

func (cs channelsService) ListChannels(ctx context.Context, session authn.Session, pm channels.PageMetadata) (channels.Page, error) {
	var ids []string
	var err error
	if !session.SuperAdmin {
		ids, err = cs.listChannelIDs(ctx, session.DomainUserID, pm.Permission)
		if err != nil {
			return channels.Page{}, errors.Wrap(svcerr.ErrNotFound, err)
		}
	}
	if len(ids) == 0 && pm.Domain == "" {
		return channels.Page{}, nil
	}
	pm.IDs = ids

	cp, err := cs.repo.RetrieveAll(ctx, pm)
	if err != nil {
		return channels.Page{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}

	if pm.ListPerms && len(cp.Channels) > 0 {
		g, ctx := errgroup.WithContext(ctx)

		for i := range cp.Channels {
			// Copying loop variable "i" to avoid "loop variable captured by func literal"
			iter := i
			g.Go(func() error {
				return cs.retrievePermissions(ctx, session.DomainUserID, &cp.Channels[iter])
			})
		}

		if err := g.Wait(); err != nil {
			return channels.Page{}, err
		}
	}
	return cp, nil
}

func (cs channelsService) ListChannelsByThing(ctx context.Context, session authn.Session, thID string, pm channels.PageMetadata) (channels.Page, error) {

	return channels.Page{}, nil
}

func (cs channelsService) RemoveChannel(ctx context.Context, session authn.Session, id string) error {

	if err := cs.policy.DeletePolicyFilter(ctx, policies.Policy{
		SubjectType: policies.ThingType,
		Subject:     id,
	}); err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}

	if err := cs.policy.DeletePolicyFilter(ctx, policies.Policy{
		ObjectType: policies.ThingType,
		Object:     id,
	}); err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}

	if err := cs.repo.Remove(ctx, id); err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}

	return nil
}

func (cs channelsService) Connect(ctx context.Context, session authn.Session, chIDs, thIDs []string) (retErr error) {

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
	if err := cs.policy.AddPolicies(ctx, prs); err != nil {
		return errors.Wrap(svcerr.ErrAddPolicies, err)
	}
	defer func() {
		if retErr != nil {
			if errRollback := cs.policy.DeletePolicies(ctx, prs); errRollback != nil {
				retErr = errors.Wrap(retErr, errRollback)
			}
		}
	}()

	if err := cs.repo.Connect(ctx, chIDs, thIDs); err != nil {
		return errors.Wrap(svcerr.ErrCreateEntity, err)
	}

	return nil
}

func (cs channelsService) Disconnect(ctx context.Context, session authn.Session, chIDs, thIDs []string) (retErr error) {

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
	if err := cs.policy.DeletePolicies(ctx, prs); err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}
	defer func() {
		if retErr != nil {
			if errRollback := cs.policy.AddPolicies(ctx, prs); errRollback != nil {
				retErr = errors.Wrap(retErr, errRollback)
			}
		}
	}()
	if err := cs.repo.Disconnect(ctx, chIDs, thIDs); err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}

	return nil
}

type identity struct {
	ID       string
	DomainID string
	UserID   string
}

func (cs channelsService) listChannelIDs(ctx context.Context, userID, permission string) ([]string, error) {
	tids, err := cs.policy.ListAllObjects(ctx, policies.Policy{
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

func (cs channelsService) retrievePermissions(ctx context.Context, userID string, channel *channels.Channel) error {
	permissions, err := cs.listUserThingPermission(ctx, userID, channel.ID)
	if err != nil {
		return err
	}
	channel.Permissions = permissions
	return nil
}

func (cs channelsService) listUserThingPermission(ctx context.Context, userID, thingID string) ([]string, error) {
	lp, err := cs.policy.ListPermissions(ctx, policies.Policy{
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

func (cs channelsService) changeChannelStatus(ctx context.Context, userID string, channel channels.Channel) (channels.Channel, error) {

	dbchannel, err := cs.repo.RetrieveByID(ctx, channel.ID)
	if err != nil {
		return channels.Channel{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	if dbchannel.Status == channel.Status {
		return channels.Channel{}, errors.ErrStatusAlreadyAssigned
	}

	channel.UpdatedBy = userID

	channel, err = cs.repo.ChangeStatus(ctx, channel)
	if err != nil {
		return channels.Channel{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}
	return channel, nil
}
