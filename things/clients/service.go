package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/things/policies"
)

const (
	MyKey             = "mine"
	usersObjectKey    = "users"
	authoritiesObject = "authorities"
	memberRelationKey = "member"
	readRelationKey   = "read"
	writeRelationKey  = "write"
	deleteRelationKey = "delete"
)

var (
	// ErrInvalidStatus indicates invalid status.
	ErrInvalidStatus = errors.New("invalid client status")

	// ErrEnableClient indicates error in enabling client.
	ErrEnableClient = errors.New("failed to enable client")

	// ErrDisableClient indicates error in disabling client.
	ErrDisableClient = errors.New("failed to disable client")

	// ErrStatusAlreadyAssigned indicated that the client or group has already been assigned the status.
	ErrStatusAlreadyAssigned = errors.New("status already assigned")
)

// Service unites Clients and Group services.
type Service interface {
	ClientService
}

type service struct {
	auth       mainflux.AuthServiceClient
	clients    ClientRepository
	thingCache ThingCache
	policies   policies.PolicyRepository
	idProvider mainflux.IDProvider
}

// NewService returns a new Clients service implementation.
func NewService(auth mainflux.AuthServiceClient, c ClientRepository, tcache ThingCache, p policies.PolicyRepository, idp mainflux.IDProvider) Service {
	return service{
		auth:       auth,
		clients:    c,
		thingCache: tcache,
		policies:   p,
		idProvider: idp,
	}
}

func (svc service) CreateThings(ctx context.Context, token string, clis ...Client) ([]Client, error) {
	res, err := svc.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return []Client{}, err
	}
	if err := svc.authorize(ctx, res.GetId(), usersObjectKey, memberRelationKey); err != nil {
		return []Client{}, err
	}
	var clients []Client
	for _, cli := range clis {
		if cli.ID == "" {
			clientID, err := svc.idProvider.ID()
			if err != nil {
				return []Client{}, err
			}
			cli.ID = clientID
		}
		if cli.Credentials.Secret == "" {
			key, err := svc.idProvider.ID()
			if err != nil {
				return []Client{}, err
			}
			cli.Credentials.Secret = key
		}

		if cli.Owner == "" {
			cli.Owner = res.Email
		}
		if cli.Status != DisabledStatus && cli.Status != EnabledStatus {
			return []Client{}, apiutil.ErrInvalidStatus
		}

		cli.CreatedAt = time.Now()
		cli.UpdatedAt = cli.CreatedAt

		c, err := svc.clients.Save(ctx, cli)
		if err != nil {
			return []Client{}, err
		}
		clients = append(clients, c)
	}
	return clients, nil
}

func (svc service) ViewClient(ctx context.Context, token string, id string) (Client, error) {
	res, err := svc.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Client{}, errors.Wrap(errors.ErrAuthentication, err)
	}

	if err := svc.authorize(ctx, res.GetId(), id, readRelationKey); err != nil {
		if err := svc.authorize(ctx, res.GetId(), authoritiesObject, memberRelationKey); err != nil {
			return Client{}, errors.Wrap(errors.ErrNotFound, err)
		}
	}

	return svc.clients.RetrieveByID(ctx, id)
}

func (svc service) ListClients(ctx context.Context, token string, pm Page) (ClientsPage, error) {
	res, err := svc.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return ClientsPage{}, errors.Wrap(errors.ErrAuthentication, err)
	}

	subject := res.GetId()
	// If the user is admin, fetch all things from database.
	if err := svc.authorize(ctx, res.GetId(), authoritiesObject, memberRelationKey); err == nil {
		pm.FetchSharedThings = true
		page, err := svc.clients.RetrieveAll(ctx, pm)
		if err != nil {
			return ClientsPage{}, err
		}
		return page, err
	}

	// If the user is not admin, check 'shared' parameter from page metadata.
	// If user provides 'shared' key, fetch things from policies. Otherwise,
	// fetch things from the database based on thing's 'owner' field.
	if pm.FetchSharedThings {
		req := &mainflux.ListPoliciesReq{Act: "read", Sub: subject}
		lpr, err := svc.auth.ListPolicies(ctx, req)
		if err != nil {
			return ClientsPage{}, err
		}

		var page ClientsPage
		for _, thingID := range lpr.Policies {
			page.Clients = append(page.Clients, Client{ID: thingID})
		}
		return page, nil
	}

	if pm.SharedBy == MyKey {
		pm.SharedBy = subject
	}
	if pm.OwnerID == MyKey {
		pm.OwnerID = subject
	}
	pm.Action = "c_list"
	pm.OwnerID = res.GetEmail()

	return svc.clients.RetrieveAll(ctx, pm)
}

func (svc service) UpdateClient(ctx context.Context, token string, cli Client) (Client, error) {
	res, err := svc.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Client{}, err
	}

	if err := svc.authorize(ctx, res.GetId(), cli.ID, writeRelationKey); err != nil {
		if err := svc.authorize(ctx, res.GetId(), authoritiesObject, memberRelationKey); err != nil {
			return Client{}, err
		}
	}

	client := Client{
		ID:        cli.ID,
		Name:      cli.Name,
		Metadata:  cli.Metadata,
		UpdatedAt: time.Now(),
	}

	return svc.clients.Update(ctx, client)
}

func (svc service) UpdateClientTags(ctx context.Context, token string, cli Client) (Client, error) {
	res, err := svc.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Client{}, err
	}

	if err := svc.authorize(ctx, res.GetId(), cli.ID, writeRelationKey); err != nil {
		if err := svc.authorize(ctx, res.GetId(), authoritiesObject, memberRelationKey); err != nil {
			return Client{}, err
		}
	}

	client := Client{
		ID:        cli.ID,
		Tags:      cli.Tags,
		UpdatedAt: time.Now(),
	}

	return svc.clients.UpdateTags(ctx, client)
}

