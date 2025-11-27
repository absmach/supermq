// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/absmach/supermq/internal/testsutil"
	"github.com/absmach/supermq/pkg/errors"
	repoerr "github.com/absmach/supermq/pkg/errors/repository"
	"github.com/absmach/supermq/users"
	"github.com/absmach/supermq/users/events"
	"github.com/absmach/supermq/users/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	inviteeUserID = testsutil.GenerateUUID(&testing.T{})
	inviterUserID = testsutil.GenerateUUID(&testing.T{})
	domainID      = testsutil.GenerateUUID(&testing.T{})
	roleID        = testsutil.GenerateUUID(&testing.T{})
)

type testEvent struct {
	data map[string]any
	err  error
}

func (e testEvent) Encode() (map[string]any, error) {
	return e.data, e.err
}

func newTestEvent(data map[string]any, err error) testEvent {
	return testEvent{data: data, err: err}
}

func TestHandleInvitationSent(t *testing.T) {
	notifier := new(mocks.Notifier)
	repo := new(mocks.Repository)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	handler := events.NewEventHandler(notifier, repo, logger)

	invitee := users.User{
		ID:        inviteeUserID,
		Email:     "invitee@example.com",
		FirstName: "John",
		LastName:  "Doe",
	}

	inviter := users.User{
		ID:        inviterUserID,
		Email:     "inviter@example.com",
		FirstName: "Jane",
		LastName:  "Smith",
	}

	cases := []struct {
		desc             string
		event            map[string]any
		encodeErr        error
		retrieveInvitee  users.User
		retrieveInviter  users.User
		inviteeErr       error
		inviterErr       error
		notifyErr        error
		shouldCallNotify bool
	}{
		{
			desc: "successful invitation sent",
			event: map[string]any{
				"operation":       "invitation.send",
				"invitee_user_id": inviteeUserID,
				"invited_by":      inviterUserID,
				"domain_id":       domainID,
				"domain_name":     "Test Domain",
				"role_id":         roleID,
				"role_name":       "Admin",
			},
			retrieveInvitee:  invitee,
			retrieveInviter:  inviter,
			shouldCallNotify: true,
		},
		{
			desc: "invitation sent with missing domain name",
			event: map[string]any{
				"operation":       "invitation.send",
				"invitee_user_id": inviteeUserID,
				"invited_by":      inviterUserID,
				"domain_id":       domainID,
				"role_id":         roleID,
			},
			retrieveInvitee:  invitee,
			retrieveInviter:  inviter,
			shouldCallNotify: true,
		},
		{
			desc: "invitation sent with missing invitee_user_id",
			event: map[string]any{
				"operation":   "invitation.send",
				"invited_by":  inviterUserID,
				"domain_id":   domainID,
				"domain_name": "Test Domain",
			},
			shouldCallNotify: false,
		},
		{
			desc: "invitation sent with missing invited_by",
			event: map[string]any{
				"operation":       "invitation.send",
				"invitee_user_id": inviteeUserID,
				"domain_id":       domainID,
				"domain_name":     "Test Domain",
			},
			shouldCallNotify: false,
		},
		{
			desc: "invitation sent with invitee not found",
			event: map[string]any{
				"operation":       "invitation.send",
				"invitee_user_id": inviteeUserID,
				"invited_by":      inviterUserID,
				"domain_id":       domainID,
				"domain_name":     "Test Domain",
			},
			inviteeErr:       repoerr.ErrNotFound,
			shouldCallNotify: false,
		},
		{
			desc: "invitation sent with inviter not found",
			event: map[string]any{
				"operation":       "invitation.send",
				"invitee_user_id": inviteeUserID,
				"invited_by":      inviterUserID,
				"domain_id":       domainID,
				"domain_name":     "Test Domain",
			},
			retrieveInvitee:  invitee,
			inviterErr:       repoerr.ErrNotFound,
			shouldCallNotify: false,
		},
		{
			desc: "invitation sent with email error",
			event: map[string]any{
				"operation":       "invitation.send",
				"invitee_user_id": inviteeUserID,
				"invited_by":      inviterUserID,
				"domain_id":       domainID,
				"domain_name":     "Test Domain",
				"role_name":       "Admin",
			},
			retrieveInvitee:  invitee,
			retrieveInviter:  inviter,
			notifyErr:        errors.New("notification send failed"),
			shouldCallNotify: true,
		},
		{
			desc: "encode error",
			event: map[string]any{
				"operation": "invitation.send",
			},
			encodeErr:        errors.New("encode error"),
			shouldCallNotify: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.shouldCallNotify || tc.inviteeErr != nil || tc.inviterErr != nil {
				inviteeCall := repo.On("RetrieveByID", context.Background(), inviteeUserID).Return(tc.retrieveInvitee, tc.inviteeErr)
				var inviterCall, notifyCall *mock.Call
				if tc.inviteeErr == nil {
					inviterCall = repo.On("RetrieveByID", context.Background(), inviterUserID).Return(tc.retrieveInviter, tc.inviterErr)
					if tc.inviterErr == nil && tc.shouldCallNotify {
						notifyCall = notifier.On("Notify", context.Background(), mock.Anything).Return(tc.notifyErr)
					}
				}

				err := handler.Handle(context.Background(), newTestEvent(tc.event, tc.encodeErr))
				if tc.inviteeErr != nil || tc.inviterErr != nil || tc.notifyErr != nil {
					assert.NotNil(t, err, "Handle should return error")
				} else {
					assert.Nil(t, err, "Handle should not return error")
				}

				inviteeCall.Unset()
				if inviterCall != nil {
					inviterCall.Unset()
				}
				if notifyCall != nil {
					notifyCall.Unset()
				}
			} else {
				err := handler.Handle(context.Background(), newTestEvent(tc.event, tc.encodeErr))
				if tc.encodeErr != nil {
					assert.NotNil(t, err, "Handle should return error on encode failure")
				} else {
					assert.NotNil(t, err, "Handle should return error for missing required fields")
				}
			}
		})
	}
}

