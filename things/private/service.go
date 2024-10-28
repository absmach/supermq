package private

import (
	"context"

	"github.com/absmach/magistrala/pkg/clients"
	"github.com/absmach/magistrala/pkg/errors"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	"github.com/absmach/magistrala/pkg/policies"
	"github.com/absmach/magistrala/things"
)

//go:generate mockery --name Service  --output=./mocks --filename service.go --quiet --note "Copyright (c) Abstract Machines"
type Service interface {
	// Identify returns thing ID for given thing key.
	Identify(ctx context.Context, key string) (string, error)

	// Authorize used for Things authorization.
	Authorize(ctx context.Context, req things.AuthzReq) (string, error)

	RetrieveById(ctx context.Context, id string) (clients.Client, error)

	RetrieveByIds(ctx context.Context, ids []string) (clients.ClientsPage, error)

	AddConnections(ctx context.Context, conns []things.Connection) error

	RemoveConnections(ctx context.Context, conns []things.Connection) error

	RemoveChannelConnections(ctx context.Context, channelID string) error

	UnsetParentGroupFormThings(ctx context.Context, parentGroupID string) error
}

var _ Service = (*service)(nil)

func New(repo things.Repository, cache things.Cache, evaluator policies.Evaluator, policy policies.Service) Service {
	return service{
		repo:      repo,
		cache:     cache,
		evaluator: evaluator,
		policy:    policy,
	}
}

type service struct {
	repo      things.Repository
	cache     things.Cache
	evaluator policies.Evaluator
	policy    policies.Service
}

func (svc service) Identify(ctx context.Context, key string) (string, error) {
	id, err := svc.cache.ID(ctx, key)
	if err == nil {
		return id, nil
	}

	client, err := svc.repo.RetrieveBySecret(ctx, key)
	if err != nil {
		return "", errors.Wrap(svcerr.ErrAuthorization, err)
	}
	if err := svc.cache.Save(ctx, key, client.ID); err != nil {
		return "", errors.Wrap(svcerr.ErrAuthorization, err)
	}

	return client.ID, nil
}

func (svc service) Authorize(ctx context.Context, req things.AuthzReq) (string, error) {
	thingID, err := svc.Identify(ctx, req.ThingKey)
	if err != nil {
		return "", err
	}

	r := policies.Policy{
		SubjectType: policies.GroupType,
		Subject:     req.ChannelID,
		ObjectType:  policies.ThingType,
		Object:      thingID,
		Permission:  req.Permission,
	}
	err = svc.evaluator.CheckPolicy(ctx, r)
	if err != nil {
		return "", errors.Wrap(svcerr.ErrAuthorization, err)
	}

	return thingID, nil
}

func (svc service) RetrieveById(ctx context.Context, ids string) (clients.Client, error) {
	return svc.repo.RetrieveByID(ctx, ids)
}

func (svc service) RetrieveByIds(ctx context.Context, ids []string) (clients.ClientsPage, error) {
	return svc.repo.RetrieveByIds(ctx, ids)
}

func (svc service) AddConnections(ctx context.Context, conns []things.Connection) (err error) {
	return svc.repo.AddConnections(ctx, conns)
}

func (svc service) RemoveConnections(ctx context.Context, conns []things.Connection) (err error) {
	return svc.repo.RemoveConnections(ctx, conns)
}

func (svc service) RemoveChannelConnections(ctx context.Context, channelID string) error {
	return svc.repo.RemoveChannelConnections(ctx, channelID)
}

func (svc service) UnsetParentGroupFormThings(ctx context.Context, parentGroupID string) (retErr error) {
	ths, err := svc.repo.RetrieveParentGroupThings(ctx, parentGroupID)
	if err != nil {
		return errors.Wrap(svcerr.ErrViewEntity, err)
	}

	if len(ths) > 0 {
		prs := []policies.Policy{}
		for _, th := range ths {
			prs = append(prs, policies.Policy{
				SubjectType: policies.GroupType,
				Subject:     th.ParentGroup,
				Relation:    policies.ParentGroupRelation,
				ObjectType:  policies.ThingType,
				Object:      th.ID,
			})
		}

		if err := svc.policy.DeletePolicies(ctx, prs); err != nil {
			return errors.Wrap(svcerr.ErrDeletePolicies, err)
		}
		defer func() {
			if retErr != nil {
				if errRollback := svc.policy.AddPolicies(ctx, prs); err != nil {
					retErr = errors.Wrap(retErr, errors.Wrap(errors.ErrRollbackTx, errRollback))
				}
			}
		}()

		if err := svc.repo.UnsetParentGroupFormThings(ctx, parentGroupID); err != nil {
			return errors.Wrap(svcerr.ErrRemoveEntity, err)
		}
	}
	return nil
}
