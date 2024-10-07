// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0
package things

import (
	"context"
	"time"

	"github.com/absmach/magistrala"
	"github.com/absmach/magistrala/auth"
	grpcclient "github.com/absmach/magistrala/auth/api/grpc"
	mgclients "github.com/absmach/magistrala/pkg/clients"
	"github.com/absmach/magistrala/pkg/entityroles"
	"github.com/absmach/magistrala/pkg/errors"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	"github.com/absmach/magistrala/pkg/roles"
	"github.com/absmach/magistrala/pkg/svcutil"
	"github.com/absmach/magistrala/things/postgres"
	"golang.org/x/sync/errgroup"
)

var (
	errCreateThingsPolicies = errors.New("failed to create things policies")
	errRollbackRepo         = errors.New("failed to rollback repo")
)

type identity struct {
	ID       string
	DomainID string
	UserID   string
}
type service struct {
	auth        grpcclient.AuthServiceClient
	policy      magistrala.PolicyServiceClient
	clients     postgres.Repository
	clientCache Cache
	idProvider  magistrala.IDProvider
	opp         svcutil.OperationPerm
	entityroles.RolesSvc
}

// NewService returns a new Clients service implementation.
func NewService(authClient grpcclient.AuthServiceClient, policyClient magistrala.PolicyServiceClient, c postgres.Repository, tcache Cache, idp magistrala.IDProvider, sidProvider magistrala.IDProvider) (Service, error) {

	rolesSvc, err := entityroles.NewRolesSvc(auth.DomainType, c, sidProvider, authClient, policyClient, ThingAvailableActions(), ThingBuiltInRoles(), NewRolesOperationPermissionMap())
	if err != nil {
		return nil, err
	}

	opp := NewOperationPerm()
	if err := opp.AddOperationPermissionMap(NewOperationPermissionMap()); err != nil {
		return service{}, err
	}
	if err := opp.Validate(); err != nil {
		return service{}, err
	}

	return service{
		auth:        authClient,
		policy:      policyClient,
		clients:     c,
		clientCache: tcache,
		idProvider:  idp,
		opp:         opp,
		RolesSvc:    rolesSvc,
	}, nil
}

func (svc service) Authorize(ctx context.Context, req *magistrala.AuthorizeReq) (string, error) {
	thingID, err := svc.Identify(ctx, req.GetSubject())
	if err != nil {
		return "", err
	}

	r := &magistrala.AuthorizeReq{
		SubjectType: auth.GroupType,
		Subject:     req.GetObject(),
		ObjectType:  auth.ThingType,
		Object:      thingID,
		Permission:  req.GetPermission(),
	}
	resp, err := svc.auth.Authorize(ctx, r)
	if err != nil {
		return "", errors.Wrap(svcerr.ErrAuthorization, err)
	}
	if !resp.GetAuthorized() {
		return "", svcerr.ErrAuthorization
	}

	return thingID, nil
}

func (svc service) CreateThings(ctx context.Context, token string, cls ...mgclients.Client) ([]mgclients.Client, error) {
	userInfo, err := svc.identify(ctx, token)
	if err != nil {
		return []mgclients.Client{}, err
	}
	// If domain is disabled , then this authorization will fail for all non-admin domain users
	if _, err := svc.authorize(ctx, "", auth.UserType, auth.UsersKind, userInfo.ID, OpCreateThing, auth.DomainType, userInfo.DomainID); err != nil {
		return []mgclients.Client{}, err
	}

	var clients []mgclients.Client
	for _, c := range cls {
		if c.ID == "" {
			clientID, err := svc.idProvider.ID()
			if err != nil {
				return []mgclients.Client{}, err
			}
			c.ID = clientID
		}
		if c.Credentials.Secret == "" {
			key, err := svc.idProvider.ID()
			if err != nil {
				return []mgclients.Client{}, err
			}
			c.Credentials.Secret = key
		}
		if c.Status != mgclients.DisabledStatus && c.Status != mgclients.EnabledStatus {
			return []mgclients.Client{}, svcerr.ErrInvalidStatus
		}
		c.Domain = userInfo.DomainID
		c.CreatedAt = time.Now()
		clients = append(clients, c)
	}

	saved, err := svc.clients.Save(ctx, clients...)
	if err != nil {
		return nil, errors.Wrap(svcerr.ErrCreateEntity, err)
	}
	clientIDs := []string{}
	for _, c := range saved {
		clientIDs = append(clientIDs, c.ID)
	}

	defer func() {
		if err != nil {
			if errRollBack := svc.clients.RemoveThings(ctx, clientIDs); errRollBack != nil {
				err = errors.Wrap(err, errors.Wrap(errRollbackRepo, errRollBack))
			}
		}
	}()

	newBuiltInRoleMembers := map[roles.BuiltInRoleName][]roles.Member{
		ThingBuiltInRoleAdmin: {roles.Member(userInfo.UserID)},
	}

	optionalPolicies := []roles.OptionalPolicy{}

	for _, clientID := range clientIDs {
		optionalPolicies = append(optionalPolicies,
			roles.OptionalPolicy{
				Namespace:   userInfo.DomainID,
				SubjectType: auth.UserType,
				Subject:     userInfo.ID,
				Relation:    auth.AdministratorRelation,
				ObjectKind:  auth.NewThingKind,
				ObjectType:  auth.ThingType,
				Object:      clientID,
			},
			roles.OptionalPolicy{

				Namespace:   userInfo.DomainID,
				SubjectType: auth.UserType,
				Subject:     userInfo.ID,
				Relation:    auth.DomainRelation,
				ObjectType:  auth.ThingType,
				Object:      clientID,
			},
		)
	}

	if _, err := svc.AddNewEntityRoles(ctx, userInfo.UserID, userInfo.DomainID, userInfo.DomainID, newBuiltInRoleMembers, optionalPolicies); err != nil {
		return []mgclients.Client{}, errors.Wrap(errCreateThingsPolicies, err)
	}

	return saved, nil
}

