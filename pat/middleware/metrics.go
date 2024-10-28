// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"time"

	"github.com/absmach/magistrala/pat"
	"github.com/absmach/magistrala/pkg/authn"
	"github.com/go-kit/kit/metrics"
)

var _ pat.Service = (*metricsMiddleware)(nil)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	svc     pat.Service
}

// MetricsMiddleware instruments core service by tracking request count and latency.
func MetricsMiddleware(svc pat.Service, counter metrics.Counter, latency metrics.Histogram) pat.Service {
	return &metricsMiddleware{
		counter: counter,
		latency: latency,
		svc:     svc,
	}
}

func (ms *metricsMiddleware) CreatePAT(ctx context.Context, session authn.Session, name, description string, duration time.Duration, scope pat.Scope) (pat.PAT, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "create_pat").Add(1)
		ms.latency.With("method", "create_pat").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.CreatePAT(ctx, session, name, description, duration, scope)
}

func (ms *metricsMiddleware) UpdatePATName(ctx context.Context, session authn.Session, patID, name string) (pat.PAT, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "update_pat_name").Add(1)
		ms.latency.With("method", "update_pat_name").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.UpdatePATName(ctx, session, patID, name)
}

func (ms *metricsMiddleware) UpdatePATDescription(ctx context.Context, session authn.Session, patID, description string) (pat.PAT, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "update_pat_description").Add(1)
		ms.latency.With("method", "update_pat_description").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.UpdatePATDescription(ctx, session, patID, description)
}

func (ms *metricsMiddleware) RetrievePAT(ctx context.Context, session authn.Session, patID string) (pat.PAT, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "retrieve_pat").Add(1)
		ms.latency.With("method", "retrieve_pat").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.RetrievePAT(ctx, session, patID)
}

func (ms *metricsMiddleware) ListPATS(ctx context.Context, session authn.Session, pm pat.PATSPageMeta) (pat.PATSPage, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "list_pats").Add(1)
		ms.latency.With("method", "list_pats").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.ListPATS(ctx, session, pm)
}

func (ms *metricsMiddleware) DeletePAT(ctx context.Context, session authn.Session, patID string) error {
	defer func(begin time.Time) {
		ms.counter.With("method", "delete_pat").Add(1)
		ms.latency.With("method", "delete_pat").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.DeletePAT(ctx, session, patID)
}

func (ms *metricsMiddleware) ResetPATSecret(ctx context.Context, session authn.Session, patID string, duration time.Duration) (pat.PAT, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "reset_pat_secret").Add(1)
		ms.latency.With("method", "reset_pat_secret").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.ResetPATSecret(ctx, session, patID, duration)
}

func (ms *metricsMiddleware) RevokePATSecret(ctx context.Context, session authn.Session, patID string) error {
	defer func(begin time.Time) {
		ms.counter.With("method", "revoke_pat_secret").Add(1)
		ms.latency.With("method", "revoke_pat_secret").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.RevokePATSecret(ctx, session, patID)
}

func (ms *metricsMiddleware) AddPATScopeEntry(ctx context.Context, session authn.Session, patID string, platformEntityType pat.PlatformEntityType, optionalDomainID string, optionalDomainEntityType pat.DomainEntityType, operation pat.OperationType, entityIDs ...string) (pat.Scope, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "add_pat_scope_entry").Add(1)
		ms.latency.With("method", "add_pat_scope_entry").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.AddPATScopeEntry(ctx, session, patID, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
}

func (ms *metricsMiddleware) RemovePATScopeEntry(ctx context.Context, session authn.Session, patID string, platformEntityType pat.PlatformEntityType, optionalDomainID string, optionalDomainEntityType pat.DomainEntityType, operation pat.OperationType, entityIDs ...string) (pat.Scope, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "remove_pat_scope_entry").Add(1)
		ms.latency.With("method", "remove_pat_scope_entry").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.RemovePATScopeEntry(ctx, session, patID, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
}

func (ms *metricsMiddleware) ClearPATAllScopeEntry(ctx context.Context, session authn.Session, patID string) error {
	defer func(begin time.Time) {
		ms.counter.With("method", "clear_pat_all_scope_entry").Add(1)
		ms.latency.With("method", "clear_pat_all_scope_entry").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.ClearPATAllScopeEntry(ctx, session, patID)
}

func (ms *metricsMiddleware) IdentifyPAT(ctx context.Context, paToken string) (pat.PAT, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "identify_pat").Add(1)
		ms.latency.With("method", "identify_pat").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.IdentifyPAT(ctx, paToken)
}

func (ms *metricsMiddleware) AuthorizePAT(ctx context.Context, paToken string, platformEntityType pat.PlatformEntityType, optionalDomainID string, optionalDomainEntityType pat.DomainEntityType, operation pat.OperationType, entityIDs ...string) error {
	defer func(begin time.Time) {
		ms.counter.With("method", "authorize_pat").Add(1)
		ms.latency.With("method", "authorize_pat").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.AuthorizePAT(ctx, paToken, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
}

func (ms *metricsMiddleware) CheckPAT(ctx context.Context, userID, patID string, platformEntityType pat.PlatformEntityType, optionalDomainID string, optionalDomainEntityType pat.DomainEntityType, operation pat.OperationType, entityIDs ...string) error {
	defer func(begin time.Time) {
		ms.counter.With("method", "check_pat").Add(1)
		ms.latency.With("method", "check_pat").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.CheckPAT(ctx, userID, patID, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
}
