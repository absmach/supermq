// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/mainflux/mainflux/internal/groups"
	"github.com/mainflux/mainflux/users"
)

var _ users.Service = (*metricsMiddleware)(nil)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	svc     users.Service
}

// MetricsMiddleware instruments core service by tracking request count and latency.
func MetricsMiddleware(svc users.Service, counter metrics.Counter, latency metrics.Histogram) users.Service {
	return &metricsMiddleware{
		counter: counter,
		latency: latency,
		svc:     svc,
	}
}

func (ms *metricsMiddleware) Register(ctx context.Context, token string, user users.User) (string, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "register").Add(1)
		ms.latency.With("method", "register").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.Register(ctx, token, user)
}

func (ms *metricsMiddleware) Login(ctx context.Context, user users.User) (string, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "login").Add(1)
		ms.latency.With("method", "login").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.Login(ctx, user)
}

func (ms *metricsMiddleware) ViewUser(ctx context.Context, token, id string) (users.User, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "view_user").Add(1)
		ms.latency.With("method", "view_user").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.ViewUser(ctx, token, id)
}

func (ms *metricsMiddleware) ViewProfile(ctx context.Context, token string) (users.User, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "view_profile").Add(1)
		ms.latency.With("method", "view_profile").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.ViewProfile(ctx, token)
}

func (ms *metricsMiddleware) ListUsers(ctx context.Context, token string, offset, limit uint64, email string, um users.Metadata) (users.UserPage, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "list_users").Add(1)
		ms.latency.With("method", "list_users").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.ListUsers(ctx, token, offset, limit, email, um)
}

func (ms *metricsMiddleware) UpdateUser(ctx context.Context, token string, u users.User) (err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "update_user").Add(1)
		ms.latency.With("method", "update_user").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.UpdateUser(ctx, token, u)
}

func (ms *metricsMiddleware) GenerateResetToken(ctx context.Context, email, host string) error {
	defer func(begin time.Time) {
		ms.counter.With("method", "generate_reset_token").Add(1)
		ms.latency.With("method", "generate_reset_token").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.GenerateResetToken(ctx, email, host)
}

func (ms *metricsMiddleware) ChangePassword(ctx context.Context, email, password, oldPassword string) error {
	defer func(begin time.Time) {
		ms.counter.With("method", "change_password").Add(1)
		ms.latency.With("method", "change_password").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.ChangePassword(ctx, email, password, oldPassword)
}

func (ms *metricsMiddleware) ResetPassword(ctx context.Context, email, password string) error {
	defer func(begin time.Time) {
		ms.counter.With("method", "reset_password").Add(1)
		ms.latency.With("method", "reset_password").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.ResetPassword(ctx, email, password)
}

func (ms *metricsMiddleware) SendPasswordReset(ctx context.Context, host, email, token string) error {
	defer func(begin time.Time) {
		ms.counter.With("method", "send_password_reset").Add(1)
		ms.latency.With("method", "send_password_reset").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.SendPasswordReset(ctx, host, email, token)
}

func (ms *metricsMiddleware) CreateGroup(ctx context.Context, token string, group groups.Group) (gr groups.Group, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "create_group").Add(1)
		ms.latency.With("method", "create_group").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.CreateGroup(ctx, token, group)
}

func (ms *metricsMiddleware) UpdateGroup(ctx context.Context, token string, group groups.Group) (gr groups.Group, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "update_group").Add(1)
		ms.latency.With("method", "update_group").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.UpdateGroup(ctx, token, group)
}

func (ms *metricsMiddleware) RemoveGroup(ctx context.Context, token string, id string) (err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "remove_group").Add(1)
		ms.latency.With("method", "remove_group").Observe(time.Since(begin).Seconds())
	}(time.Now())
	return ms.svc.RemoveGroup(ctx, token, id)
}

func (ms *metricsMiddleware) ViewGroup(ctx context.Context, token, id string) (group groups.Group, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "view_group").Add(1)
		ms.latency.With("method", "view_group").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.ViewGroup(ctx, token, id)
}

func (ms *metricsMiddleware) ListGroups(ctx context.Context, token string, pm groups.PageMetadata) (gp groups.GroupPage, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "list_groups").Add(1)
		ms.latency.With("method", "list_groups").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.ListGroups(ctx, token, pm)
}

func (ms *metricsMiddleware) ListParents(ctx context.Context, token, childID string, pm groups.PageMetadata) (gp groups.GroupPage, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "parents").Add(1)
		ms.latency.With("method", "parents").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.ListParents(ctx, token, childID, pm)
}

func (ms *metricsMiddleware) ListChildren(ctx context.Context, token, parentID string, pm groups.PageMetadata) (gp groups.GroupPage, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "list_children").Add(1)
		ms.latency.With("method", "list_children").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.ListChildren(ctx, token, parentID, pm)
}

// func (ms *metricsMiddleware) ListMembers(ctx context.Context, token, groupID string, pm groups.PageMetadata) (up users.UserPage, err error) {
// 	defer func(begin time.Time) {
// 		ms.counter.With("method", "list_members").Add(1)
// 		ms.latency.With("method", "list_members").Observe(time.Since(begin).Seconds())
// 	}(time.Now())

// 	return ms.svc.ListMembers(ctx, token, groupID, pm)
// }

func (ms *metricsMiddleware) ListMemberships(ctx context.Context, token, memberID string, pm groups.PageMetadata) (gp groups.GroupPage, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "list_memberships").Add(1)
		ms.latency.With("method", "list_memberships").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.ListMemberships(ctx, token, memberID, pm)
}

func (ms *metricsMiddleware) Assign(ctx context.Context, token, groupID string, memberIDs ...string) (err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "assign").Add(1)
		ms.latency.With("method", "assign").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.Assign(ctx, token, groupID, memberIDs...)
}

func (ms *metricsMiddleware) Unassign(ctx context.Context, token, groupID string, memberIDs ...string) (err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "unassign").Add(1)
		ms.latency.With("method", "unassign").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.Unassign(ctx, token, groupID, memberIDs...)
}

// func (ms *metricsMiddleware) AssignGroupAccessRights(ctx context.Context, token, thingGroupID, userGroupID string) error {
// 	defer func(begin time.Time) {
// 		ms.counter.With("method", "share_group_access").Add(1)
// 		ms.latency.With("method", "share_group_access").Observe(time.Since(begin).Seconds())
// 	}(time.Now())

// 	return ms.svc.AssignGroupAccessRights(ctx, token, thingGroupID, userGroupID)
// }
