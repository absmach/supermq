// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"

	"github.com/absmach/supermq/notifications"
	smqTracing "github.com/absmach/supermq/pkg/tracing"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var _ notifications.Notifier = (*tracing)(nil)

type tracing struct {
	tracer   trace.Tracer
	notifier notifications.Notifier
}

// NewTracing returns a new notifier with tracing capabilities.
func NewTracing(notifier notifications.Notifier, tracer trace.Tracer) notifications.Notifier {
	return &tracing{tracer, notifier}
}

func (tm *tracing) SendInvitationNotification(ctx context.Context, inviterID, inviteeID, domainID, domainName, roleID, roleName string) error {
	ctx, span := smqTracing.StartSpan(ctx, tm.tracer, "send_invitation_notification", trace.WithAttributes(
		attribute.String("inviter_id", inviterID),
		attribute.String("invitee_id", inviteeID),
		attribute.String("domain_id", domainID),
		attribute.String("domain_name", domainName),
		attribute.String("role_id", roleID),
		attribute.String("role_name", roleName),
	))
	defer span.End()

	return tm.notifier.SendInvitationNotification(ctx, inviterID, inviteeID, domainID, domainName, roleID, roleName)
}

func (tm *tracing) SendAcceptanceNotification(ctx context.Context, inviterID, inviteeID, domainID, domainName, roleID, roleName string) error {
	ctx, span := smqTracing.StartSpan(ctx, tm.tracer, "send_acceptance_notification", trace.WithAttributes(
		attribute.String("inviter_id", inviterID),
		attribute.String("invitee_id", inviteeID),
		attribute.String("domain_id", domainID),
		attribute.String("domain_name", domainName),
		attribute.String("role_id", roleID),
		attribute.String("role_name", roleName),
	))
	defer span.End()

	return tm.notifier.SendAcceptanceNotification(ctx, inviterID, inviteeID, domainID, domainName, roleID, roleName)
}

func (tm *tracing) SendRejectionNotification(ctx context.Context, inviterID, inviteeID, domainID, domainName, roleID, roleName string) error {
	ctx, span := smqTracing.StartSpan(ctx, tm.tracer, "send_rejection_notification", trace.WithAttributes(
		attribute.String("inviter_id", inviterID),
		attribute.String("invitee_id", inviteeID),
		attribute.String("domain_id", domainID),
		attribute.String("domain_name", domainName),
		attribute.String("role_id", roleID),
		attribute.String("role_name", roleName),
	))
	defer span.End()

	return tm.notifier.SendRejectionNotification(ctx, inviterID, inviteeID, domainID, domainName, roleID, roleName)
}
