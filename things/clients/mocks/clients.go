package mocks

import (
	"context"

	mfclients "github.com/mainflux/mainflux/pkg/clients"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/things/clients"
	"github.com/stretchr/testify/mock"
)

const WrongID = "wrongID"

var _ clients.Repository = (*ClientRepository)(nil)

type ClientRepository struct {
	mock.Mock
}

func (m *ClientRepository) ChangeStatus(ctx context.Context, client clients.Client) (clients.Client, error) {
	ret := m.Called(ctx, client)

	if client.ID == WrongID {
		return clients.Client{}, errors.ErrNotFound
	}

	if client.Status != mfclients.EnabledStatus && client.Status != mfclients.DisabledStatus {
		return clients.Client{}, errors.ErrMalformedEntity
	}

	return ret.Get(0).(clients.Client), ret.Error(1)
}

func (m *ClientRepository) Members(ctx context.Context, groupID string, pm clients.Page) (clients.MembersPage, error) {
	ret := m.Called(ctx, groupID, pm)
	if groupID == WrongID {
		return clients.MembersPage{}, errors.ErrNotFound
	}

	return ret.Get(0).(clients.MembersPage), ret.Error(1)
}

func (m *ClientRepository) RetrieveAll(ctx context.Context, pm clients.Page) (clients.ClientsPage, error) {
	ret := m.Called(ctx, pm)

	return ret.Get(0).(clients.ClientsPage), ret.Error(1)
}

func (m *ClientRepository) RetrieveByID(ctx context.Context, id string) (clients.Client, error) {
	ret := m.Called(ctx, id)

	if id == WrongID {
		return clients.Client{}, errors.ErrNotFound
	}

	return ret.Get(0).(clients.Client), ret.Error(1)
}

func (m *ClientRepository) RetrieveBySecret(ctx context.Context, secret string) (clients.Client, error) {
	ret := m.Called(ctx, secret)

	if secret == "" {
		return clients.Client{}, errors.ErrMalformedEntity
	}

	return ret.Get(0).(clients.Client), ret.Error(1)
}

func (m *ClientRepository) Save(ctx context.Context, clis ...clients.Client) ([]clients.Client, error) {
	ret := m.Called(ctx, clis)
	for _, cli := range clis {
		if cli.Owner == WrongID {
			return []clients.Client{}, errors.ErrMalformedEntity
		}
		if cli.Credentials.Secret == "" {
			return []clients.Client{}, errors.ErrMalformedEntity
		}
	}
	return clis, ret.Error(1)
}

func (m *ClientRepository) Update(ctx context.Context, client clients.Client) (clients.Client, error) {
	ret := m.Called(ctx, client)

	if client.ID == WrongID {
		return clients.Client{}, errors.ErrNotFound
	}
	return ret.Get(0).(clients.Client), ret.Error(1)
}

func (m *ClientRepository) UpdateIdentity(ctx context.Context, client clients.Client) (clients.Client, error) {
	ret := m.Called(ctx, client)

	if client.ID == WrongID {
		return clients.Client{}, errors.ErrNotFound
	}
	if client.Credentials.Identity == "" {
		return clients.Client{}, errors.ErrMalformedEntity
	}

	return ret.Get(0).(clients.Client), ret.Error(1)
}

func (m *ClientRepository) UpdateSecret(ctx context.Context, client clients.Client) (clients.Client, error) {
	ret := m.Called(ctx, client)

	if client.ID == WrongID {
		return clients.Client{}, errors.ErrNotFound
	}
	if client.Credentials.Secret == "" {
		return clients.Client{}, errors.ErrMalformedEntity
	}

	return ret.Get(0).(clients.Client), ret.Error(1)
}

func (m *ClientRepository) UpdateTags(ctx context.Context, client clients.Client) (clients.Client, error) {
	ret := m.Called(ctx, client)

	if client.ID == WrongID {
		return clients.Client{}, errors.ErrNotFound
	}

	return ret.Get(0).(clients.Client), ret.Error(1)
}

func (m *ClientRepository) UpdateOwner(ctx context.Context, client clients.Client) (clients.Client, error) {
	ret := m.Called(ctx, client)

	if client.ID == WrongID {
		return clients.Client{}, errors.ErrNotFound
	}

	return ret.Get(0).(clients.Client), ret.Error(1)
}
