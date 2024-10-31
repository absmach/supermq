// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

//go:build !test

package api

import (
	"context"
	"time"

	"github.com/absmach/supermq/auth"
	"github.com/absmach/supermq/pkg/authn"
	"github.com/absmach/supermq/pkg/policies"
	"github.com/go-kit/kit/metrics"
)

var _ auth.Service = (*metricsMiddleware)(nil)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	svc     auth.Service
}

// MetricsMiddleware instruments core service by tracking request count and latency.
func MetricsMiddleware(svc auth.Service, counter metrics.Counter, latency metrics.Histogram) auth.Service {
	return &metricsMiddleware{
		counter: counter,
		latency: latency,
		svc:     svc,
	}
}

func (ms *metricsMiddleware) Issue(ctx context.Context, token string, key auth.Key) (auth.Token, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "issue_key").Add(1)
		ms.latency.With("method", "issue_key").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.Issue(ctx, token, key)
}

func (ms *metricsMiddleware) Revoke(ctx context.Context, token, id string) error {
	defer func(begin time.Time) {
		ms.counter.With("method", "revoke_key").Add(1)
		ms.latency.With("method", "revoke_key").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.Revoke(ctx, token, id)
}

func (ms *metricsMiddleware) RetrieveKey(ctx context.Context, token, id string) (auth.Key, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "retrieve_key").Add(1)
		ms.latency.With("method", "retrieve_key").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.RetrieveKey(ctx, token, id)
}

func (ms *metricsMiddleware) Identify(ctx context.Context, token string) (auth.Key, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "identify").Add(1)
		ms.latency.With("method", "identify").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.Identify(ctx, token)
}

func (ms *metricsMiddleware) Authorize(ctx context.Context, pr policies.Policy) error {
	defer func(begin time.Time) {
		ms.counter.With("method", "authorize").Add(1)
		ms.latency.With("method", "authorize").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.Authorize(ctx, pr)
}

func (ms *metricsMiddleware) CreatePAT(ctx context.Context, session authn.Session, name, description string, duration time.Duration, scope auth.Scope) (auth.PAT, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "create_pat").Add(1)
		ms.latency.With("method", "create_pat").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.CreatePAT(ctx, session, name, description, duration, scope)
}

func (ms *metricsMiddleware) UpdatePATName(ctx context.Context, session authn.Session, patID, name string) (auth.PAT, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "update_pat_name").Add(1)
		ms.latency.With("method", "update_pat_name").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.UpdatePATName(ctx, session, patID, name)
}

func (ms *metricsMiddleware) UpdatePATDescription(ctx context.Context, session authn.Session, patID, description string) (auth.PAT, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "update_pat_description").Add(1)
		ms.latency.With("method", "update_pat_description").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.UpdatePATDescription(ctx, session, patID, description)
}

func (ms *metricsMiddleware) RetrievePAT(ctx context.Context, session authn.Session, patID string) (auth.PAT, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "retrieve_pat").Add(1)
		ms.latency.With("method", "retrieve_pat").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.RetrievePAT(ctx, session, patID)
}

func (ms *metricsMiddleware) ListPATS(ctx context.Context, session authn.Session, pm auth.PATSPageMeta) (auth.PATSPage, error) {
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

func (ms *metricsMiddleware) ResetPATSecret(ctx context.Context, session authn.Session, patID string, duration time.Duration) (auth.PAT, error) {
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

func (ms *metricsMiddleware) AddPATScopeEntry(ctx context.Context, session authn.Session, patID string, platformEntityType auth.PlatformEntityType, optionalDomainID string, optionalDomainEntityType auth.DomainEntityType, operation auth.OperationType, entityIDs ...string) (auth.Scope, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "add_pat_scope_entry").Add(1)
		ms.latency.With("method", "add_pat_scope_entry").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.AddPATScopeEntry(ctx, session, patID, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
}

func (ms *metricsMiddleware) RemovePATScopeEntry(ctx context.Context, session authn.Session, patID string, platformEntityType auth.PlatformEntityType, optionalDomainID string, optionalDomainEntityType auth.DomainEntityType, operation auth.OperationType, entityIDs ...string) (auth.Scope, error) {
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

func (ms *metricsMiddleware) IdentifyPAT(ctx context.Context, paToken string) (auth.PAT, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "identify_pat").Add(1)
		ms.latency.With("method", "identify_pat").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.IdentifyPAT(ctx, paToken)
}

func (ms *metricsMiddleware) AuthorizePAT(ctx context.Context, paToken string, platformEntityType auth.PlatformEntityType, optionalDomainID string, optionalDomainEntityType auth.DomainEntityType, operation auth.OperationType, entityIDs ...string) error {
	defer func(begin time.Time) {
		ms.counter.With("method", "authorize_pat").Add(1)
		ms.latency.With("method", "authorize_pat").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.AuthorizePAT(ctx, paToken, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
}

func (ms *metricsMiddleware) CheckPAT(ctx context.Context, userID, patID string, platformEntityType auth.PlatformEntityType, optionalDomainID string, optionalDomainEntityType auth.DomainEntityType, operation auth.OperationType, entityIDs ...string) error {
	defer func(begin time.Time) {
		ms.counter.With("method", "check_pat").Add(1)
		ms.latency.With("method", "check_pat").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.CheckPAT(ctx, userID, patID, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
}
