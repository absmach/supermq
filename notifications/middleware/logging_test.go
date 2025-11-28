// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/absmach/supermq/notifications/middleware"
	"github.com/absmach/supermq/notifications/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLoggingMiddleware(t *testing.T) {
	notifier := new(mocks.Notifier)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	lm := middleware.NewLogging(notifier, logger)

	cases := []struct {
		desc              string
		inviterID         string
		inviteeID         string
		domainID          string
		domainName        string
		roleID            string
		roleName          string
		sendInvitationErr error
		sendAcceptanceErr error
		sendRejectionErr  error
	}{
		{
			desc:              "send invitation notification successfully",
			inviterID:         "inviter-1",
			inviteeID:         "invitee-1",
			domainID:          "domain-1",
			domainName:        "Test Domain",
			roleID:            "role-1",
			roleName:          "Admin",
			sendInvitationErr: nil,
		},
		{
			desc:              "send acceptance notification successfully",
			inviterID:         "inviter-1",
			inviteeID:         "invitee-1",
			domainID:          "domain-1",
			domainName:        "Test Domain",
			roleID:            "role-1",
			roleName:          "Admin",
			sendAcceptanceErr: nil,
		},
		{
			desc:             "send rejection notification successfully",
			inviterID:        "inviter-1",
			inviteeID:        "invitee-1",
			domainID:         "domain-1",
			domainName:       "Test Domain",
			roleID:           "role-1",
			roleName:         "Admin",
			sendRejectionErr: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			switch tc.desc {
			case "send invitation notification successfully":
				notifier.On("SendInvitationNotification", mock.Anything, tc.inviterID, tc.inviteeID, tc.domainID, tc.domainName, tc.roleID, tc.roleName).
					Return(tc.sendInvitationErr).Once()
				err := lm.SendInvitationNotification(context.Background(), tc.inviterID, tc.inviteeID, tc.domainID, tc.domainName, tc.roleID, tc.roleName)
				assert.Equal(t, tc.sendInvitationErr, err)
			case "send acceptance notification successfully":
				notifier.On("SendAcceptanceNotification", mock.Anything, tc.inviterID, tc.inviteeID, tc.domainID, tc.domainName, tc.roleID, tc.roleName).
					Return(tc.sendAcceptanceErr).Once()
				err := lm.SendAcceptanceNotification(context.Background(), tc.inviterID, tc.inviteeID, tc.domainID, tc.domainName, tc.roleID, tc.roleName)
				assert.Equal(t, tc.sendAcceptanceErr, err)
			case "send rejection notification successfully":
				notifier.On("SendRejectionNotification", mock.Anything, tc.inviterID, tc.inviteeID, tc.domainID, tc.domainName, tc.roleID, tc.roleName).
					Return(tc.sendRejectionErr).Once()
				err := lm.SendRejectionNotification(context.Background(), tc.inviterID, tc.inviteeID, tc.domainID, tc.domainName, tc.roleID, tc.roleName)
				assert.Equal(t, tc.sendRejectionErr, err)
			}
			notifier.AssertExpectations(t)
		})
	}
}