func (svc service) ViewClient(ctx context.Context, token, id string) (mgclients.Client, error) {
	_, err := svc.authorize(ctx, "", auth.UserType, auth.TokenKind, token, OpViewThing, auth.ThingType, id)
	if err != nil {
		return mgclients.Client{}, err
	}
	client, err := svc.clients.RetrieveByID(ctx, id)
	if err != nil {
		return mgclients.Client{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	return client, nil
}

func (svc service) ListClients(ctx context.Context, token, reqUserID string, pm mgclients.Page) (mgclients.ClientsPage, error) {
	var ids []string

	userInfo, err := svc.identify(ctx, token)
	if err != nil {
		return mgclients.ClientsPage{}, err
	}

	switch {
	case (reqUserID != "" && reqUserID != userInfo.ID):
		// Check user is admin of domain, if yes then show listing on domain context
		if _, err := svc.authorize(ctx, "", auth.UserType, auth.UsersKind, userInfo.ID, OpListThing, auth.DomainType, userInfo.DomainID); err != nil {
			return mgclients.ClientsPage{}, err
		}
		rtids, err := svc.listClientIDs(ctx, auth.EncodeDomainUserID(userInfo.DomainID, reqUserID), pm.Permission)
		if err != nil {
			return mgclients.ClientsPage{}, errors.Wrap(svcerr.ErrNotFound, err)
		}
		ids, err = svc.filterAllowedThingIDs(ctx, userInfo.ID, pm.Permission, rtids)
		if err != nil {
			return mgclients.ClientsPage{}, errors.Wrap(svcerr.ErrNotFound, err)
		}
	default:
		err := svc.checkSuperAdmin(ctx, userInfo.UserID)
		switch {
		case err == nil:
			pm.Domain = userInfo.DomainID
		default:
			// If domain is disabled , then this authorization will fail for all non-admin domain users
			if _, err := svc.authorize(ctx, "", auth.UserType, auth.UsersKind, userInfo.ID, OpListThing, auth.DomainType, userInfo.DomainID); err != nil {
				return mgclients.ClientsPage{}, err
			}
			ids, err = svc.listClientIDs(ctx, userInfo.ID, pm.Permission)
			if err != nil {
				return mgclients.ClientsPage{}, errors.Wrap(svcerr.ErrNotFound, err)
			}
		}
	}

	if len(ids) == 0 && pm.Domain == "" {
		return mgclients.ClientsPage{}, nil
	}
	pm.IDs = ids
	tp, err := svc.clients.SearchClients(ctx, pm)
	if err != nil {
		return mgclients.ClientsPage{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}

	if pm.ListPerms && len(tp.Clients) > 0 {
		g, ctx := errgroup.WithContext(ctx)

		for i := range tp.Clients {
			// Copying loop variable "i" to avoid "loop variable captured by func literal"
			iter := i
			g.Go(func() error {
				return svc.retrievePermissions(ctx, userInfo.ID, &tp.Clients[iter])
			})
		}

		if err := g.Wait(); err != nil {
			return mgclients.ClientsPage{}, err
		}
	}
	return tp, nil
}

// Experimental functions used for async calling of svc.listUserThingPermission. This might be helpful during listing of large number of entities.
func (svc service) retrievePermissions(ctx context.Context, userID string, client *mgclients.Client) error {
	permissions, err := svc.listUserThingPermission(ctx, userID, client.ID)
	if err != nil {
		return err
	}
	client.Permissions = permissions
	return nil
}

func (svc service) listUserThingPermission(ctx context.Context, userID, thingID string) ([]string, error) {
	lp, err := svc.policy.ListPermissions(ctx, &magistrala.ListPermissionsReq{
		SubjectType: auth.UserType,
		Subject:     userID,
		Object:      thingID,
		ObjectType:  auth.ThingType,
	})
	if err != nil {
		return []string{}, errors.Wrap(svcerr.ErrAuthorization, err)
	}
	return lp.GetPermissions(), nil
}

func (svc service) listClientIDs(ctx context.Context, userID, permission string) ([]string, error) {
	tids, err := svc.policy.ListAllObjects(ctx, &magistrala.ListObjectsReq{
		SubjectType: auth.UserType,
		Subject:     userID,
		Permission:  permission,
		ObjectType:  auth.ThingType,
	})
	if err != nil {
		return nil, errors.Wrap(svcerr.ErrNotFound, err)
	}
	return tids.Policies, nil
}

func (svc service) filterAllowedThingIDs(ctx context.Context, userID, permission string, thingIDs []string) ([]string, error) {
	var ids []string
	tids, err := svc.policy.ListAllObjects(ctx, &magistrala.ListObjectsReq{
		SubjectType: auth.UserType,
		Subject:     userID,
		Permission:  permission,
		ObjectType:  auth.ThingType,
	})
	if err != nil {
		return nil, errors.Wrap(svcerr.ErrNotFound, err)
	}
	for _, thingID := range thingIDs {
		for _, tid := range tids.Policies {
			if thingID == tid {
				ids = append(ids, thingID)
			}
		}
	}
	return ids, nil
}

func (svc service) checkSuperAdmin(ctx context.Context, userID string) error {
	res, err := svc.auth.Authorize(ctx, &magistrala.AuthorizeReq{
		SubjectType: auth.UserType,
		Subject:     userID,
		Permission:  auth.AdminPermission,
		ObjectType:  auth.PlatformType,
		Object:      auth.MagistralaObject,
	})
	if err != nil {
		return err
	}
	if !res.Authorized {
		return svcerr.ErrAuthorization
	}
	return nil
}

func (svc service) UpdateClient(ctx context.Context, token string, cli mgclients.Client) (mgclients.Client, error) {
	userID, err := svc.authorize(ctx, "", auth.UserType, auth.TokenKind, token, OpUpdateThing, auth.ThingType, cli.ID)
	if err != nil {
		return mgclients.Client{}, errors.Wrap(svcerr.ErrAuthorization, err)
	}

	client := mgclients.Client{
		ID:        cli.ID,
		Name:      cli.Name,
		Metadata:  cli.Metadata,
		UpdatedAt: time.Now(),
		UpdatedBy: userID,
	}
	client, err = svc.clients.Update(ctx, client)
	if err != nil {
		return mgclients.Client{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}
	return client, nil
}

func (svc service) UpdateClientTags(ctx context.Context, token string, cli mgclients.Client) (mgclients.Client, error) {
	userID, err := svc.authorize(ctx, "", auth.UserType, auth.TokenKind, token, OpUpdateThing, auth.ThingType, cli.ID)
	if err != nil {
		return mgclients.Client{}, errors.Wrap(svcerr.ErrAuthorization, err)
	}

	client := mgclients.Client{
		ID:        cli.ID,
		Tags:      cli.Tags,
		UpdatedAt: time.Now(),
		UpdatedBy: userID,
	}
	client, err = svc.clients.UpdateTags(ctx, client)
	if err != nil {
		return mgclients.Client{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}
	return client, nil
}

func (svc service) UpdateClientSecret(ctx context.Context, token, id, key string) (mgclients.Client, error) {
	userID, err := svc.authorize(ctx, "", auth.UserType, auth.TokenKind, token, OpUpdateThing, auth.ThingType, id)
	if err != nil {
		return mgclients.Client{}, errors.Wrap(svcerr.ErrAuthorization, err)
	}

	client := mgclients.Client{
		ID: id,
		Credentials: mgclients.Credentials{
			Secret: key,
		},
		UpdatedAt: time.Now(),
		UpdatedBy: userID,
		Status:    mgclients.EnabledStatus,
	}
	client, err = svc.clients.UpdateSecret(ctx, client)
	if err != nil {
		return mgclients.Client{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}
	return client, nil
}

func (svc service) EnableClient(ctx context.Context, token, id string) (mgclients.Client, error) {
	userInfo, err := svc.identify(ctx, token)
	if err != nil {
		return mgclients.Client{}, err
	}

	if _, err := svc.authorize(ctx, "", auth.UserType, auth.UsersKind, userInfo.ID, OpEnableThing, auth.ThingType, id); err != nil {
		return mgclients.Client{}, errors.Wrap(svcerr.ErrAuthorization, err)
	}

	client := mgclients.Client{
		ID:        id,
		Status:    mgclients.EnabledStatus,
		UpdatedAt: time.Now(),
	}
	client, err = svc.changeClientStatus(ctx, client, userInfo.ID)
	if err != nil {
		return mgclients.Client{}, errors.Wrap(mgclients.ErrEnableClient, err)
	}

	return client, nil
}

func (svc service) DisableClient(ctx context.Context, token, id string) (mgclients.Client, error) {
	userInfo, err := svc.identify(ctx, token)
	if err != nil {
		return mgclients.Client{}, err
	}

	if _, err := svc.authorize(ctx, "", auth.UserType, auth.UsersKind, userInfo.ID, OpEnableThing, auth.ThingType, id); err != nil {
		return mgclients.Client{}, errors.Wrap(svcerr.ErrAuthorization, err)
	}

	client := mgclients.Client{
		ID:        id,
		Status:    mgclients.DisabledStatus,
		UpdatedAt: time.Now(),
	}
	client, err = svc.changeClientStatus(ctx, client, userInfo.ID)
	if err != nil {
		return mgclients.Client{}, errors.Wrap(mgclients.ErrDisableClient, err)
	}

	if err := svc.clientCache.Remove(ctx, client.ID); err != nil {
		return client, errors.Wrap(svcerr.ErrRemoveEntity, err)
	}

	return client, nil
}

func (svc service) DeleteClient(ctx context.Context, token, id string) error {
	userInfo, err := svc.identify(ctx, token)
	if err != nil {
		return err
	}
	if _, err := svc.authorize(ctx, userInfo.DomainID, auth.UserType, auth.UsersKind, userInfo.ID, OpDeleteThing, auth.ThingType, id); err != nil {
		return err
	}

	if err := svc.clientCache.Remove(ctx, id); err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}

	deleteRes, err := svc.policy.DeleteEntityPolicies(ctx, &magistrala.DeleteEntityPoliciesReq{
		EntityType: auth.ThingType,
		Id:         id,
	})
	if err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}
	if !deleteRes.Deleted {
		return svcerr.ErrAuthorization
	}

	if err := svc.clients.Delete(ctx, id); err != nil {
		return errors.Wrap(svcerr.ErrRemoveEntity, err)
	}

	return nil
}

func (svc service) changeClientStatus(ctx context.Context, client mgclients.Client, userID string) (mgclients.Client, error) {
	dbClient, err := svc.clients.RetrieveByID(ctx, client.ID)
	if err != nil {
		return mgclients.Client{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}
	if dbClient.Status == client.Status {
		return mgclients.Client{}, errors.ErrStatusAlreadyAssigned
	}

	client.UpdatedBy = userID

	client, err = svc.clients.ChangeStatus(ctx, client)
	if err != nil {
		return mgclients.Client{}, errors.Wrap(svcerr.ErrUpdateEntity, err)
	}
	return client, nil
}

func (svc service) Identify(ctx context.Context, key string) (string, error) {
	id, err := svc.clientCache.ID(ctx, key)
	if err == nil {
		return id, nil
	}

	client, err := svc.clients.RetrieveBySecret(ctx, key)
	if err != nil {
		return "", errors.Wrap(svcerr.ErrAuthorization, err)
	}
	if err := svc.clientCache.Save(ctx, key, client.ID); err != nil {
		return "", errors.Wrap(svcerr.ErrAuthorization, err)
	}

	return client.ID, nil
}

func (svc service) identify(ctx context.Context, token string) (identity, error) {
	resp, err := svc.auth.Identify(ctx, &magistrala.IdentityReq{Token: token})
	if err != nil {
		return identity{}, errors.Wrap(svcerr.ErrAuthentication, err)
	}
	return identity{ID: resp.GetId(), DomainID: resp.GetDomainId(), UserID: resp.GetUserId()}, nil
}

func (svc *service) authorize(ctx context.Context, domainID, subjType, subjKind, subj string, op svcutil.Operation, objType, obj string) (string, error) {
	perm, err := svc.opp.GetPermission(op)
	if err != nil {
		return "", err
	}

	req := &magistrala.AuthorizeReq{
		Domain:      domainID,
		SubjectType: subjType,
		SubjectKind: subjKind,
		Subject:     subj,
		Permission:  perm.String(),
		ObjectType:  objType,
		Object:      obj,
	}
	res, err := svc.auth.Authorize(ctx, req)
	if err != nil {
		return "", errors.Wrap(svcerr.ErrAuthorization, err)
	}
	if !res.GetAuthorized() {
		return "", errors.Wrap(svcerr.ErrAuthorization, err)
	}

	return res.GetId(), nil
}
