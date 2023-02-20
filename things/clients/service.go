package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/users/policies"
)

const (
	MyKey             = "mine"
	thingsObjectKey   = "things"
	createKey         = "c_add"
	updateRelationKey = "c_update"
	listRelationKey   = "c_list"
	deleteRelationKey = "c_delete"
	readRelationKey   = "m_read"
	writeRelationKey  = "m_write"
	entityType        = "group"
)

var AdminRelationKey = []string{createKey, updateRelationKey, listRelationKey, deleteRelationKey, readRelationKey, writeRelationKey}

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
	auth       policies.AuthServiceClient
	clients    ClientRepository
	thingCache ThingCache
	idProvider mainflux.IDProvider
}

// NewService returns a new Clients service implementation.
func NewService(auth policies.AuthServiceClient, c ClientRepository, tcache ThingCache, idp mainflux.IDProvider) Service {
	return service{
		auth:       auth,
		clients:    c,
		thingCache: tcache,
		idProvider: idp,
	}
}

func (svc service) CreateThings(ctx context.Context, token string, clis ...Client) ([]Client, error) {
	res, err := svc.auth.Identify(ctx, &policies.Token{Value: token})
	if err != nil {
		return []Client{}, err
	}
	if err := svc.authorize(ctx, token, thingsObjectKey, createKey); err != nil {
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
		clients = append(clients, cli)
	}
	return svc.clients.Save(ctx, clients...)
}

func (svc service) ViewClient(ctx context.Context, token string, id string) (Client, error) {
	if err := svc.authorize(ctx, token, id, listRelationKey); err != nil {
		return Client{}, errors.Wrap(errors.ErrNotFound, err)
	}
	return svc.clients.RetrieveByID(ctx, id)
}

func (svc service) ListClients(ctx context.Context, token string, pm Page) (ClientsPage, error) {
	fmt.Println(pm)
	res, err := svc.auth.Identify(ctx, &policies.Token{Value: token})
	if err != nil {
		return ClientsPage{}, errors.Wrap(errors.ErrAuthentication, err)
	}

	subject := res.GetId()
	// If the user is admin, fetch all things from database.
	if err := svc.authorize(ctx, token, thingsObjectKey, listRelationKey); err == nil {
		pm.OwnerID = ""
		page, err := svc.clients.RetrieveAll(ctx, pm)
		if err != nil {
			return ClientsPage{}, err
		}
		return page, err
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
	if err := svc.authorize(ctx, token, cli.ID, updateRelationKey); err != nil {
		return Client{}, err
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
	if err := svc.authorize(ctx, token, cli.ID, updateRelationKey); err != nil {
		return Client{}, err
	}

	client := Client{
		ID:        cli.ID,
		Tags:      cli.Tags,
		UpdatedAt: time.Now(),
	}

	return svc.clients.UpdateTags(ctx, client)
}

func (svc service) UpdateClientSecret(ctx context.Context, token, id, key string) (Client, error) {
	if err := svc.authorize(ctx, token, id, updateRelationKey); err != nil {
		return Client{}, err
	}

	client := Client{
		ID: id,
		Credentials: Credentials{
			Secret: key,
		},
		UpdatedAt: time.Now(),
	}

	return svc.clients.UpdateSecret(ctx, client)
}

func (svc service) UpdateClientOwner(ctx context.Context, token string, cli Client) (Client, error) {
	if err := svc.authorize(ctx, token, cli.ID, updateRelationKey); err != nil {
		return Client{}, err
	}

	client := Client{
		ID:        cli.ID,
		Owner:     cli.Owner,
		UpdatedAt: time.Now(),
	}

	return svc.clients.UpdateOwner(ctx, client)
}

func (svc service) EnableClient(ctx context.Context, token, id string) (Client, error) {
	client, err := svc.changeClientStatus(ctx, token, id, EnabledStatus)
	if err != nil {
		return Client{}, errors.Wrap(ErrEnableClient, err)
	}

	return client, nil
}

func (svc service) DisableClient(ctx context.Context, token, id string) (Client, error) {
	client, err := svc.changeClientStatus(ctx, token, id, DisabledStatus)
	if err != nil {
		return Client{}, errors.Wrap(ErrDisableClient, err)
	}

	return client, nil
}

func (svc service) ListThingsByChannel(ctx context.Context, token, channelID string, pm Page) (MembersPage, error) {
	res, err := svc.auth.Identify(ctx, &policies.Token{Value: token})
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

func (svc service) changeClientStatus(ctx context.Context, token, id string, status Status) (Client, error) {
	if err := svc.authorize(ctx, token, id, deleteRelationKey); err != nil {
		return Client{}, err
	}
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
	req := &policies.AuthorizeReq{
		Sub:        subject,
		Obj:        object,
		Act:        relation,
		EntityType: entityType,
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
	if err := svc.authorize(ctx, token, thingID, updateRelationKey); err != nil {
		return err
	}
	return svc.claimOwnership(ctx, thingID, AdminRelationKey, userIDs)
}

func (svc service) claimOwnership(ctx context.Context, objectID string, actions, userIDs []string) error {
	var errs error
	for _, userID := range userIDs {
		for _, action := range actions {
			apr, err := svc.auth.AddPolicy(ctx, &policies.AddPolicyReq{Obj: objectID, Act: []string{action}, Sub: userID})
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
