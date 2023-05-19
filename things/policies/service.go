package policies

import (
	"context"
	"time"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/things/clients"
	upolicies "github.com/mainflux/mainflux/users/policies"
)

type ClientType int

const (
	UserClientType = iota
	ThingClinetType

	ReadAction       = "m_read"
	WriteAction      = "m_write"
	ClientEntityType = "client"
	GroupEntityType  = "group"
)

type service struct {
	auth        upolicies.AuthServiceClient
	policies    Repository
	policyCache Cache
	idProvider  mainflux.IDProvider
}

// NewService returns a new Clients service implementation.
func NewService(auth upolicies.AuthServiceClient, p Repository, tcache clients.ClientCache, ccache Cache, idp mainflux.IDProvider) Service {
	return service{
		auth:        auth,
		policies:    p,
		policyCache: ccache,
		idProvider:  idp,
	}
}

func (svc service) Authorize(ctx context.Context, ar AccessRequest, entity string) (string, error) {
	// fetch from cache first
	p := Policy{
		Subject: ar.Subject,
		Object:  ar.Object,
	}
	policy, err := svc.policyCache.Get(ctx, p)
	if err == nil {
		for _, action := range policy.Actions {
			if action == ar.Action {
				return policy.Subject, nil
			}
		}
		return "", errors.ErrAuthorization
	}
	if !errors.Contains(err, errors.ErrNotFound) {
		return "", err
	}
	// fetch from repo as a fallback if not found in cache
	policy, err = svc.policies.RetrieveOne(ctx, p.Subject, p.Object)
	if err != nil {
		return "", err
	}

	// Replace Subject since AccessRequest Subject is Thing Key,
	// and Policy subject is Thing ID.
	policy.Subject = ar.Subject

	for _, action := range policy.Actions {
		if action == ar.Action {
			if err := svc.policyCache.Put(ctx, policy); err != nil {
				return policy.Subject, err
			}

			return policy.Subject, nil
		}
	}
	return "", errors.ErrAuthorization

}

// AddPolicy adds a policy is added if:
//
//  1. The client is admin
//
//  2. The client has `g_add` action on the object or is the owner of the object.
func (svc service) AddPolicy(ctx context.Context, token string, clientType ClientType, p Policy) (Policy, error) {
	userID, err := svc.identifyUser(ctx, token)
	if err != nil {
		return Policy{}, err
	}
	if err := p.Validate(); err != nil {
		return Policy{}, err
	}
	pm := Page{Subject: p.Subject, Object: p.Object, Offset: 0, Limit: 1}
	page, err := svc.policies.Retrieve(ctx, pm)
	if err != nil {
		return Policy{}, err
	}

	// If the policy already exists, replace the actions
	if len(page.Policies) == 1 {
		p.UpdatedAt = time.Now()
		p.UpdatedBy = userID
		return svc.policies.Update(ctx, p)
	}

	p.OwnerID = userID
	p.CreatedAt = time.Now()

	if err := svc.checkSubject(ctx, clientType, userID, p); err != nil {
		return Policy{}, err
	}

	// If the client is admin, add the policy
	if err := svc.checkAdmin(ctx, userID); err == nil {
		if err := svc.policyCache.Put(ctx, p); err != nil {
			return Policy{}, err
		}
		return svc.policies.Save(ctx, p)
	}

	// If the client has `g_add` action on the object or is the owner of the object, add the policy
	pol := Policy{Subject: userID, Object: p.Object, Actions: []string{"g_add"}}
	if err := svc.policies.Evaluate(ctx, "group", pol); err == nil {
		if err := svc.policyCache.Put(ctx, p); err != nil {
			return Policy{}, err
		}
		return svc.policies.Save(ctx, p)
	}

	return Policy{}, errors.ErrAuthorization
}

