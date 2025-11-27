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
	invitationReject = "invitation.reject"

	defaultUserName   = "A user"
	defaultDomainName = "a domain"
	defaultRoleName   = "member"
)

var _ events.EventHandler = (*eventHandler)(nil)

type eventHandler struct {
	notifier users.Notifier
	userRepo users.Repository
	logger   *slog.Logger
}

// NewEventHandler creates a new event handler for testing purposes.
func NewEventHandler(notifier users.Notifier, userRepo users.Repository, logger *slog.Logger) events.EventHandler {
	return &eventHandler{
		notifier: notifier,
		userRepo: userRepo,
		logger:   logger,
	}
}

// Start starts the event consumer for invitation events.
func Start(ctx context.Context, consumer string, sub events.Subscriber, notifier users.Notifier, userRepo users.Repository, logger *slog.Logger) error {
	handler := &eventHandler{
		notifier: notifier,
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
		return err
	}

	operation, ok := data["operation"].(string)
	if !ok {
		return nil
	}

	switch operation {
	case invitationSend:
		return handleInvitationSent(ctx, data, h.notifier, h.userRepo, h.logger)
	case invitationAccept:
		return handleInvitationAccepted(ctx, data, h.notifier, h.userRepo, h.logger)
	case invitationReject:
		return handleInvitationRejected(ctx, data, h.notifier, h.userRepo, h.logger)
	default:
		return nil
	}
}

func handleInvitationSent(ctx context.Context, data map[string]any, notifier users.Notifier, userRepo users.Repository, logger *slog.Logger) error {
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
	if inviteeName == " " {
		inviteeName = invitee.Credentials.Username
	}
	if inviteeName == "" {
		inviteeName = invitee.Email
	}

	inviterName := inviter.FirstName + " " + inviter.LastName
	if inviterName == " " {
		inviterName = inviter.Credentials.Username
	}
	if inviterName == "" {
		inviterName = inviter.Email
	}
	if inviterName == "" {
		inviterName = defaultUserName
	}

	if domainName == "" {
		domainName = defaultDomainName
	}

	if roleName == "" {
		roleName = defaultRoleName
	}

	// Send invitation notification
	notification := &users.InvitationSentNotification{
		To:          []string{invitee.Email},
		InviteeName: inviteeName,
		InviterName: inviterName,
		DomainName:  domainName,
		RoleName:    roleName,
	}

	if err := notifier.Notify(ctx, notification); err != nil {
		logger.Error("failed to send invitation notification",
			slog.String("to", invitee.Email),
			slog.Any("error", err),
		)
		return nil
	}

	logger.Info("invitation notification sent",
		slog.String("to", invitee.Email),
		slog.String("domain", domainName),
	)

	return nil
}

func handleInvitationAccepted(ctx context.Context, data map[string]any, notifier users.Notifier, userRepo users.Repository, logger *slog.Logger) error {
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
	if inviteeName == " " {
		inviteeName = invitee.Credentials.Username
	}
	if inviteeName == "" {
		inviteeName = invitee.Email
	}
	if inviteeName == "" {
		inviteeName = defaultUserName
	}

	inviterName := inviter.FirstName + " " + inviter.LastName
	if inviterName == " " {
		inviterName = inviter.Credentials.Username
	}
	if inviterName == "" {
		inviterName = inviter.Email
	}

	if domainName == "" {
		domainName = defaultDomainName
	}

	if roleName == "" {
		roleName = defaultRoleName
	}

	// Send invitation accepted notification
	notification := &users.InvitationAcceptedNotification{
		To:          []string{inviter.Email},
		InviteeName: inviteeName,
		InviterName: inviterName,
		DomainName:  domainName,
		RoleName:    roleName,
	}

	if err := notifier.Notify(ctx, notification); err != nil {
		logger.Error("failed to send invitation accepted notification",
			slog.String("to", inviter.Email),
			slog.Any("error", err),
		)
		return nil
	}

	logger.Info("invitation accepted notification sent",
		slog.String("to", inviter.Email),
		slog.String("domain", domainName),
	)

	return nil
}

func handleInvitationRejected(ctx context.Context, data map[string]any, notifier users.Notifier, userRepo users.Repository, logger *slog.Logger) error {
	inviteeUserID, _ := data["invitee_user_id"].(string)
	invitedBy, _ := data["invited_by"].(string)
	domainName, _ := data["domain_name"].(string)
	roleName, _ := data["role_name"].(string)

	if inviteeUserID == "" || invitedBy == "" {
		logger.Warn("missing required fields in invitation.reject event",
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
	if inviteeName == " " {
		inviteeName = invitee.Credentials.Username
	}
	if inviteeName == "" {
		inviteeName = invitee.Email
	}
	if inviteeName == "" {
		inviteeName = defaultUserName
	}

	inviterName := inviter.FirstName + " " + inviter.LastName
	if inviterName == " " {
		inviterName = inviter.Credentials.Username
	}
	if inviterName == "" {
		inviterName = inviter.Email
	}

	if domainName == "" {
		domainName = defaultDomainName
	}

	if roleName == "" {
		roleName = defaultRoleName
	}

	// Send invitation rejected notification to the inviter
	notification := &users.InvitationRejectedNotification{
		To:          []string{inviter.Email},
		InviteeName: inviteeName,
		InviterName: inviterName,
		DomainName:  domainName,
		RoleName:    roleName,
	}

	if err := notifier.Notify(ctx, notification); err != nil {
		logger.Error("failed to send invitation rejected notification",
			slog.String("to", inviter.Email),
			slog.Any("error", err),
		)
		return nil
	}

	logger.Info("invitation rejected notification sent",
		slog.String("to", inviter.Email),
		slog.String("domain", domainName),
	)

	return nil
}
