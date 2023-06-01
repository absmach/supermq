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
	idProvider  mainflux.IDProvider
}

// NewService returns a new Clients service implementation.
func NewService(auth upolicies.AuthServiceClient, t mfclients.Repository, p Repository, tcache clients.ClientCache, ccache Cache, idp mainflux.IDProvider) Service {
	return service{
		auth:        auth,
		things:      t,
		policies:    p,
		policyCache: ccache,
		idProvider:  idp,
	}
}

func (svc service) Authorize(ctx context.Context, ar AccessRequest, entity string, p Policy) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if ar.Subject == p.Subject && ar.Object == p.Object {
		for _, a := range p.Actions {
			if a == ar.Action {
				return nil
			}
		}
	}

	return errors.New("unauthorized")
}

func (svc service) AuthorizeByKey(ctx context.Context, ar AccessRequest, entity string) (string, error) {
	// Fetch policy from cache...
	p := Policy{
		Subject: ar.Subject,
		Object:  ar.Object,
	}
	policy, err := svc.policyCache.Get(ctx, p)
	if err == nil {
		if err := svc.Authorize(ctx, ar, "thing", policy); err != nil {
			return "", err
		}
		return ar.Subject, nil
	}
	if !errors.Contains(err, errors.ErrNotFound) {
		return "", err
	}
	// and fallback to repo if policy is not found in cache.
	policy, err = svc.policies.RetrieveOne(ctx, p.Subject, p.Object)
	if err != nil {
		return "", err
	}
	// Replace Subject since AccessRequest Subject is Thing Key,
	// and Policy subject is Thing ID.
	policy.Subject = ar.Subject
	if err := svc.Authorize(ctx, ar, "thing", policy); err != nil {
		return "", err
	}
	if err := svc.policyCache.Put(ctx, policy); err != nil {
		return policy.Subject, errors.Wrap(errors.New("failed to store to cache"), err)
	}

	return policy.Subject, nil
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

	if err := svc.policyCache.Remove(ctx, p); err != nil {
		return err
	}
	return svc.policies.Delete(ctx, p)
}

// checkActionRank check if client updating the policy has the sufficient privileges
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