// checkSubject checks if the subject:
//
//  1. If it is a user - the client is authorized to manage it.
//  2. If it is a thing - the client is authorized to manage it.
func (svc service) checkSubject(ctx context.Context, clientType ClientType, userID string, p Policy) error {
	switch clientType {
	case UserClientType:
		req := &upolicies.AuthorizeReq{
			Sub:        userID,
			Obj:        p.Subject,
			Act:        "m_manage",
			EntityType: GroupEntityType,
		}
		res, err := svc.auth.Authorize(ctx, req)
		if err != nil {
			return errors.Wrap(errors.ErrAuthorization, err)
		}
		if !res.GetAuthorized() {
			return errors.ErrAuthorization
		}
		return nil
	case ThingClinetType:
		policy := Policy{Subject: userID, Object: p.Subject, Actions: []string{"m_manage"}}
		if err := svc.Authorize(ctx, GroupEntityType, policy); err != nil {
			return errors.Wrap(errors.ErrAuthorization, err)
		}
		return nil
	default:
		return errors.ErrAuthorization
	}
}
func (svc service) UpdatePolicy(ctx context.Context, token string, p Policy) (Policy, error) {
	userID, err := svc.identifyUser(ctx, token)
	if err != nil {
		return Policy{}, err
	}

	if err := p.Validate(); err != nil {
		return Policy{}, err
	}
	if err := svc.checkPolicy(ctx, userID, p); err != nil {
		return Policy{}, err
	}
	p.UpdatedAt = time.Now()
	p.UpdatedBy = userID

	return svc.policies.Update(ctx, p)
}

func (svc service) ListPolicies(ctx context.Context, token string, pm Page) (PolicyPage, error) {
	userID, err := svc.identifyUser(ctx, token)
	if err != nil {
		return PolicyPage{}, err
	}
	if err := pm.Validate(); err != nil {
		return PolicyPage{}, err
	}
	// If the user is admin, return all policies
	if err := svc.checkAdmin(ctx, userID); err == nil {
		return svc.policies.Retrieve(ctx, pm)
	}

	// If the user is not admin, return only the policies that they are in
	pm.Subject = userID
	pm.Object = userID

	return svc.policies.Retrieve(ctx, pm)
}

func (svc service) DeletePolicy(ctx context.Context, token string, p Policy) error {
	userID, err := svc.identifyUser(ctx, token)
	if err != nil {
		return err
	}
	if err := svc.checkPolicy(ctx, userID, p); err != nil {
		return err
	}
	if err := svc.checkAction(ctx, res.GetId(), p); err != nil {
		return err
	}
	if err := svc.policyCache.Remove(ctx, p); err != nil {
		return err
	}
	return svc.policies.Delete(ctx, p)
}

// checkPolicy checks for the following:
//
//  1. Check if the client is admin
//  2. Check if the client is the owner of the policy
func (svc service) checkPolicy(ctx context.Context, clientID string, p Policy) error {
	// Check if the client is admin
	if err := svc.checkAdmin(ctx, clientID); err == nil {
		return nil
	}

	// Check if the client is the owner of the policy
	pm := Page{Subject: p.Subject, Object: p.Object, OwnerID: clientID, Offset: 0, Limit: 1}
	page, err := svc.policies.Retrieve(ctx, pm)
	if err != nil {
		return err
	}
	if len(page.Policies) == 1 && page.Policies[0].OwnerID == clientID {
		return nil
	}

	return errors.ErrAuthorization
}

func (svc service) identifyUser(ctx context.Context, token string) (string, error) {
	req := &upolicies.Token{Value: token}
	res, err := svc.auth.Identify(ctx, req)
	if err != nil {
		return "", errors.Wrap(errors.ErrAuthorization, err)
	}
	return res.GetId(), nil
}

func (svc service) checkAdmin(ctx context.Context, id string) error {
	req := &upolicies.AuthorizeReq{
		Sub:        id,
		Obj:        "object",
		Act:        "action",
		EntityType: GroupEntityType,
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
