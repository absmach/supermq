package clients

import (
	"context"
	"regexp"
	"time"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/clients/jwt"
	"github.com/mainflux/mainflux/clients/policies"
	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/pkg/errors"
)

const MyKey = "mine"

var (
	// ErrInvalidStatus indicates invalid status.
	ErrInvalidStatus = errors.New("invalid client status")

	// ErrEnableClient indicates error in enabling client.
	ErrEnableClient = errors.New("failed to enable client")

	// ErrDisableClient indicates error in disabling client.
	ErrDisableClient = errors.New("failed to disable client")

	// ErrStatusAlreadyAssigned indicated that the client or group has already been assigned the status.
	ErrStatusAlreadyAssigned = errors.New("status already assigned")

	// ErrMissingResetToken indicates malformed or missing reset token
	// for reseting password.
	ErrMissingResetToken = errors.New("missing reset token")

	// ErrRecoveryToken indicates error in generating password recovery token.
	ErrRecoveryToken = errors.New("failed to generate password recovery token")

	// ErrGetToken indicates error in getting signed token.
	ErrGetToken = errors.New("failed to fetch signed token")

	// ErrPasswordFormat indicates weak password.
	ErrPasswordFormat = errors.New("password does not meet the requirements")
)

// Service unites Clients and Group services.
type Service interface {
	ClientService
	jwt.TokenService
}

type service struct {
	clients    ClientRepository
	policies   policies.PolicyRepository
	idProvider mainflux.IDProvider
	hasher     Hasher
	tokens     jwt.TokenRepository
	email      Emailer
	passRegex  *regexp.Regexp
}

// NewService returns a new Clients service implementation.
func NewService(c ClientRepository, p policies.PolicyRepository, t jwt.TokenRepository, h Hasher, idp mainflux.IDProvider, pr *regexp.Regexp) Service {
	return service{
		clients:    c,
		policies:   p,
		hasher:     h,
		tokens:     t,
		idProvider: idp,
		passRegex:  pr,
	}
}

func (svc service) RegisterClient(ctx context.Context, token string, cli Client) (Client, error) {
	clientID, err := svc.idProvider.ID()
	if err != nil {
		return Client{}, err
	}

	// We don't check the error currently since we can register client with empty token
	owner, _ := svc.Identify(ctx, token)
	if owner.ID != "" && cli.Owner == "" {
		cli.Owner = owner.ID
	}
	if cli.Credentials.Secret == "" {
		return Client{}, apiutil.ErrMissingSecret
	}
	hash, err := svc.hasher.Hash(cli.Credentials.Secret)
	if err != nil {
		return Client{}, errors.Wrap(errors.ErrMalformedEntity, err)
	}
	cli.Credentials.Secret = hash
	if cli.Status != DisabledStatus && cli.Status != EnabledStatus {
		return Client{}, apiutil.ErrInvalidStatus
	}

	cli.ID = clientID
	cli.CreatedAt = time.Now()
	cli.UpdatedAt = cli.CreatedAt

	return svc.clients.Save(ctx, cli)
}

func (svc service) IssueToken(ctx context.Context, identity, secret string) (jwt.Token, error) {
	dbUser, err := svc.clients.RetrieveByIdentity(ctx, identity)
	if err != nil {
		return jwt.Token{}, errors.Wrap(errors.ErrAuthentication, err)
	}
	if err := svc.hasher.Compare(secret, dbUser.Credentials.Secret); err != nil {
		return jwt.Token{}, errors.Wrap(errors.ErrAuthentication, err)
	}

	claims := jwt.Claims{
		ClientID: dbUser.ID,
	}

	return svc.tokens.Issue(ctx, claims)
}

