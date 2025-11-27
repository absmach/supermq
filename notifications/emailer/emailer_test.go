// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package emailer_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	grpcUsersV1 "github.com/absmach/supermq/api/grpc/users/v1"
	"github.com/absmach/supermq/notifications/emailer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

const (
	inviterID  = "inviter-id"
	inviteeID  = "invitee-id"
	domainID   = "domain-id"
	domainName = "Test Domain"
	roleID     = "role-id"
	roleName   = "Admin"

	inviterEmail  = "inviter@example.com"
	inviteeEmail  = "invitee@example.com"
	inviterFirst  = "John"
	inviterLast   = "Doe"
	inviteeFirst  = "Jane"
	inviteeLast   = "Smith"
)

type mockUsersClient struct {
	mock.Mock
	grpcUsersV1.UsersServiceClient
}

func (m *mockUsersClient) RetrieveUsers(ctx context.Context, req *grpcUsersV1.RetrieveUsersReq, opts ...grpc.CallOption) (*grpcUsersV1.RetrieveUsersRes, error) {
	args := m.Called(ctx, req, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*grpcUsersV1.RetrieveUsersRes), args.Error(1)
}

func TestSendInvitationNotification(t *testing.T) {
	if os.Getenv("SMQ_RUN_EMAIL_TESTS") != "true" {
		t.Skip("Skipping email tests. Set SMQ_RUN_EMAIL_TESTS=true to run.")
	}

	usersClient := new(mockUsersClient)

	cfg := emailer.Config{
		FromAddress:        "test@example.com",
		FromName:           "Test Service",
		InvitationTemplate: "../../docker/templates/invitation-sent-email.tmpl",
		AcceptanceTemplate: "../../docker/templates/invitation-accepted-email.tmpl",
		RejectionTemplate:  "../../docker/templates/invitation-rejected-email.tmpl",
		EmailHost:          "localhost",
		EmailPort:          "1025",
		EmailUsername:      "",
		EmailPassword:      "",
	}

	notifier, err := emailer.New(usersClient, cfg)
	assert.NoError(t, err)
	assert.NotNil(t, notifier)

	cases := []struct {
		desc          string
		inviterID     string
		inviteeID     string
		domainID      string
		domainName    string
		roleID        string
		roleName      string
		setupMock     func()
		expectedError error
	}{
		{
			desc:       "successful invitation notification",
			inviterID:  inviterID,
			inviteeID:  inviteeID,
			domainID:   domainID,
			domainName: domainName,
			roleID:     roleID,
			roleName:   roleName,
			setupMock: func() {
				usersClient.On("RetrieveUsers", mock.Anything, mock.MatchedBy(func(req *grpcUsersV1.RetrieveUsersReq) bool {
					return len(req.Ids) == 1 && req.Ids[0] == inviterID
				}), mock.Anything).Return(&grpcUsersV1.RetrieveUsersRes{
					Users: []*grpcUsersV1.User{
						{
							Id:        inviterID,
							Email:     inviterEmail,
							FirstName: inviterFirst,
							LastName:  inviterLast,
						},
					},
				}, nil).Once()

				usersClient.On("RetrieveUsers", mock.Anything, mock.MatchedBy(func(req *grpcUsersV1.RetrieveUsersReq) bool {
					return len(req.Ids) == 1 && req.Ids[0] == inviteeID
				}), mock.Anything).Return(&grpcUsersV1.RetrieveUsersRes{
					Users: []*grpcUsersV1.User{
						{
							Id:        inviteeID,
							Email:     inviteeEmail,
							FirstName: inviteeFirst,
							LastName:  inviteeLast,
						},
					},
				}, nil).Once()
			},
			expectedError: nil,
		},
		{
			desc:       "failed to fetch inviter",
			inviterID:  inviterID,
			inviteeID:  inviteeID,
			domainID:   domainID,
			domainName: domainName,
			roleID:     roleID,
			roleName:   roleName,
			setupMock: func() {
				usersClient.On("RetrieveUsers", mock.Anything, mock.MatchedBy(func(req *grpcUsersV1.RetrieveUsersReq) bool {
					return len(req.Ids) == 1 && req.Ids[0] == inviterID
				}), mock.Anything).Return(nil, fmt.Errorf("user not found")).Once()
			},
			expectedError: fmt.Errorf("user not found"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			tc.setupMock()
			err := notifier.SendInvitationNotification(context.Background(), tc.inviterID, tc.inviteeID, tc.domainID, tc.domainName, tc.roleID, tc.roleName)
			if tc.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			usersClient.AssertExpectations(t)
		})
	}
}

