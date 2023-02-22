package mocks

import (
	"context"

	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/things/clients"
	"github.com/stretchr/testify/mock"
)

const WrongID = "wrongID"

var _ clients.Repository = (*crepo)(nil)

type crepo struct {
	mock.Mock
}

func (m *crepo) ChangeStatus(ctx context.Context, id string, status clients.Status) (clients.Client, error) {
	ret := m.Called(ctx, id, status)

	if id == WrongID {
		return clients.Client{}, errors.ErrNotFound
	}

	if status != clients.EnabledStatus && status != clients.DisabledStatus {
		return clients.Client{}, errors.ErrMalformedEntity
	}

	return ret.Get(0).(clients.Client), ret.Error(1)
}

func (m *crepo) Members(ctx context.Context, groupID string, pm clients.Page) (clients.MembersPage, error) {
	ret := m.Called(ctx, groupID, pm)
	if groupID == WrongID {
		return clients.MembersPage{}, errors.ErrNotFound
	}

	return ret.Get(0).(clients.MembersPage), ret.Error(1)
}

func (m *crepo) RetrieveAll(ctx context.Context, pm clients.Page) (clients.ClientsPage, error) {
	ret := m.Called(ctx, pm)

	return ret.Get(0).(clients.ClientsPage), ret.Error(1)
}

func (m *crepo) RetrieveByID(ctx context.Context, id string) (clients.Client, error) {
	ret := m.Called(ctx, id)

	if id == WrongID {
		return clients.Client{}, errors.ErrNotFound
	}

	return ret.Get(0).(clients.Client), ret.Error(1)
}

func (m *crepo) RetrieveBySecret(ctx context.Context, secret string) (clients.Client, error) {
	ret := m.Called(ctx, secret)

	if secret == "" {
		return clients.Client{}, errors.ErrMalformedEntity
	}

	return ret.Get(0).(clients.Client), ret.Error(1)
}

func (m *crepo) Save(ctx context.Context, clis ...clients.Client) ([]clients.Client, error) {
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

func (m *crepo) Update(ctx context.Context, client clients.Client) (clients.Client, error) {
	ret := m.Called(ctx, client)

	if client.ID == WrongID {
		return clients.Client{}, errors.ErrNotFound
	}
	return ret.Get(0).(clients.Client), ret.Error(1)
}

func (m *crepo) UpdateIdentity(ctx context.Context, client clients.Client) (clients.Client, error) {
	ret := m.Called(ctx, client)

	if client.ID == WrongID {
		return clients.Client{}, errors.ErrNotFound
	}
	if client.Credentials.Identity == "" {
		return clients.Client{}, errors.ErrMalformedEntity
	}

	return ret.Get(0).(clients.Client), ret.Error(1)
}

func (m *crepo) UpdateSecret(ctx context.Context, client clients.Client) (clients.Client, error) {
	ret := m.Called(ctx, client)

	if client.ID == WrongID {
		return clients.Client{}, errors.ErrNotFound
	}
	if client.Credentials.Secret == "" {
		return clients.Client{}, errors.ErrMalformedEntity
	}

	return ret.Get(0).(clients.Client), ret.Error(1)
}

func (m *crepo) UpdateTags(ctx context.Context, client clients.Client) (clients.Client, error) {
	ret := m.Called(ctx, client)

	if client.ID == WrongID {
		return clients.Client{}, errors.ErrNotFound
	}

	return ret.Get(0).(clients.Client), ret.Error(1)
}

func (m *crepo) UpdateOwner(ctx context.Context, client clients.Client) (clients.Client, error) {
	ret := m.Called(ctx, client)

	if client.ID == WrongID {
		return clients.Client{}, errors.ErrNotFound
	}

	return ret.Get(0).(clients.Client), ret.Error(1)
}