func (svc service) UpdateClientSecret(ctx context.Context, token, id, key string) (Client, error) {
	res, err := svc.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Client{}, errors.Wrap(errors.ErrAuthentication, err)
	}

	if err := svc.authorize(ctx, res.GetId(), id, writeRelationKey); err != nil {
		if err := svc.authorize(ctx, res.GetId(), authoritiesObject, memberRelationKey); err != nil {
			return Client{}, errors.Wrap(errors.ErrNotFound, err)
		}
	}

	dbClient, err := svc.clients.RetrieveByID(ctx, id)
	if err != nil {
		return Client{}, err
	}
	if dbClient.Owner != res.GetEmail() {
		return Client{}, err
	}
	dbClient.Credentials.Secret = key
	return svc.clients.UpdateSecret(ctx, dbClient)
}

func (svc service) UpdateClientOwner(ctx context.Context, token string, cli Client) (Client, error) {
	res, err := svc.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Client{}, err
	}

	if err := svc.authorize(ctx, res.GetId(), cli.ID, writeRelationKey); err != nil {
		if err := svc.authorize(ctx, res.GetId(), authoritiesObject, memberRelationKey); err != nil {
			return Client{}, err
		}
	}

	client := Client{
		ID:        cli.ID,
		Owner:     cli.Owner,
		UpdatedAt: time.Now(),
	}

	return svc.clients.UpdateOwner(ctx, client)
}

func (svc service) EnableClient(ctx context.Context, token, id string) (Client, error) {
	res, err := svc.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Client{}, errors.Wrap(errors.ErrAuthentication, err)

	}

	if err := svc.authorize(ctx, res.GetId(), id, deleteRelationKey); err != nil {
		if err := svc.authorize(ctx, res.GetId(), authoritiesObject, memberRelationKey); err != nil {
			return Client{}, errors.Wrap(errors.ErrNotFound, err)
		}
	}

	client, err := svc.changeClientStatus(ctx, id, EnabledStatus)
	if err != nil {
		return Client{}, errors.Wrap(ErrEnableClient, err)
	}

	return client, nil
}

func (svc service) DisableClient(ctx context.Context, token, id string) (Client, error) {
	res, err := svc.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Client{}, errors.Wrap(errors.ErrAuthentication, err)

	}

	if err := svc.authorize(ctx, res.GetId(), id, deleteRelationKey); err != nil {
		if err := svc.authorize(ctx, res.GetId(), authoritiesObject, memberRelationKey); err != nil {
			return Client{}, errors.Wrap(errors.ErrNotFound, err)
		}
	}

	client, err := svc.changeClientStatus(ctx, id, DisabledStatus)
	if err != nil {
		return Client{}, errors.Wrap(ErrDisableClient, err)
	}

	return client, nil
}

func (svc service) ListThingsByChannel(ctx context.Context, token, channelID string, pm Page) (MembersPage, error) {
	res, err := svc.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return MembersPage{}, errors.Wrap(errors.ErrAuthentication, err)
	}
	pm.Subject = res.GetId()
	pm.Action = "g_list"

	return svc.clients.Members(ctx, channelID, pm)
}

func (svc service) Identify(ctx context.Context, key string) (string, error) {
	id, err := svc.thingCache.ID(ctx, key)
	if err == nil {
		return id, nil
	}

	client, err := svc.clients.RetrieveBySecret(ctx, key)
	if err != nil {
		return "", err
	}

	if err := svc.thingCache.Save(ctx, key, client.ID); err != nil {
		return "", err
	}
	return client.ID, nil
}

func (svc service) changeClientStatus(ctx context.Context, id string, status Status) (Client, error) {
	dbClient, err := svc.clients.RetrieveByID(ctx, id)
	if err != nil {
		return Client{}, err
	}
	if dbClient.Status == status {
		return Client{}, ErrStatusAlreadyAssigned
	}

	return svc.clients.ChangeStatus(ctx, id, status)
}

func (svc service) authorize(ctx context.Context, subject, object string, relation string) error {
	req := &mainflux.AuthorizeReq{
		Sub: subject,
		Obj: object,
		Act: relation,
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

func (svc service) ShareThing(ctx context.Context, token, thingID string, actions, userIDs []string) error {
	res, err := svc.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return err
	}
	if err := svc.authorize(ctx, res.GetId(), thingID, writeRelationKey); err != nil {
		if err := svc.authorize(ctx, res.GetId(), authoritiesObject, memberRelationKey); err != nil {
			return err
		}
	}
	return svc.claimOwnership(ctx, thingID, actions, userIDs)
}

func (svc service) claimOwnership(ctx context.Context, objectID string, actions, userIDs []string) error {
	var errs error
	for _, userID := range userIDs {
		for _, action := range actions {
			apr, err := svc.auth.AddPolicy(ctx, &mainflux.AddPolicyReq{Obj: objectID, Act: action, Sub: userID})
			if err != nil {
				errs = errors.Wrap(fmt.Errorf("cannot claim ownership on object '%s' by user '%s': %s", objectID, userID, err), errs)
			}
			if !apr.GetAuthorized() {
				errs = errors.Wrap(fmt.Errorf("cannot claim ownership on object '%s' by user '%s': unauthorized", objectID, userID), errs)
			}
		}
	}
	return errs
}
