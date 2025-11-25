// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/absmach/supermq/notifications"
	"github.com/absmach/supermq/pkg/events"
	"github.com/absmach/supermq/pkg/events/store"
)

const (
	invitationSend   = "invitation.send"
	invitationAccept = "invitation.accept"
)

func Start(ctx context.Context, consumer string, sub events.Subscriber, svc notifications.Service, userRepo notifications.UserRepository, domainRepo notifications.DomainRepository, invitationRepo notifications.InvitationRepository, logger *slog.Logger) error {
	subCfg := events.SubscriberConfig{
		Consumer: consumer,
		Stream:   store.StreamAllEvents,
		Handler:  Handle(svc, userRepo, domainRepo, invitationRepo, logger),
	}

	return sub.Subscribe(ctx, subCfg)
}

func Handle(svc notifications.Service, userRepo notifications.UserRepository, domainRepo notifications.DomainRepository, invitationRepo notifications.InvitationRepository, logger *slog.Logger) handleFunc {
	return func(ctx context.Context, event events.Event) error {
		data, err := event.Encode()
		if err != nil {
			logger.Error("failed to encode event", slog.Any("error", err))
			return nil
		}

		operation, ok := data["operation"].(string)
		if !ok {
			logger.Error("missing operation in event")
			return nil
		}
		fmt.Println("operation:", operation)

		switch operation {
		case invitationSend:
			return handleInvitationSent(ctx, data, svc, userRepo, domainRepo, logger)
		case invitationAccept:
			return handleInvitationAccepted(ctx, data, svc, userRepo, domainRepo, invitationRepo, logger)
		default:
			return nil
		}
	}
}

func handleInvitationSent(ctx context.Context, data map[string]any, svc notifications.Service, userRepo notifications.UserRepository, domainRepo notifications.DomainRepository, logger *slog.Logger) error {
	inviteeUserID, _ := data["invitee_user_id"].(string)
	domainID, _ := data["domain_id"].(string)
	invitedBy, _ := data["invited_by"].(string)
	roleID, _ := data["role_id"].(string)

	if inviteeUserID == "" || domainID == "" || invitedBy == "" {
		logger.Warn("missing required fields in invitation.send event",
			slog.String("invitee_user_id", inviteeUserID),
			slog.String("domain_id", domainID),
			slog.String("invited_by", invitedBy),
		)
		return nil
	}

	invitee, err := userRepo.RetrieveByID(ctx, inviteeUserID)
	if err != nil {
		logger.Error("failed to retrieve invitee user",
			slog.String("user_id", inviteeUserID),
			slog.Any("error", err),
		)
		return nil
	}

	inviter, err := userRepo.RetrieveByID(ctx, invitedBy)
	if err != nil {
		logger.Error("failed to retrieve inviter user",
			slog.String("user_id", invitedBy),
			slog.Any("error", err),
		)
		return nil
	}

	domain, err := domainRepo.RetrieveByID(ctx, domainID)
	if err != nil {
		logger.Error("failed to retrieve domain",
			slog.String("domain_id", domainID),
			slog.Any("error", err),
		)
		return nil
	}

	notification := notifications.Notification{
		Type:            notifications.InvitationSent,
		InviteeUserID:   inviteeUserID,
		InvitedBy:       invitedBy,
		DomainID:        domainID,
		DomainName:      domain.Name,
		RoleID:          roleID,
		InviteeEmail:    invitee.Email,
		InviterEmail:    inviter.Email,
		InviteeUsername: invitee.Username,
		InviterUsername: inviter.Username,
	}

	if err := svc.SendNotification(ctx, notification); err != nil {
		logger.Error("failed to send invitation notification",
			slog.String("invitee_email", invitee.Email),
			slog.Any("error", err),
		)
	}

	return nil
}

func handleInvitationAccepted(ctx context.Context, data map[string]any, svc notifications.Service, userRepo notifications.UserRepository, domainRepo notifications.DomainRepository, invitationRepo notifications.InvitationRepository, logger *slog.Logger) error {
	inviteeUserID, _ := data["invitee_user_id"].(string)
	domainID, _ := data["domain_id"].(string)

	if inviteeUserID == "" || domainID == "" {
		logger.Warn("missing required fields in invitation.accept event",
			slog.String("invitee_user_id", inviteeUserID),
			slog.String("domain_id", domainID),
		)
		return nil
	}

	// Retrieve invitation to find the inviter
	invitation, err := invitationRepo.RetrieveInvitation(ctx, inviteeUserID, domainID)
	if err != nil {
		logger.Error("failed to retrieve invitation",
			slog.String("invitee_user_id", inviteeUserID),
			slog.String("domain_id", domainID),
			slog.Any("error", err),
		)
		return nil
	}

	// Retrieve invitee user details
	invitee, err := userRepo.RetrieveByID(ctx, inviteeUserID)
	if err != nil {
		logger.Error("failed to retrieve invitee user",
			slog.String("user_id", inviteeUserID),
			slog.Any("error", err),
		)
		return nil
	}

	// Retrieve inviter user details
	inviter, err := userRepo.RetrieveByID(ctx, invitation.InvitedBy)
	if err != nil {
		logger.Error("failed to retrieve inviter user",
			slog.String("user_id", invitation.InvitedBy),
			slog.Any("error", err),
		)
		return nil
	}

	// Retrieve domain details
	domain, err := domainRepo.RetrieveByID(ctx, domainID)
	if err != nil {
		logger.Error("failed to retrieve domain",
			slog.String("domain_id", domainID),
			slog.Any("error", err),
		)
		return nil
	}

	notification := notifications.Notification{
		Type:            notifications.InvitationAccepted,
		InviteeUserID:   inviteeUserID,
		InvitedBy:       invitation.InvitedBy,
		DomainID:        domainID,
		DomainName:      domain.Name,
		InviteeEmail:    invitee.Email,
		InviterEmail:    inviter.Email,
		InviteeUsername: invitee.Username,
		InviterUsername: inviter.Username,
	}

	if err := svc.SendNotification(ctx, notification); err != nil {
		logger.Error("failed to send invitation accepted notification",
			slog.String("inviter_email", inviter.Email),
			slog.Any("error", err),
		)
	}

	return nil
}

type handleFunc func(ctx context.Context, event events.Event) error

func (h handleFunc) Handle(ctx context.Context, event events.Event) error {
	return h(ctx, event)
}

func (h handleFunc) Cancel() error {
	return nil
}
