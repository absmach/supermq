package policies

import (
	"context"
	"time"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/internal/apiutil"
	mfclients "github.com/mainflux/mainflux/pkg/clients"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/things/clients"
	upolicies "github.com/mainflux/mainflux/users/policies"
)

// Possible token types are access and refresh tokens.
const (
	ReadAction       = "m_read"
	WriteAction      = "m_write"
	ClientEntityType = "client"
	GroupEntityType  = "group"
)

type service struct {
	auth        upolicies.AuthServiceClient
	things      mfclients.Repository
	policies    Repository
	policyCache Cache
	thingCache  clients.ClientCache
	idProvider  mainflux.IDProvider
}

// NewService returns a new Clients service implementation.
func NewService(auth upolicies.AuthServiceClient, t mfclients.Repository, p Repository, tcache clients.ClientCache, ccache Cache, idp mainflux.IDProvider) Service {
	return service{
		auth:        auth,
		things:      t,
		policies:    p,
		thingCache:  tcache,
		policyCache: ccache,
		idProvider:  idp,
	}
}

func (svc service) Authorize(ctx context.Context, entityType string, p Policy) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if connected := svc.policyCache.Evaluate(ctx, p); connected {
		return nil
	}
	if err := svc.policies.Evaluate(ctx, entityType, p); err != nil {
		return err
	}
	if err := svc.policyCache.AddPolicy(ctx, p); err != nil {
		return err
	}
	return nil
}

func (svc service) AuthorizeByKey(ctx context.Context, entityType string, p Policy) (string, error) {
	thingID, err := svc.hasThing(ctx, p)
	if err == nil {
		if err := svc.thingCache.Save(ctx, p.Subject, thingID); err != nil {
			return "", err
		}
		return thingID, nil
	}

	if err := svc.Authorize(ctx, entityType, p); err != nil {
		return "", err
	}
	return thingID, nil
}

func (svc service) AddPolicy(ctx context.Context, token string, p Policy) (Policy, error) {
	res, err := svc.auth.Identify(ctx, &upolicies.Token{Value: token})
	if err != nil {
		return Policy{}, errors.Wrap(errors.ErrAuthentication, err)
	}

	if err := p.Validate(); err != nil {
		return Policy{}, err
	}

	p.OwnerID = res.GetId()
	p.CreatedAt = time.Now()
	p.UpdatedAt = p.CreatedAt
	p.UpdatedBy = res.GetId()

	if err := svc.policyCache.AddPolicy(ctx, p); err != nil {
		return Policy{}, err
	}
	return svc.policies.Save(ctx, p)
}

func (svc service) UpdatePolicy(ctx context.Context, token string, p Policy) (Policy, error) {
	res, err := svc.auth.Identify(ctx, &upolicies.Token{Value: token})
	if err != nil {
		return Policy{}, errors.Wrap(errors.ErrAuthentication, err)
	}
	if err := p.Validate(); err != nil {
		return Policy{}, err
	}
	if err := svc.checkActionRank(ctx, res.GetId(), p); err != nil {
		return Policy{}, err
	}
	p.UpdatedAt = time.Now()
	p.UpdatedBy = res.GetId()

	return svc.policies.Update(ctx, p)
}

func (svc service) ListPolicies(ctx context.Context, token string, pm Page) (PolicyPage, error) {
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

func (svc service) DeletePolicy(ctx context.Context, token string, p Policy) error {
	if _, err := svc.auth.Identify(ctx, &upolicies.Token{Value: token}); err != nil {
		return errors.Wrap(errors.ErrAuthentication, err)
	}

	if err := svc.policyCache.DeletePolicy(ctx, p); err != nil {
		return err
	}
	return svc.policies.Delete(ctx, p)
}

// checkActionRank check if client updating the policy has the sufficient priviledges
// WriteAction has a higher priority to ReadAction
func (svc service) checkActionRank(ctx context.Context, clientID string, p Policy) error {
	page, err := svc.policies.Retrieve(ctx, Page{Subject: clientID, Object: p.Object, Total: 1})
	if err != nil {
		return err
	}
	if len(page.Policies) != 0 {
		// Check if the client is the owner
		if page.Policies[0].OwnerID == clientID {
			return nil
		}

		// If I am not the owner I can't add a policy of a higher priority
		for _, act := range page.Policies[0].Actions {
			if act == WriteAction {
				return nil
			}
		}
	}

	return apiutil.ErrHigherPolicyRank

}

func (svc service) hasThing(ctx context.Context, p Policy) (string, error) {
	thingID, err := svc.thingCache.ID(ctx, p.Subject)
	if errors.Contains(err, errors.ErrNotFound) {
		thing, err := svc.things.RetrieveBySecret(ctx, p.Subject)
		if err != nil {
			return "", err
		}
		return thing.ID, nil
	}
	if err != nil {
		return "", err
	}
	p.Subject = thingID
	if connected := svc.policyCache.Evaluate(ctx, p); !connected {
		return "", errors.ErrAuthorization
	}
	return thingID, nil
}
