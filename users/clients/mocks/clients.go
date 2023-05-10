package mocks

import (
	"context"

	mfclients "github.com/mainflux/mainflux/pkg/clients"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/stretchr/testify/mock"
)

const WrongID = "wrongID"

var _ mfclients.Repository = (*ClientRepository)(nil)

type ClientRepository struct {
	mock.Mock
}

func (m *ClientRepository) ChangeStatus(ctx context.Context, client mfclients.Client) (mfclients.Client, error) {
	ret := m.Called(ctx, client)

	if client.ID == WrongID {
		return mfclients.Client{}, errors.ErrNotFound
	}

	if client.Status != mfclients.EnabledStatus && client.Status != mfclients.DisabledStatus {
		return mfclients.Client{}, errors.ErrMalformedEntity
	}

	return ret.Get(0).(mfclients.Client), ret.Error(1)
}

func (m *ClientRepository) Members(ctx context.Context, groupID string, pm mfclients.Page) (mfclients.MembersPage, error) {
	ret := m.Called(ctx, groupID, pm)
	if groupID == WrongID {
		return mfclients.MembersPage{}, errors.ErrNotFound
	}

	return ret.Get(0).(mfclients.MembersPage), ret.Error(1)
}

func (m *ClientRepository) RetrieveAll(ctx context.Context, pm mfclients.Page) (mfclients.ClientsPage, error) {
	ret := m.Called(ctx, pm)

	return ret.Get(0).(mfclients.ClientsPage), ret.Error(1)
}

func (m *ClientRepository) RetrieveByID(ctx context.Context, id string) (mfclients.Client, error) {
	ret := m.Called(ctx, id)

	if id == WrongID {
		return mfclients.Client{}, errors.ErrNotFound
	}

	return ret.Get(0).(mfclients.Client), ret.Error(1)
}

func (m *ClientRepository) RetrieveByIdentity(ctx context.Context, identity string) (mfclients.Client, error) {
	ret := m.Called(ctx, identity)

	if identity == "" {
		return mfclients.Client{}, errors.ErrMalformedEntity
	}

	return ret.Get(0).(mfclients.Client), ret.Error(1)
}

func (m *ClientRepository) Save(ctx context.Context, clients ...mfclients.Client) ([]mfclients.Client, error) {
	client := clients[0]
	ret := m.Called(ctx, client)
	if client.Owner == WrongID {
		return []mfclients.Client{}, errors.ErrMalformedEntity
	}
	if client.Credentials.Secret == "" {
		return []mfclients.Client{}, errors.ErrMalformedEntity
	}

	return clients, ret.Error(1)
}

func (m *ClientRepository) Update(ctx context.Context, client mfclients.Client) (mfclients.Client, error) {
	ret := m.Called(ctx, client)

	if client.ID == WrongID {
		return mfclients.Client{}, errors.ErrNotFound
	}
	return ret.Get(0).(mfclients.Client), ret.Error(1)
}

func (m *ClientRepository) UpdateIdentity(ctx context.Context, client mfclients.Client) (mfclients.Client, error) {
	ret := m.Called(ctx, client)

	if client.ID == WrongID {
		return mfclients.Client{}, errors.ErrNotFound
	}
	if client.Credentials.Identity == "" {
		return mfclients.Client{}, errors.ErrMalformedEntity
	}

	return ret.Get(0).(mfclients.Client), ret.Error(1)
}

func (m *ClientRepository) UpdateSecret(ctx context.Context, client mfclients.Client) (mfclients.Client, error) {
	ret := m.Called(ctx, client)

	if client.ID == WrongID {
		return mfclients.Client{}, errors.ErrNotFound
	}
	if client.Credentials.Secret == "" {
		return mfclients.Client{}, errors.ErrMalformedEntity
	}

	return ret.Get(0).(mfclients.Client), ret.Error(1)
}

func (m *ClientRepository) UpdateTags(ctx context.Context, client mfclients.Client) (mfclients.Client, error) {
	ret := m.Called(ctx, client)

	if client.ID == WrongID {
		return mfclients.Client{}, errors.ErrNotFound
	}

	return ret.Get(0).(mfclients.Client), ret.Error(1)
}

func (m *ClientRepository) UpdateOwner(ctx context.Context, client mfclients.Client) (mfclients.Client, error) {
	ret := m.Called(ctx, client)

	if client.ID == WrongID {
		return mfclients.Client{}, errors.ErrNotFound
	}

	return ret.Get(0).(mfclients.Client), ret.Error(1)
}

func (*ClientRepository) RetrieveBySecret(ctx context.Context, key string) (mfclients.Client, error) {
	return mfclients.Client{}, nil
}
