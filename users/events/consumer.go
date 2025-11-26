// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"context"
	"log/slog"

	"github.com/absmach/supermq/pkg/events"
	"github.com/absmach/supermq/pkg/events/store"
	"github.com/absmach/supermq/users"
)

const (
	invitationSend   = "invitation.send"
	invitationAccept = "invitation.accept"
)

var _ events.EventHandler = (*eventHandler)(nil)

type eventHandler struct {
	emailer  users.Emailer
	userRepo users.Repository
	logger   *slog.Logger
}

// NewEventHandler creates a new event handler for testing purposes.
func NewEventHandler(emailer users.Emailer, userRepo users.Repository, logger *slog.Logger) events.EventHandler {
	return &eventHandler{
		emailer:  emailer,
		userRepo: userRepo,
		logger:   logger,
	}
}

// Start starts the event consumer for invitation events.
func Start(ctx context.Context, consumer string, sub events.Subscriber, emailer users.Emailer, userRepo users.Repository, logger *slog.Logger) error {
	handler := &eventHandler{
		emailer:  emailer,
		userRepo: userRepo,
		logger:   logger,
	}

	subCfg := events.SubscriberConfig{
		Consumer: consumer,
		Stream:   store.StreamAllEvents,
		Handler:  handler,
	}

	return sub.Subscribe(ctx, subCfg)
}

// Handle handles invitation events.
func (h *eventHandler) Handle(ctx context.Context, event events.Event) error {
	data, err := event.Encode()
	if err != nil {
		h.logger.Error("failed to encode event", slog.Any("error", err))
		return nil
	}

	operation, ok := data["operation"].(string)
	if !ok {
		return nil
	}

	switch operation {
	case invitationSend:
		return handleInvitationSent(ctx, data, h.emailer, h.userRepo, h.logger)
	case invitationAccept:
		return handleInvitationAccepted(ctx, data, h.emailer, h.userRepo, h.logger)
	default:
		return nil
	}
}

func handleInvitationSent(ctx context.Context, data map[string]any, emailer users.Emailer, userRepo users.Repository, logger *slog.Logger) error {
	inviteeUserID, _ := data["invitee_user_id"].(string)
	invitedBy, _ := data["invited_by"].(string)
	domainName, _ := data["domain_name"].(string)
	roleName, _ := data["role_name"].(string)

	if inviteeUserID == "" || invitedBy == "" {
		logger.Warn("missing required fields in invitation.send event",
			slog.String("invitee_user_id", inviteeUserID),
			slog.String("invited_by", invitedBy),
		)
		return nil
	}

	// Retrieve invitee user
	invitee, err := userRepo.RetrieveByID(ctx, inviteeUserID)
	if err != nil {
		logger.Error("failed to retrieve invitee user",
			slog.String("user_id", inviteeUserID),
			slog.Any("error", err),
		)
		return nil
	}

	// Retrieve inviter user
	inviter, err := userRepo.RetrieveByID(ctx, invitedBy)
	if err != nil {
		logger.Error("failed to retrieve inviter user",
			slog.String("user_id", invitedBy),
			slog.Any("error", err),
		)
		return nil
	}

	// Normalize names for display
	inviteeName := invitee.FirstName + " " + invitee.LastName
	if inviteeName == " " || inviteeName == "" {
		inviteeName = invitee.Credentials.Username
	}
	if inviteeName == "" {
		inviteeName = invitee.Email
	}

	inviterName := inviter.FirstName + " " + inviter.LastName
	if inviterName == " " || inviterName == "" {
		inviterName = inviter.Credentials.Username
	}
	if inviterName == "" {
		inviterName = inviter.Email
	}
	if inviterName == "" {
		inviterName = "A user"
	}

	if domainName == "" {
		domainName = "a domain"
	}

	if roleName == "" {
		roleName = "member"
	}

	// Send invitation email
	if err := emailer.SendInvitation([]string{invitee.Email}, inviteeName, inviterName, domainName, roleName); err != nil {
		logger.Error("failed to send invitation email",
			slog.String("to", invitee.Email),
			slog.Any("error", err),
		)
		return nil
	}

	logger.Info("invitation email sent",
		slog.String("to", invitee.Email),
		slog.String("domain", domainName),
	)

	return nil
}

func handleInvitationAccepted(ctx context.Context, data map[string]any, emailer users.Emailer, userRepo users.Repository, logger *slog.Logger) error {
	inviteeUserID, _ := data["invitee_user_id"].(string)
	invitedBy, _ := data["invited_by"].(string)
	domainName, _ := data["domain_name"].(string)
	roleName, _ := data["role_name"].(string)

	if inviteeUserID == "" || invitedBy == "" {
		logger.Warn("missing required fields in invitation.accept event",
			slog.String("invitee_user_id", inviteeUserID),
			slog.String("invited_by", invitedBy),
		)
		return nil
	}

	// Retrieve invitee user
	invitee, err := userRepo.RetrieveByID(ctx, inviteeUserID)
	if err != nil {
		logger.Error("failed to retrieve invitee user",
			slog.String("user_id", inviteeUserID),
			slog.Any("error", err),
		)
		return nil
	}

	// Retrieve inviter user
	inviter, err := userRepo.RetrieveByID(ctx, invitedBy)
	if err != nil {
		logger.Error("failed to retrieve inviter user",
			slog.String("user_id", invitedBy),
			slog.Any("error", err),
		)
		return nil
	}

	// Normalize names for display
	inviteeName := invitee.FirstName + " " + invitee.LastName
	if inviteeName == " " || inviteeName == "" {
		inviteeName = invitee.Credentials.Username
	}
	if inviteeName == "" {
		inviteeName = invitee.Email
	}
	if inviteeName == "" {
		inviteeName = "A user"
	}

	inviterName := inviter.FirstName + " " + inviter.LastName
	if inviterName == " " || inviterName == "" {
		inviterName = inviter.Credentials.Username
	}
	if inviterName == "" {
		inviterName = inviter.Email
	}

	if domainName == "" {
		domainName = "a domain"
	}

	if roleName == "" {
		roleName = "member"
	}

	// Send invitation accepted email
	if err := emailer.SendInvitationAccepted([]string{inviter.Email}, inviterName, inviteeName, domainName, roleName); err != nil {
		logger.Error("failed to send invitation accepted email",
			slog.String("to", inviter.Email),
			slog.Any("error", err),
		)
		return nil
	}

	logger.Info("invitation accepted email sent",
		slog.String("to", inviter.Email),
		slog.String("domain", domainName),
	)

	return nil
}