func (svc service) RefreshToken(ctx context.Context, accessToken string) (jwt.Token, error) {
	claims, err := svc.tokens.Parse(ctx, accessToken)
	if err != nil {
		return jwt.Token{}, errors.Wrap(errors.ErrAuthentication, err)
	}
	if claims.Type != jwt.RefreshToken {
		return jwt.Token{}, errors.Wrap(errors.ErrAuthentication, err)
	}
	if _, err := svc.clients.RetrieveByID(ctx, claims.ClientID); err != nil {
		return jwt.Token{}, errors.Wrap(errors.ErrAuthentication, err)
	}

	return svc.tokens.Issue(ctx, claims)
}

func (svc service) ViewClient(ctx context.Context, token string, id string) (Client, error) {
	ir, err := svc.Identify(ctx, token)
	if err != nil {
		return Client{}, err
	}
	if ir.ID == id {
		return svc.clients.RetrieveByID(ctx, id)
	}

	if err := svc.policies.Evaluate(ctx, "client", policies.Policy{Subject: ir.ID, Object: id, Actions: []string{"c_list"}}); err != nil {
		return Client{}, err
	}

	return svc.clients.RetrieveByID(ctx, id)
}

func (svc service) ViewProfile(ctx context.Context, token string) (Client, error) {
	ir, err := svc.Identify(ctx, token)
	if err != nil {
		return Client{}, err
	}
	return svc.clients.RetrieveByID(ctx, ir.ID)
}

func (svc service) ListClients(ctx context.Context, token string, pm Page) (ClientsPage, error) {
	ir, err := svc.Identify(ctx, token)
	if err != nil {
		return ClientsPage{}, err
	}
	if pm.SharedBy == MyKey {
		pm.SharedBy = ir.ID
	}
	if pm.OwnerID == MyKey {
		pm.OwnerID = ir.ID
	}
	pm.Action = "c_list"
	clients, err := svc.clients.RetrieveAll(ctx, pm)
	if err != nil {
		return ClientsPage{}, err
	}
	for i, client := range clients.Clients {
		if client.ID == ir.ID {
			clients.Clients = append(clients.Clients[:i], clients.Clients[i+1:]...)
			if clients.Total != 0 {
				clients.Total = clients.Total - 1
			}
		}
	}

	return clients, nil
}