func TestSendAcceptanceNotification(t *testing.T) {
	if os.Getenv("SMQ_RUN_EMAIL_TESTS") != "true" {
		t.Skip("Skipping email tests. Set SMQ_RUN_EMAIL_TESTS=true to run.")
	}

	usersClient := new(mockUsersClient)

	cfg := emailer.Config{
		FromAddress:        "test@example.com",
		FromName:           "Test Service",
		InvitationTemplate: "../../docker/templates/invitation-sent-email.tmpl",
		AcceptanceTemplate: "../../docker/templates/invitation-accepted-email.tmpl",
		RejectionTemplate:  "../../docker/templates/invitation-rejected-email.tmpl",
		EmailHost:          "localhost",
		EmailPort:          "1025",
		EmailUsername:      "",
		EmailPassword:      "",
	}

	notifier, err := emailer.New(usersClient, cfg)
	assert.NoError(t, err)

	usersClient.On("RetrieveUsers", mock.Anything, mock.MatchedBy(func(req *grpcUsersV1.RetrieveUsersReq) bool {
		return len(req.Ids) == 1 && req.Ids[0] == inviterID
	}), mock.Anything).Return(&grpcUsersV1.RetrieveUsersRes{
		Users: []*grpcUsersV1.User{
			{
				Id:        inviterID,
				Email:     inviterEmail,
				FirstName: inviterFirst,
				LastName:  inviterLast,
			},
		},
	}, nil).Once()

	usersClient.On("RetrieveUsers", mock.Anything, mock.MatchedBy(func(req *grpcUsersV1.RetrieveUsersReq) bool {
		return len(req.Ids) == 1 && req.Ids[0] == inviteeID
	}), mock.Anything).Return(&grpcUsersV1.RetrieveUsersRes{
		Users: []*grpcUsersV1.User{
			{
				Id:        inviteeID,
				Email:     inviteeEmail,
				FirstName: inviteeFirst,
				LastName:  inviteeLast,
			},
		},
	}, nil).Once()

	err = notifier.SendAcceptanceNotification(context.Background(), inviterID, inviteeID, domainID, domainName, roleID, roleName)
	assert.NoError(t, err)
	usersClient.AssertExpectations(t)
}

func TestSendRejectionNotification(t *testing.T) {
	if os.Getenv("SMQ_RUN_EMAIL_TESTS") != "true" {
		t.Skip("Skipping email tests. Set SMQ_RUN_EMAIL_TESTS=true to run.")
	}

	usersClient := new(mockUsersClient)

	cfg := emailer.Config{
		FromAddress:        "test@example.com",
		FromName:           "Test Service",
		InvitationTemplate: "../../docker/templates/invitation-sent-email.tmpl",
		AcceptanceTemplate: "../../docker/templates/invitation-accepted-email.tmpl",
		RejectionTemplate:  "../../docker/templates/invitation-rejected-email.tmpl",
		EmailHost:          "localhost",
		EmailPort:          "1025",
		EmailUsername:      "",
		EmailPassword:      "",
	}

	notifier, err := emailer.New(usersClient, cfg)
	assert.NoError(t, err)

	usersClient.On("RetrieveUsers", mock.Anything, mock.MatchedBy(func(req *grpcUsersV1.RetrieveUsersReq) bool {
		return len(req.Ids) == 1 && req.Ids[0] == inviterID
	}), mock.Anything).Return(&grpcUsersV1.RetrieveUsersRes{
		Users: []*grpcUsersV1.User{
			{
				Id:        inviterID,
				Email:     inviterEmail,
				FirstName: inviterFirst,
				LastName:  inviterLast,
			},
		},
	}, nil).Once()

	usersClient.On("RetrieveUsers", mock.Anything, mock.MatchedBy(func(req *grpcUsersV1.RetrieveUsersReq) bool {
		return len(req.Ids) == 1 && req.Ids[0] == inviteeID
	}), mock.Anything).Return(&grpcUsersV1.RetrieveUsersRes{
		Users: []*grpcUsersV1.User{
			{
				Id:        inviteeID,
				Email:     inviteeEmail,
				FirstName: inviteeFirst,
				LastName:  inviteeLast,
			},
		},
	}, nil).Once()

	err = notifier.SendRejectionNotification(context.Background(), inviterID, inviteeID, domainID, domainName, roleID, roleName)
	assert.NoError(t, err)
	usersClient.AssertExpectations(t)
}