func TestHandleInvitationAccepted(t *testing.T) {
	notifier := new(mocks.Notifier)
	repo := new(mocks.Repository)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	handler := events.NewEventHandler(notifier, repo, logger)

	invitee := users.User{
		ID:        inviteeUserID,
		Email:     "invitee@example.com",
		FirstName: "John",
		LastName:  "Doe",
	}

	inviter := users.User{
		ID:        inviterUserID,
		Email:     "inviter@example.com",
		FirstName: "Jane",
		LastName:  "Smith",
	}

	cases := []struct {
		desc             string
		event            map[string]any
		encodeErr        error
		retrieveInvitee  users.User
		retrieveInviter  users.User
		inviteeErr       error
		inviterErr       error
		notifyErr        error
		shouldCallNotify bool
	}{
		{
			desc: "successful invitation accepted",
			event: map[string]any{
				"operation":       "invitation.accept",
				"invitee_user_id": inviteeUserID,
				"invited_by":      inviterUserID,
				"domain_id":       domainID,
				"domain_name":     "Test Domain",
				"role_id":         roleID,
				"role_name":       "Admin",
			},
			retrieveInvitee:  invitee,
			retrieveInviter:  inviter,
			shouldCallNotify: true,
		},
		{
			desc: "invitation accepted with missing domain name",
			event: map[string]any{
				"operation":       "invitation.accept",
				"invitee_user_id": inviteeUserID,
				"invited_by":      inviterUserID,
				"domain_id":       domainID,
				"role_id":         roleID,
			},
			retrieveInvitee:  invitee,
			retrieveInviter:  inviter,
			shouldCallNotify: true,
		},
		{
			desc: "invitation accepted with missing invitee_user_id",
			event: map[string]any{
				"operation":   "invitation.accept",
				"invited_by":  inviterUserID,
				"domain_id":   domainID,
				"domain_name": "Test Domain",
			},
			shouldCallNotify: false,
		},
		{
			desc: "invitation accepted with missing invited_by",
			event: map[string]any{
				"operation":       "invitation.accept",
				"invitee_user_id": inviteeUserID,
				"domain_id":       domainID,
				"domain_name":     "Test Domain",
			},
			shouldCallNotify: false,
		},
		{
			desc: "invitation accepted with invitee not found",
			event: map[string]any{
				"operation":       "invitation.accept",
				"invitee_user_id": inviteeUserID,
				"invited_by":      inviterUserID,
				"domain_id":       domainID,
				"domain_name":     "Test Domain",
			},
			inviteeErr:       repoerr.ErrNotFound,
			shouldCallNotify: false,
		},
		{
			desc: "invitation accepted with inviter not found",
			event: map[string]any{
				"operation":       "invitation.accept",
				"invitee_user_id": inviteeUserID,
				"invited_by":      inviterUserID,
				"domain_id":       domainID,
				"domain_name":     "Test Domain",
			},
			retrieveInvitee:  invitee,
			inviterErr:       repoerr.ErrNotFound,
			shouldCallNotify: false,
		},
		{
			desc: "invitation accepted with email error",
			event: map[string]any{
				"operation":       "invitation.accept",
				"invitee_user_id": inviteeUserID,
				"invited_by":      inviterUserID,
				"domain_id":       domainID,
				"domain_name":     "Test Domain",
				"role_name":       "Admin",
			},
			retrieveInvitee:  invitee,
			retrieveInviter:  inviter,
			notifyErr:        errors.New("notification send failed"),
			shouldCallNotify: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			if tc.shouldCallNotify || tc.inviteeErr != nil || tc.inviterErr != nil {
				inviteeCall := repo.On("RetrieveByID", context.Background(), inviteeUserID).Return(tc.retrieveInvitee, tc.inviteeErr)
				var inviterCall, notifyCall *mock.Call
				if tc.inviteeErr == nil {
					inviterCall = repo.On("RetrieveByID", context.Background(), inviterUserID).Return(tc.retrieveInviter, tc.inviterErr)
					if tc.inviterErr == nil && tc.shouldCallNotify {
						notifyCall = notifier.On("Notify", context.Background(), mock.Anything).Return(tc.notifyErr)
					}
				}

				err := handler.Handle(context.Background(), newTestEvent(tc.event, tc.encodeErr))
				if tc.inviteeErr != nil || tc.inviterErr != nil || tc.notifyErr != nil {
					assert.NotNil(t, err, "Handle should return error")
				} else {
					assert.Nil(t, err, "Handle should not return error")
				}

				inviteeCall.Unset()
				if inviterCall != nil {
					inviterCall.Unset()
				}
				if notifyCall != nil {
					notifyCall.Unset()
				}
			} else {
				err := handler.Handle(context.Background(), newTestEvent(tc.event, tc.encodeErr))
				if tc.encodeErr != nil {
					assert.NotNil(t, err, "Handle should return error on encode failure")
				} else {
					assert.NotNil(t, err, "Handle should return error for missing required fields")
				}
			}
		})
	}
}

func TestHandleUnknownOperation(t *testing.T) {
	notifier := new(mocks.Notifier)
	repo := new(mocks.Repository)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	handler := events.NewEventHandler(notifier, repo, logger)

	event := map[string]any{
		"operation": "unknown.operation",
		"data":      "some data",
	}

	err := handler.Handle(context.Background(), newTestEvent(event, nil))
	assert.Nil(t, err, "Handle should not return error for unknown operations")
}
