// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"context"
	"log/slog"

	"github.com/absmach/supermq/notifications"
	"github.com/absmach/supermq/pkg/events"
)

const (
	// Stream names.
	sendInvitationStream   = "supermq.invitation.send"
	acceptInvitationStream = "supermq.invitation.accept"
	rejectInvitationStream = "supermq.invitation.reject"
)

// Start starts consuming invitation events from the event store.
func Start(ctx context.Context, consumer string, sub events.Subscriber, notifier notifications.Notifier) error {
	handlers := []events.SubscriberConfig{
		{
			Consumer: consumer,
			Stream:   sendInvitationStream,
			Handler:  handleInvitationSent(notifier),
		},
		{
			Consumer: consumer,
			Stream:   acceptInvitationStream,
			Handler:  handleInvitationAccepted(notifier),
		},
		{
			Consumer: consumer,
			Stream:   rejectInvitationStream,
			Handler:  handleInvitationRejected(notifier),
		},
	}

	for _, handler := range handlers {
		if err := sub.Subscribe(ctx, handler); err != nil {
			return err
		}
	}

	return nil
}

func handleInvitationSent(notifier notifications.Notifier) handleFunc {
	return func(ctx context.Context, event events.Event) error {
		data, err := event.Encode()
		if err != nil {
			slog.Error("failed to encode invitation sent event", "error", err)
			return nil
		}

		invitedBy, ok := data["invited_by"].(string)
		if !ok || invitedBy == "" {
			slog.Error("missing or invalid invited_by in invitation sent event")
			return nil
		}

		inviteeUserID, ok := data["invitee_user_id"].(string)
		if !ok || inviteeUserID == "" {
			slog.Error("missing or invalid invitee_user_id in invitation sent event")
			return nil
		}

		domainID, ok := data["domain_id"].(string)
		if !ok || domainID == "" {
			slog.Error("missing or invalid domain_id in invitation sent event")
			return nil
		}

		roleID, _ := data["role_id"].(string)
		domainName, _ := data["domain_name"].(string)
		roleName, _ := data["role_name"].(string)

		if err := notifier.SendInvitationNotification(ctx, invitedBy, inviteeUserID, domainID, domainName, roleID, roleName); err != nil {
			slog.Error("failed to send invitation notification", "error", err)
		}

		return nil
	}
}

func handleInvitationAccepted(notifier notifications.Notifier) handleFunc {
	return func(ctx context.Context, event events.Event) error {
		data, err := event.Encode()
		if err != nil {
			slog.Error("failed to encode invitation accepted event", "error", err)
			return nil
		}

		invitedBy, ok := data["invited_by"].(string)
		if !ok || invitedBy == "" {
			slog.Error("missing or invalid invited_by in invitation accepted event")
			return nil
		}

		inviteeUserID, ok := data["invitee_user_id"].(string)
		if !ok || inviteeUserID == "" {
			slog.Error("missing or invalid invitee_user_id in invitation accepted event")
			return nil
		}

		domainID, ok := data["domain_id"].(string)
		if !ok || domainID == "" {
			slog.Error("missing or invalid domain_id in invitation accepted event")
			return nil
		}

		roleID, _ := data["role_id"].(string)
		domainName, _ := data["domain_name"].(string)
		roleName, _ := data["role_name"].(string)

		if err := notifier.SendAcceptanceNotification(ctx, invitedBy, inviteeUserID, domainID, domainName, roleID, roleName); err != nil {
			slog.Error("failed to send acceptance notification", "error", err)
		}

		return nil
	}
}

func handleInvitationRejected(notifier notifications.Notifier) handleFunc {
	return func(ctx context.Context, event events.Event) error {
		data, err := event.Encode()
		if err != nil {
			slog.Error("failed to encode invitation rejected event", "error", err)
			return nil
		}

		invitedBy, ok := data["invited_by"].(string)
		if !ok || invitedBy == "" {
			slog.Error("missing or invalid invited_by in invitation rejected event")
			return nil
		}

		inviteeUserID, ok := data["invitee_user_id"].(string)
		if !ok || inviteeUserID == "" {
			slog.Error("missing or invalid invitee_user_id in invitation rejected event")
			return nil
		}

		domainID, ok := data["domain_id"].(string)
		if !ok || domainID == "" {
			slog.Error("missing or invalid domain_id in invitation rejected event")
			return nil
		}

		roleID, _ := data["role_id"].(string)
		domainName, _ := data["domain_name"].(string)
		roleName, _ := data["role_name"].(string)

		if err := notifier.SendRejectionNotification(ctx, invitedBy, inviteeUserID, domainID, domainName, roleID, roleName); err != nil {
			slog.Error("failed to send rejection notification", "error", err)
		}

		return nil
	}
}

type handleFunc func(ctx context.Context, event events.Event) error

func (h handleFunc) Handle(ctx context.Context, event events.Event) error {
	return h(ctx, event)
}

func (h handleFunc) Cancel() error {
	return nil
}
