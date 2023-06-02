package mocks

import (
	"context"

	"github.com/mainflux/mainflux/things/policies"
	"github.com/stretchr/testify/mock"
)

type PolicyRepository struct {
	mock.Mock
}

func (m *PolicyRepository) Delete(ctx context.Context, p policies.Policy) error {
	ret := m.Called(ctx, p)

	return ret.Error(0)
}

func (m *PolicyRepository) Retrieve(ctx context.Context, pm policies.Page) (policies.PolicyPage, error) {
	ret := m.Called(ctx, pm)

	return ret.Get(0).(policies.PolicyPage), ret.Error(1)
}

func (m *PolicyRepository) Save(ctx context.Context, p policies.Policy) (policies.Policy, error) {
	ret := m.Called(ctx, p)

	return ret.Get(0).(policies.Policy), ret.Error(1)
}

func (m *PolicyRepository) Update(ctx context.Context, p policies.Policy) (policies.Policy, error) {
	ret := m.Called(ctx, p)

	return ret.Get(0).(policies.Policy), ret.Error(1)
}

func (m *PolicyRepository) Evaluate(ctx context.Context, entityType string, p policies.Policy) error {
	ret := m.Called(ctx, entityType, p)

	return ret.Error(0)
}

func (m *PolicyRepository) RetrieveOne(ctx context.Context, subject, object string) (policies.Policy, error) {
	ret := m.Called(ctx, subject, object)

	return ret.Get(0).(policies.Policy), ret.Error(1)
}