func (svc service) UpdateClient(ctx context.Context, token string, cli Client) (Client, error) {
	if err := svc.authorize(ctx, "client", policies.Policy{Subject: token, Object: cli.ID, Actions: []string{"c_update"}}); err != nil {
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
	if err := svc.authorize(ctx, "client", policies.Policy{Subject: token, Object: cli.ID, Actions: []string{"c_update"}}); err != nil {
		return Client{}, err
	}

	client := Client{
		ID:        cli.ID,
		Tags:      cli.Tags,
		UpdatedAt: time.Now(),
	}

	return svc.clients.UpdateTags(ctx, client)
}

func (svc service) UpdateClientIdentity(ctx context.Context, token, id, identity string) (Client, error) {
	if err := svc.authorize(ctx, "client", policies.Policy{Subject: token, Object: id, Actions: []string{"c_update"}}); err != nil {
		return Client{}, err
	}

	cli := Client{
		ID: id,
		Credentials: Credentials{
			Identity: identity,
		},
	}
	return svc.clients.UpdateIdentity(ctx, cli)
}

func (svc service) GenerateResetToken(ctx context.Context, email, host string) error {
	client, err := svc.clients.RetrieveByIdentity(ctx, email)
	if err != nil || client.Credentials.Identity == "" {
		return errors.ErrNotFound
	}
	t, err := svc.IssueToken(ctx, client.Credentials.Identity, client.Credentials.Secret)
	if err != nil {
		return errors.Wrap(ErrRecoveryToken, err)
	}
	return svc.SendPasswordReset(ctx, host, email, t.AccessToken)
}

func (svc service) ResetSecret(ctx context.Context, resetToken, secret string) error {
	ir, err := svc.Identify(ctx, resetToken)
	if err != nil {
		return errors.Wrap(errors.ErrAuthentication, err)
	}
	c, err := svc.clients.RetrieveByID(ctx, ir.ID)
	if err != nil {
		return err
	}
	if c.Credentials.Identity == "" {
		return errors.ErrNotFound
	}
	if !svc.passRegex.MatchString(secret) {
		return ErrPasswordFormat
	}
	secret, err = svc.hasher.Hash(secret)
	if err != nil {
		return err
	}
	c = Client{
		Credentials: Credentials{
			Identity: c.Credentials.Identity,
			Secret:   secret,
		},
	}
	if _, err := svc.clients.UpdateSecret(ctx, c); err != nil {
		return err
	}
	return nil
}

func (svc service) UpdateClientSecret(ctx context.Context, token, oldSecret, newSecret string) (Client, error) {
	ir, err := svc.Identify(ctx, token)
	if err != nil {
		return Client{}, err
	}
	// if !svc.passRegex.MatchString(newSecret) {
	// 	return Client{}, ErrPasswordFormat
	// }
	dbClient, err := svc.clients.RetrieveByID(ctx, ir.ID)
	if err != nil {
		return Client{}, err
	}
	if _, err := svc.IssueToken(ctx, dbClient.Credentials.Identity, oldSecret); err != nil {
		return Client{}, errors.ErrAuthentication
	}
	newSecret, err = svc.hasher.Hash(newSecret)
	if err != nil {
		return Client{}, err
	}
	dbClient.Credentials.Secret = newSecret
	return svc.clients.UpdateSecret(ctx, dbClient)
}

func (svc service) SendPasswordReset(_ context.Context, host, email, token string) error {
	to := []string{email}
	return svc.email.SendPasswordReset(to, host, token)
}

func (svc service) UpdateClientOwner(ctx context.Context, token string, cli Client) (Client, error) {
	if err := svc.authorize(ctx, "client", policies.Policy{Subject: token, Object: cli.ID, Actions: []string{"c_update"}}); err != nil {
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
	if err := svc.authorize(ctx, "client", policies.Policy{Subject: token, Object: id, Actions: []string{"c_delete"}}); err != nil {
		return Client{}, err
	}
	client, err := svc.changeClientStatus(ctx, id, EnabledStatus)
	if err != nil {
		return Client{}, errors.Wrap(ErrEnableClient, err)
	}

	return client, nil
}

func (svc service) DisableClient(ctx context.Context, token, id string) (Client, error) {
	if err := svc.authorize(ctx, "client", policies.Policy{Subject: token, Object: id, Actions: []string{"c_delete"}}); err != nil {
		return Client{}, err
	}
	client, err := svc.changeClientStatus(ctx, id, DisabledStatus)
	if err != nil {
		return Client{}, errors.Wrap(ErrDisableClient, err)
	}

	return client, nil
}

func (svc service) ListMembers(ctx context.Context, token, groupID string, pm Page) (MembersPage, error) {
	ir, err := svc.Identify(ctx, token)
	if err != nil {
		return MembersPage{}, err
	}
	pm.Subject = ir.ID
	pm.Action = "g_list"

	return svc.clients.Members(ctx, groupID, pm)
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

func (svc service) authorize(ctx context.Context, entityType string, p policies.Policy) error {
	if err := p.Validate(); err != nil {
		return err
	}
	ir, err := svc.Identify(ctx, p.Subject)
	if err != nil {
		return err
	}
	p.Subject = ir.ID
	return svc.policies.Evaluate(ctx, entityType, p)
}

func (svc service) Identify(ctx context.Context, tkn string) (UserIdentity, error) {
	claims, err := svc.tokens.Parse(ctx, tkn)
	if err != nil {
		return UserIdentity{}, errors.Wrap(errors.ErrAuthentication, err)
	}
	if claims.Type != jwt.AccessToken {
		return UserIdentity{}, errors.ErrAuthentication
	}

	return UserIdentity{ID: claims.ClientID, Email: claims.Email}, nil
}
