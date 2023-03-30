package mocks

import (
	"context"

	mfgroups "github.com/mainflux/mainflux/internal/mainflux/groups"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/users/groups"
	"github.com/stretchr/testify/mock"
)

const WrongID = "wrongID"

var _ groups.GroupRepository = (*GroupRepository)(nil)

type GroupRepository struct {
	mock.Mock
}

func (m *GroupRepository) ChangeStatus(ctx context.Context, group groups.Group) (groups.Group, error) {
	ret := m.Called(ctx, group)

	if group.ID == WrongID {
		return groups.Group{}, errors.ErrNotFound
	}
	if group.Status != groups.EnabledStatus && group.Status != groups.DisabledStatus {
		return groups.Group{}, errors.ErrMalformedEntity
	}

	return ret.Get(0).(mfgroups.Group), ret.Error(1)
}

func (m *GroupRepository) Memberships(ctx context.Context, clientID string, gm groups.GroupsPage) (groups.MembershipsPage, error) {
	ret := m.Called(ctx, clientID, gm)

	if clientID == WrongID {
		return groups.MembershipsPage{}, errors.ErrNotFound
	}

	return ret.Get(0).(groups.MembershipsPage), ret.Error(1)
}

func (m *GroupRepository) RetrieveAll(ctx context.Context, gm groups.GroupsPage) (groups.GroupsPage, error) {
	ret := m.Called(ctx, gm)

	return ret.Get(0).(groups.GroupsPage), ret.Error(1)
}

func (m *GroupRepository) RetrieveByID(ctx context.Context, id string) (mfgroups.Group, error) {
	ret := m.Called(ctx, id)
	if id == WrongID {
		return mfgroups.Group{}, errors.ErrNotFound
	}

	return ret.Get(0).(mfgroups.Group), ret.Error(1)
}

func (m *GroupRepository) Save(ctx context.Context, g mfgroups.Group) (mfgroups.Group, error) {
	ret := m.Called(ctx, g)
	if g.Parent == WrongID {
		return mfgroups.Group{}, errors.ErrCreateEntity
	}
	if g.Owner == WrongID {
		return mfgroups.Group{}, errors.ErrCreateEntity
	}

	return g, ret.Error(1)
}

func (m *GroupRepository) Update(ctx context.Context, g mfgroups.Group) (mfgroups.Group, error) {
	ret := m.Called(ctx, g)
	if g.ID == WrongID {
		return mfgroups.Group{}, errors.ErrNotFound
	}

	return ret.Get(0).(mfgroups.Group), ret.Error(1)
}
