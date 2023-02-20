package policies

import (
	"context"
	"time"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/pkg/errors"
	upolicies "github.com/mainflux/mainflux/users/policies"
)

// Possible token types are access and refresh tokens.
const (
	AccessToken = "access"
)

// Service unites Clients and Group services.
type Service interface {
	PolicyService
}

type service struct {
	auth         upolicies.AuthServiceClient
	policies     PolicyRepository
	channelCache ChannelCache
	thingCache   ThingCache
	idProvider   mainflux.IDProvider
}

// NewService returns a new Clients service implementation.
func NewService(auth upolicies.AuthServiceClient, p PolicyRepository, tcache ThingCache, ccache ChannelCache, idp mainflux.IDProvider) Service {
	return service{
		auth:         auth,
		policies:     p,
		thingCache:   tcache,
		channelCache: ccache,
		idProvider:   idp,
	}
}

func (svc service) Authorize(ctx context.Context, entityType string, p Policy) error {
	if err := p.Validate(); err != nil {
		return err
	}
	res, err := svc.auth.Identify(ctx, &upolicies.Token{Value: p.Subject})
	if err != nil {
		return errors.Wrap(errors.ErrAuthentication, err)
	}
	p.Subject = res.GetId()
	return svc.policies.Evaluate(ctx, entityType, p)
}
func (svc service) UpdatePolicy(ctx context.Context, token string, p Policy) error {
	res, err := svc.auth.Identify(ctx, &upolicies.Token{Value: token})
	if err != nil {
		return errors.Wrap(errors.ErrAuthentication, err)
	}
	if err := p.Validate(); err != nil {
		return err
	}
	if err := svc.checkActionRank(ctx, res.GetId(), p); err != nil {
		return err
	}
	p.UpdatedAt = time.Now()

	return svc.policies.Update(ctx, p)
}

func (svc service) AddPolicy(ctx context.Context, token string, p Policy) error {
	res, err := svc.auth.Identify(ctx, &upolicies.Token{Value: token})
	if err != nil {
		return errors.Wrap(errors.ErrAuthentication, err)
	}

	if err := p.Validate(); err != nil {
		return err
	}

	p.OwnerID = res.GetId()
	p.CreatedAt = time.Now()
	p.UpdatedAt = p.CreatedAt

	return svc.policies.Save(ctx, p)
}

func (svc service) DeletePolicy(ctx context.Context, token string, p Policy) error {
	if _, err := svc.auth.Identify(ctx, &upolicies.Token{Value: token}); err != nil {
		return errors.Wrap(errors.ErrAuthentication, err)
	}

	if err := svc.channelCache.Disconnect(ctx, p.Object, p.Subject); err != nil {
		return err
	}

	return svc.policies.Delete(ctx, p)
}

func (svc service) ListPolicy(ctx context.Context, token string, pm Page) (PolicyPage, error) {
	if _, err := svc.auth.Identify(ctx, &upolicies.Token{Value: token}); err != nil {
		return PolicyPage{}, errors.Wrap(errors.ErrAuthentication, err)
	}
	if err := pm.Validate(); err != nil {
		return PolicyPage{}, err
	}

	page, err := svc.policies.Retrieve(ctx, pm)
	if err != nil {
		return PolicyPage{}, err
	}

	return page, err
}

// checkActionRank check if an action is in the provide list of actions
func (svc service) checkActionRank(ctx context.Context, clientID string, p Policy) error {
	page, err := svc.policies.Retrieve(ctx, Page{Subject: clientID, Object: p.Object})
	if err != nil {
		return err
	}
	if len(page.Policies) != 0 {
		for _, a := range p.Actions {
			var found = false
			for _, v := range page.Policies[0].Actions {
				if v == a {
					found = true
					break
				}
			}
			if !found {
				return apiutil.ErrHigherPolicyRank
			}
		}
	}

	return nil

}

func (svc service) CanAccessByKey(ctx context.Context, chanID, thingKey string) (string, error) {
	thingID, err := svc.hasThing(ctx, chanID, thingKey)
	if err == nil {
		return thingID, nil
	}

	thingID, err = svc.policies.HasThing(ctx, chanID, thingKey)
	if err != nil {
		return "", err
	}

	if err := svc.thingCache.Save(ctx, thingKey, thingID); err != nil {
		return "", err
	}
	if err := svc.channelCache.Connect(ctx, chanID, thingID); err != nil {
		return "", err
	}
	return thingID, nil
}

func (svc service) CanAccessByID(ctx context.Context, chanID, thingID string) error {
	if connected := svc.channelCache.HasThing(ctx, chanID, thingID); connected {
		return nil
	}

	if err := svc.policies.HasThingByID(ctx, chanID, thingID); err != nil {
		return err
	}

	if err := svc.channelCache.Connect(ctx, chanID, thingID); err != nil {
		return err
	}
	return nil
}

func (svc service) hasThing(ctx context.Context, chanID, thingKey string) (string, error) {
	thingID, err := svc.thingCache.ID(ctx, thingKey)
	if err != nil {
		return "", err
	}

	if connected := svc.channelCache.HasThing(ctx, chanID, thingID); !connected {
		return "", errors.ErrAuthorization
	}
	return thingID, nil
}
