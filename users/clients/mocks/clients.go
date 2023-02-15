package mocks

import (
	"context"

	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/users/clients"
	"github.com/stretchr/testify/mock"
)

const WrongID = "wrongID"

var _ clients.ClientRepository = (*ClientRepository)(nil)

type ClientRepository struct {
	mock.Mock
}

func (m *ClientRepository) ChangeStatus(ctx context.Context, id string, status clients.Status) (clients.Client, error) {
	ret := m.Called(ctx, id, status)

	if id == WrongID {
		return clients.Client{}, errors.ErrNotFound
	}

	if status != clients.EnabledStatus && status != clients.DisabledStatus {
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

func (m *ClientRepository) RetrieveByIdentity(ctx context.Context, identity string) (clients.Client, error) {
	ret := m.Called(ctx, identity)

	if identity == "" {
		return clients.Client{}, errors.ErrMalformedEntity
	}

	return ret.Get(0).(clients.Client), ret.Error(1)
}

func (m *ClientRepository) Save(ctx context.Context, client clients.Client) (clients.Client, error) {
	ret := m.Called(ctx, client)
	if client.Owner == WrongID {
		return clients.Client{}, errors.ErrMalformedEntity
	}
	if client.Credentials.Secret == "" {
		return clients.Client{}, errors.ErrMalformedEntity
	}

	return client, ret.Error(1)
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
