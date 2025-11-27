// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"time"

	"github.com/absmach/supermq/notifications"
	"github.com/go-kit/kit/metrics"
)

var _ notifications.Notifier = (*metricsMiddleware)(nil)

type metricsMiddleware struct {
	counter  metrics.Counter
	latency  metrics.Histogram
	notifier notifications.Notifier
}

// NewMetrics returns new notifier with metrics middleware.
func NewMetrics(notifier notifications.Notifier, counter metrics.Counter, latency metrics.Histogram) notifications.Notifier {
	return &metricsMiddleware{
		counter:  counter,
		latency:  latency,
		notifier: notifier,
	}
}

func (mm *metricsMiddleware) SendInvitationNotification(ctx context.Context, inviterID, inviteeID, domainID, domainName, roleID, roleName string) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "send_invitation_notification").Add(1)
		mm.latency.With("method", "send_invitation_notification").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.notifier.SendInvitationNotification(ctx, inviterID, inviteeID, domainID, domainName, roleID, roleName)
}

func (mm *metricsMiddleware) SendAcceptanceNotification(ctx context.Context, inviterID, inviteeID, domainID, domainName, roleID, roleName string) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "send_acceptance_notification").Add(1)
		mm.latency.With("method", "send_acceptance_notification").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.notifier.SendAcceptanceNotification(ctx, inviterID, inviteeID, domainID, domainName, roleID, roleName)
}

func (mm *metricsMiddleware) SendRejectionNotification(ctx context.Context, inviterID, inviteeID, domainID, domainName, roleID, roleName string) error {
	defer func(begin time.Time) {
		mm.counter.With("method", "send_rejection_notification").Add(1)
		mm.latency.With("method", "send_rejection_notification").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return mm.notifier.SendRejectionNotification(ctx, inviterID, inviteeID, domainID, domainName, roleID, roleName)
}
