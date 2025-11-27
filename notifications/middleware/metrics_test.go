// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"context"
	"testing"

	"github.com/absmach/supermq/notifications/middleware"
	"github.com/absmach/supermq/notifications/mocks"
	"github.com/go-kit/kit/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockCounter struct {
	mock.Mock
	metrics.Counter
}

func (m *mockCounter) Add(delta float64) {
	m.Called(delta)
}

func (m *mockCounter) With(labelValues ...string) metrics.Counter {
	args := m.Called(labelValues)
	return args.Get(0).(metrics.Counter)
}

type mockHistogram struct {
	mock.Mock
	metrics.Histogram
}

func (m *mockHistogram) Observe(value float64) {
	m.Called(value)
}

func (m *mockHistogram) With(labelValues ...string) metrics.Histogram {
	args := m.Called(labelValues)
	return args.Get(0).(metrics.Histogram)
}

func TestMetricsMiddleware(t *testing.T) {
	notifier := new(mocks.Notifier)
	counter := new(mockCounter)
	histogram := new(mockHistogram)

	counter.On("With", mock.Anything).Return(counter)
	counter.On("Add", mock.Anything).Return()
	histogram.On("With", mock.Anything).Return(histogram)
	histogram.On("Observe", mock.Anything).Return()

	mm := middleware.NewMetrics(notifier, counter, histogram)

	notifier.On("SendInvitationNotification", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).Once()

	err := mm.SendInvitationNotification(context.Background(), "inv1", "inv2", "dom1", "Domain", "role1", "Admin")
	assert.NoError(t, err)
	notifier.AssertExpectations(t)
	counter.AssertCalled(t, "With", []string{"method", "send_invitation_notification"})
	counter.AssertCalled(t, "Add", mock.Anything)
	histogram.AssertCalled(t, "With", []string{"method", "send_invitation_notification"})
	histogram.AssertCalled(t, "Observe", mock.Anything)
}
