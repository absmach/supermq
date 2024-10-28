package private

import (
	"context"

	"github.com/absmach/magistrala/channels"
	"github.com/absmach/magistrala/pkg/errors"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	"github.com/absmach/magistrala/pkg/policies"
)

type Service interface {
	UnsetParentGroupFromChannels(ctx context.Context, parentGroupID string) error

	RemoveThingConnections(ctx context.Context, thingID string) error
}

type service struct {
	repo   channels.Repository
	policy policies.Service
}

var _ Service = (*service)(nil)

func New(repo channels.Repository, policy policies.Service) Service {
	return service{repo, policy}
}

func (svc service) RemoveThingConnections(ctx context.Context, thingID string) error {
	return svc.repo.RemoveThingConnections(ctx, thingID)
}

func (svc service) UnsetParentGroupFromChannels(ctx context.Context, parentGroupID string) (retErr error) {
	chs, err := svc.repo.RetrieveParentGroupChannels(ctx, parentGroupID)
	if err != nil {
		return errors.Wrap(svcerr.ErrViewEntity, err)
	}

	if len(chs) > 0 {
		prs := []policies.Policy{}
		for _, ch := range chs {
			prs = append(prs, policies.Policy{
				SubjectType: policies.GroupType,
				Subject:     ch.ParentGroup,
				Relation:    policies.ParentGroupRelation,
				ObjectType:  policies.ChannelType,
				Object:      ch.ID,
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

		if err := svc.repo.UnsetParentGroupFromChannels(ctx, parentGroupID); err != nil {
			return errors.Wrap(svcerr.ErrRemoveEntity, err)
		}
	}
	return nil
}
