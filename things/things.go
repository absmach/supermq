// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package things

import (
	"context"

	"github.com/absmach/magistrala"
	"github.com/absmach/magistrala/pkg/clients"
	"github.com/absmach/magistrala/pkg/errors"
)

// Service specifies an API that must be fullfiled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
//
//go:generate mockery --name Service --filename service.go --quiet --note "Copyright (c) Abstract Machines"
type Service interface {
	// CreateThings creates new client. In case of the failed registration, a
	// non-nil error value is returned.
	CreateThings(ctx context.Context, token string, client ...clients.Client) ([]clients.Client, errors.Error)

	// ViewClient retrieves client info for a given client ID and an authorized token.
	ViewClient(ctx context.Context, token, id string) (clients.Client, errors.Error)

	// ViewClientPerms retrieves permissions on the client id for the given authorized token.
	ViewClientPerms(ctx context.Context, token, id string) ([]string, errors.Error)

	// ListClients retrieves clients list for a valid auth token.
	ListClients(ctx context.Context, token string, reqUserID string, pm clients.Page) (clients.ClientsPage, errors.Error)

	// ListClientsByGroup retrieves data about subset of things that are
	// connected or not connected to specified channel and belong to the user identified by
	// the provided key.
	ListClientsByGroup(ctx context.Context, token, groupID string, pm clients.Page) (clients.MembersPage, errors.Error)

	// UpdateClient updates the client's name and metadata.
	UpdateClient(ctx context.Context, token string, client clients.Client) (clients.Client, errors.Error)

	// UpdateClientTags updates the client's tags.
	UpdateClientTags(ctx context.Context, token string, client clients.Client) (clients.Client, errors.Error)

	// UpdateClientSecret updates the client's secret
	UpdateClientSecret(ctx context.Context, token, id, key string) (clients.Client, errors.Error)

	// EnableClient logically enableds the client identified with the provided ID
	EnableClient(ctx context.Context, token, id string) (clients.Client, errors.Error)

	// DisableClient logically disables the client identified with the provided ID
	DisableClient(ctx context.Context, token, id string) (clients.Client, errors.Error)

	// Share add share policy to thing id with given relation for given user ids
	Share(ctx context.Context, token, id string, relation string, userids ...string) errors.Error

	// Unshare remove share policy to thing id with given relation for given user ids
	Unshare(ctx context.Context, token, id string, relation string, userids ...string) errors.Error

	// Identify returns thing ID for given thing key.
	Identify(ctx context.Context, key string) (string, errors.Error)

	// Authorize used for AuthZ gRPC server implementation and Things authorization.
	Authorize(ctx context.Context, req *magistrala.AuthorizeReq) (string, errors.Error)

	// DeleteClient deletes client with given ID.
	DeleteClient(ctx context.Context, token, id string) errors.Error
}

// Cache contains thing caching interface.
//
//go:generate mockery --name Cache --filename cache.go --quiet --note "Copyright (c) Abstract Machines"
type Cache interface {
	// Save stores pair thing secret, thing id.
	Save(ctx context.Context, thingSecret, thingID string) errors.Error

	// ID returns thing ID for given thing secret.
	ID(ctx context.Context, thingSecret string) (string, errors.Error)

	// Removes thing from cache.
	Remove(ctx context.Context, thingID string) errors.Error
}
