package policies

import (
	"context"
	"time"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/users/jwt"
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
	policies   PolicyRepository
	idProvider mainflux.IDProvider
	tokens     jwt.TokenRepository
}

// NewService returns a new Clients service implementation.
func NewService(p PolicyRepository, t jwt.TokenRepository, idp mainflux.IDProvider) Service {
	return service{
		policies:   p,
		tokens:     t,
		idProvider: idp,
	}
}

func (svc service) Authorize(ctx context.Context, entityType string, p Policy) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if err := svc.policies.CheckAdmin(ctx, p.Subject); err == nil {
		return nil
	}

	return svc.policies.Evaluate(ctx, entityType, p)
}

func (svc service) UpdatePolicy(ctx context.Context, token string, p Policy) error {
	id, err := svc.identify(ctx, token)
	if err != nil {
		return err
	}
	if err := p.Validate(); err != nil {
		return err
	}
	if err := svc.checkPolicy(ctx, id, p); err != nil {
		return err
	}
	p.UpdatedAt = time.Now()
	p.UpdatedBy = id

	return svc.policies.Update(ctx, p)
}

// AddPolicy adds a policy is added if:
//
//  1. The client is admin
//  2. The client has `g_add` action on the object.
//  3. The client is the owner of both the subject and the object.
func (svc service) AddPolicy(ctx context.Context, token string, p Policy) error {
	id, err := svc.identify(ctx, token)
	if err != nil {
		return err
	}
	if err := p.Validate(); err != nil {
		return err
	}

	pm := Page{Subject: p.Subject, Object: p.Object, Offset: 0, Limit: 1}
	page, err := svc.policies.Retrieve(ctx, pm)
	if err != nil {
		return err
	}

	// If the policy already exists, replace the actions
	if len(page.Policies) == 1 {
		p.UpdatedAt = time.Now()
		p.UpdatedBy = id
		return svc.policies.Update(ctx, p)
	}

	if err := svc.checkPolicy(ctx, id, p); err != nil {
		return err
	}
	p.OwnerID = id
	p.CreatedAt = time.Now()

	// check if the client is admin
	if err = svc.policies.CheckAdmin(ctx, id); err == nil {
		return svc.policies.Save(ctx, p)
	}

	// check if the client has `g_add` action on the object
	pol := Policy{Subject: id, Object: p.Object, Actions: []string{"g_add"}}
	if err := svc.policies.Evaluate(ctx, "group", pol); err == nil {
		return svc.policies.Save(ctx, p)
	}

	// check if the client is the owner of the subject and the object
	if err := svc.policies.CheckClientOwner(ctx, p.Subject, id); err != nil {
		return err
	}
	if err := svc.policies.CheckGroupOwner(ctx, p.Object, id); err != nil {
		return err
	}

	return svc.policies.Save(ctx, p)
}

func (svc service) DeletePolicy(ctx context.Context, token string, p Policy) error {
	id, err := svc.identify(ctx, token)
	if err != nil {
		return err
	}
	if err := svc.checkPolicy(ctx, id, p); err != nil {
		return err
	}

	return svc.policies.Delete(ctx, p)
}

func (svc service) ListPolicy(ctx context.Context, token string, pm Page) (PolicyPage, error) {
	id, err := svc.identify(ctx, token)
	if err != nil {
		return PolicyPage{}, err
	}
	if err := pm.Validate(); err != nil {
		return PolicyPage{}, err
	}
	// If the user is admin, return all policies
	if err := svc.policies.CheckAdmin(ctx, id); err == nil {
		return svc.policies.Retrieve(ctx, pm)
	}

	// If the user is not admin, return only the policies that they are in
	pm.Subject = id
	pm.Object = id

	return svc.policies.Retrieve(ctx, pm)
}

// checkPolicy checks for the following:
//
//  1. Check if the client is admin
//  2. Check if the client is the owner of the policy
func (svc service) checkPolicy(ctx context.Context, clientID string, p Policy) error {
	// Check if the client is admin
	if err := svc.policies.CheckAdmin(ctx, clientID); err == nil {
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

func (svc service) identify(ctx context.Context, tkn string) (string, error) {
	claims, err := svc.tokens.Parse(ctx, tkn)
	if err != nil {
		return "", errors.Wrap(errors.ErrAuthentication, err)
	}
	if claims.Type != AccessToken {
		return "", errors.ErrAuthentication
	}

	return claims.ClientID, nil
}
