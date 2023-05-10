package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/internal/apiutil"
	mfclients "github.com/mainflux/mainflux/pkg/clients"
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
	entityType        = "group"
)

var AdminRelationKey = []string{createKey, updateRelationKey, listRelationKey, deleteRelationKey}

type service struct {
	auth        policies.AuthServiceClient
	clients     mfclients.Repository
	clientCache ClientCache
	idProvider  mainflux.IDProvider
}

// NewService returns a new Clients service implementation.
func NewService(auth policies.AuthServiceClient, c mfclients.Repository, tcache ClientCache, idp mainflux.IDProvider) Service {
	return service{
		auth:        auth,
		clients:     c,
		clientCache: tcache,
		idProvider:  idp,
	}
}

func (svc service) CreateThings(ctx context.Context, token string, clis ...mfclients.Client) ([]mfclients.Client, error) {
	res, err := svc.auth.Identify(ctx, &policies.Token{Value: token})
	if err != nil {
		return []mfclients.Client{}, err
	}
	var clients []mfclients.Client
	for _, cli := range clis {
		if cli.ID == "" {
			clientID, err := svc.idProvider.ID()
			if err != nil {
				return []mfclients.Client{}, err
			}
			cli.ID = clientID
		}
		if cli.Credentials.Secret == "" {
			key, err := svc.idProvider.ID()
			if err != nil {
				return []mfclients.Client{}, err
			}
			cli.Credentials.Secret = key
		}
		if cli.Owner == "" {
			cli.Owner = res.GetId()
		}
		if cli.Status != mfclients.DisabledStatus && cli.Status != mfclients.EnabledStatus {
			return []mfclients.Client{}, apiutil.ErrInvalidStatus
		}
		cli.CreatedAt = time.Now()
		cli.UpdatedAt = cli.CreatedAt
		cli.UpdatedBy = cli.Owner
		clients = append(clients, cli)
	}
	return svc.clients.Save(ctx, clients...)
}

func (svc service) ViewClient(ctx context.Context, token string, id string) (mfclients.Client, error) {
	if err := svc.authorize(ctx, token, id, listRelationKey); err != nil {
		return mfclients.Client{}, errors.Wrap(errors.ErrNotFound, err)
	}
	return svc.clients.RetrieveByID(ctx, id)
}

func (svc service) ListClients(ctx context.Context, token string, pm mfclients.Page) (mfclients.ClientsPage, error) {
	res, err := svc.auth.Identify(ctx, &policies.Token{Value: token})
	if err != nil {
		return mfclients.ClientsPage{}, errors.Wrap(errors.ErrAuthentication, err)
	}

	// If the user is admin, fetch all things from database.
	if err := svc.authorize(ctx, token, thingsObjectKey, listRelationKey); err == nil {
		pm.Owner = ""
		pm.SharedBy = ""
		return svc.clients.RetrieveAll(ctx, pm)
	}

	// If the user is not admin, check 'sharedby' parameter from page metadata.
	// If user provides 'sharedby' key, fetch things from policies. Otherwise,
	// fetch things from the database based on thing's 'owner' field.
	if pm.SharedBy == MyKey {
		pm.SharedBy = res.GetId()
	}
	if pm.Owner == MyKey {
		pm.Owner = res.GetId()
	}
	pm.Action = "c_list"

	return svc.clients.RetrieveAll(ctx, pm)
}

func (svc service) UpdateClient(ctx context.Context, token string, cli mfclients.Client) (mfclients.Client, error) {
	userID, err := svc.identifyUser(ctx, token)
	if err != nil {
		return mfclients.Client{}, err
	}
	if err := svc.authorize(ctx, token, cli.ID, updateRelationKey); err != nil {
		return mfclients.Client{}, err
	}

	client := mfclients.Client{
		ID:        cli.ID,
		Name:      cli.Name,
		Metadata:  cli.Metadata,
		UpdatedAt: time.Now(),
		UpdatedBy: userID,
	}

	return svc.clients.Update(ctx, client)
}

func (svc service) UpdateClientTags(ctx context.Context, token string, cli mfclients.Client) (mfclients.Client, error) {
	userID, err := svc.identifyUser(ctx, token)
	if err != nil {
		return mfclients.Client{}, err
	}
	if err := svc.authorize(ctx, token, cli.ID, updateRelationKey); err != nil {
		return mfclients.Client{}, err
	}

	client := mfclients.Client{
		ID:        cli.ID,
		Tags:      cli.Tags,
		UpdatedAt: time.Now(),
		UpdatedBy: userID,
	}

	return svc.clients.UpdateTags(ctx, client)
}

func (svc service) UpdateClientSecret(ctx context.Context, token, id, key string) (mfclients.Client, error) {
	userID, err := svc.identifyUser(ctx, token)
	if err != nil {
		return mfclients.Client{}, err
	}
	if err := svc.authorize(ctx, token, id, updateRelationKey); err != nil {
		return mfclients.Client{}, err
	}

	client := mfclients.Client{
		ID: id,
		Credentials: mfclients.Credentials{
			Secret: key,
		},
		UpdatedAt: time.Now(),
		UpdatedBy: userID,
	}

	return svc.clients.UpdateSecret(ctx, client)
}

