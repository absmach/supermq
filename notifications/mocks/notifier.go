// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"

	"github.com/absmach/supermq/notifications"
	"github.com/stretchr/testify/mock"
)

var _ notifications.Notifier = (*Notifier)(nil)

type Notifier struct {
	mock.Mock
}

func (m *Notifier) SendInvitationNotification(ctx context.Context, inviterID, inviteeID, domainID, domainName, roleID, roleName string) error {
	ret := m.Called(ctx, inviterID, inviteeID, domainID, domainName, roleID, roleName)
	return ret.Error(0)
}

func (m *Notifier) SendAcceptanceNotification(ctx context.Context, inviterID, inviteeID, domainID, domainName, roleID, roleName string) error {
	ret := m.Called(ctx, inviterID, inviteeID, domainID, domainName, roleID, roleName)
	return ret.Error(0)
}

func (m *Notifier) SendRejectionNotification(ctx context.Context, inviterID, inviteeID, domainID, domainName, roleID, roleName string) error {
	ret := m.Called(ctx, inviterID, inviteeID, domainID, domainName, roleID, roleName)
	return ret.Error(0)
}
