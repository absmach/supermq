// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/absmach/supermq/notifications"
	"github.com/absmach/supermq/pkg/events"
)

const (
	// Stream names.
	sendInvitationStream   = "events.supermq.invitation.send"
	acceptInvitationStream = "events.supermq.invitation.accept"
	rejectInvitationStream = "events.supermq.invitation.reject"
)

// Start starts consuming invitation events from the event store.
func Start(ctx context.Context, consumer string, sub events.Subscriber, notifier notifications.Notifier) error {
	handlers := []struct {
		stream       string
		notifType    notifications.NotificationType
		errorContext string
	}{
		{sendInvitationStream, notifications.Invitation, "invitation sent"},
		{acceptInvitationStream, notifications.Acceptance, "invitation accepted"},
		{rejectInvitationStream, notifications.Rejection, "invitation rejected"},
	}

	for _, h := range handlers {
		config := events.SubscriberConfig{
			Consumer: consumer,
			Stream:   h.stream,
			Handler:  handleInvitationEvent(notifier, h.notifType, h.errorContext),
		}
		if err := sub.Subscribe(ctx, config); err != nil {
			return err
		}
	}

	return nil
}

func handleInvitationEvent(notifier notifications.Notifier, notifType notifications.NotificationType, errorContext string) handleFunc {
	return func(ctx context.Context, event events.Event) error {
		notif, err := parseNotificationFromEvent(event, errorContext)
		if err != nil {
			return nil
		}

		notif.Type = notifType

		if err := notifier.Notify(ctx, notif); err != nil {
			slog.Error("failed to send notification", "error", err, "type", notifType, "context", errorContext)
		}

		return nil
	}
}

func parseNotificationFromEvent(event events.Event, errorContext string) (notifications.Notification, error) {
	data, err := event.Encode()
	if err != nil {
		slog.Error(fmt.Sprintf("failed to encode %s event", errorContext), "error", err)
		return notifications.Notification{}, err
	}

	invitedBy, ok := data["invited_by"].(string)
	if !ok || invitedBy == "" {
		slog.Error(fmt.Sprintf("missing or invalid invited_by in %s event", errorContext))
		return notifications.Notification{}, fmt.Errorf("missing or invalid invited_by")
	}

	inviteeUserID, ok := data["invitee_user_id"].(string)
	if !ok || inviteeUserID == "" {
		slog.Error(fmt.Sprintf("missing or invalid invitee_user_id in %s event", errorContext))
		return notifications.Notification{}, fmt.Errorf("missing or invalid invitee_user_id")
	}

	domainID, ok := data["domain_id"].(string)
	if !ok || domainID == "" {
		slog.Error(fmt.Sprintf("missing or invalid domain_id in %s event", errorContext))
		return notifications.Notification{}, fmt.Errorf("missing or invalid domain_id")
	}

	roleID, _ := data["role_id"].(string)
	domainName, _ := data["domain_name"].(string)
	roleName, _ := data["role_name"].(string)

	return notifications.Notification{
		InviterID:  invitedBy,
		InviteeID:  inviteeUserID,
		DomainID:   domainID,
		DomainName: domainName,
		RoleID:     roleID,
		RoleName:   roleName,
	}, nil
}

type handleFunc func(ctx context.Context, event events.Event) error

func (h handleFunc) Handle(ctx context.Context, event events.Event) error {
	return h(ctx, event)
}

func (h handleFunc) Cancel() error {
	return nil
}