func (svc service) UpdateClientOwner(ctx context.Context, token string, cli mfclients.Client) (mfclients.Client, error) {
	userID, err := svc.identifyUser(ctx, token)
	if err != nil {
		return mfclients.Client{}, err
	}
	if err := svc.authorize(ctx, token, cli.ID, updateRelationKey); err != nil {
		return mfclients.Client{}, err
	}

	client := mfclients.Client{
		ID:        cli.ID,
		Owner:     cli.Owner,
		UpdatedAt: time.Now(),
		UpdatedBy: userID,
	}

	return svc.clients.UpdateOwner(ctx, client)
}

func (svc service) EnableClient(ctx context.Context, token, id string) (mfclients.Client, error) {
	client := mfclients.Client{
		ID:        id,
		Status:    mfclients.EnabledStatus,
		UpdatedAt: time.Now(),
	}
	client, err := svc.changeClientStatus(ctx, token, client)
	if err != nil {
		return mfclients.Client{}, errors.Wrap(mfclients.ErrEnableClient, err)
	}

	return client, nil
}

func (svc service) DisableClient(ctx context.Context, token, id string) (mfclients.Client, error) {
	client := mfclients.Client{
		ID:        id,
		Status:    mfclients.DisabledStatus,
		UpdatedAt: time.Now(),
	}
	client, err := svc.changeClientStatus(ctx, token, client)
	if err != nil {
		return mfclients.Client{}, errors.Wrap(mfclients.ErrDisableClient, err)
	}

	return client, nil
}

func (svc service) ListClientsByGroup(ctx context.Context, token, groupID string, pm mfclients.Page) (mfclients.MembersPage, error) {
	userID, err := svc.identifyUser(ctx, token)
	if err != nil {
		return mfclients.MembersPage{}, err
	}
	// If the user is admin, fetch all things connected to the channel.
	if err := svc.authorize(ctx, token, thingsObjectKey, listRelationKey); err == nil {
		return svc.clients.Members(ctx, groupID, pm)
	}
	pm.Owner = userID

	return svc.clients.Members(ctx, groupID, pm)
}

func (svc service) Identify(ctx context.Context, key string) (string, error) {
	id, err := svc.clientCache.ID(ctx, key)
	if err == nil {
		return id, nil
	}
	client, err := svc.clients.RetrieveBySecret(ctx, key)
	if err != nil {
		return "", err
	}
	if err := svc.clientCache.Save(ctx, key, client.ID); err != nil {
		return "", err
	}
	return client.ID, nil
}

func (svc service) changeClientStatus(ctx context.Context, token string, client mfclients.Client) (mfclients.Client, error) {
	userID, err := svc.identifyUser(ctx, token)
	if err != nil {
		return mfclients.Client{}, err
	}
	if err := svc.authorize(ctx, token, client.ID, deleteRelationKey); err != nil {
		return mfclients.Client{}, err
	}
	dbClient, err := svc.clients.RetrieveByID(ctx, client.ID)
	if err != nil {
		return mfclients.Client{}, err
	}
	if dbClient.Status == client.Status {
		return mfclients.Client{}, mfclients.ErrStatusAlreadyAssigned
	}
	client.UpdatedBy = userID
	return svc.clients.ChangeStatus(ctx, client)
}

func (svc service) identifyUser(ctx context.Context, token string) (string, error) {
	req := &policies.Token{Value: token}
	res, err := svc.auth.Identify(ctx, req)
	if err != nil {
		return "", errors.Wrap(errors.ErrAuthorization, err)
	}
	return res.GetId(), nil
}

func (svc service) authorize(ctx context.Context, subject, object string, relation string) error {
	// Check if the client is the owner of the thing.
	userID, err := svc.identifyUser(ctx, subject)
	if err != nil {
		return err
	}
	dbThing, err := svc.clients.RetrieveByID(ctx, object)
	if err != nil {
		return err
	}
	if dbThing.Owner == userID {
		return nil
	}
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

func (svc service) ShareClient(ctx context.Context, token, clientID string, actions, userIDs []string) error {
	if err := svc.authorize(ctx, token, clientID, updateRelationKey); err != nil {
		return err
	}
	var errs error
	for _, userID := range userIDs {
		apr, err := svc.auth.AddPolicy(ctx, &policies.AddPolicyReq{Token: token, Sub: userID, Obj: clientID, Act: actions})
		if err != nil {
			errs = errors.Wrap(fmt.Errorf("cannot claim ownership on object '%s' by user '%s': %s", clientID, userID, err), errs)
		}
		if !apr.GetAuthorized() {
			errs = errors.Wrap(fmt.Errorf("cannot claim ownership on object '%s' by user '%s': unauthorized", clientID, userID), errs)
		}
	}
	return errs
}
