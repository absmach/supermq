// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/mainflux/mainflux/things/policies"
)

const (
	streamID  = "mainflux.clients.things"
	streamLen = 1000
)

var _ policies.Service = (*eventStore)(nil)

type eventStore struct {
	svc    policies.Service
	client *redis.Client
}

// NewEventStoreMiddleware returns wrapper around policy service that sends
// events to event store.
func NewEventStoreMiddleware(svc policies.Service, client *redis.Client) policies.Service {
	return eventStore{
		svc:    svc,
		client: client,
	}
}

func (es eventStore) Authorize(ctx context.Context, ar policies.AccessRequest, entity string) (string, error) {
	id, err := es.svc.Authorize(ctx, ar, entity)
	if err != nil {
		return "", err
	}

	event := policyEvent{
		entityType: entity,
		groupID:    ar.Object,
		clientID:   id,
		actions:    []string{ar.Action},
	}

	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       event.Encode(authorize),
	}
	if err := es.client.XAdd(ctx, record).Err(); err != nil {
		return id, err
	}

	return id, nil
}

func (es eventStore) AddPolicy(ctx context.Context, token string, policy policies.Policy) (policies.Policy, error) {
	policy, err := es.svc.AddPolicy(ctx, token, policy)
	if err != nil {
		return policies.Policy{}, err
	}

	event := policyEvent{
		groupID:  policy.Object,
		clientID: policy.Subject,
		actions:  policy.Actions,
	}

	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       event.Encode(addPolicy),
	}
	if err := es.client.XAdd(ctx, record).Err(); err != nil {
		return policies.Policy{}, err
	}

	return policy, nil
}

func (es eventStore) UpdatePolicy(ctx context.Context, token string, policy policies.Policy) (policies.Policy, error) {
	policy, err := es.svc.UpdatePolicy(ctx, token, policy)
	if err != nil {
		return policies.Policy{}, err
	}

	event := policyEvent{
		groupID:  policy.Object,
		clientID: policy.Subject,
		actions:  policy.Actions,
	}

	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       event.Encode(updatePolicy),
	}
	if err := es.client.XAdd(ctx, record).Err(); err != nil {
		return policies.Policy{}, err
	}

	return policy, nil
}

func (es eventStore) ListPolicies(ctx context.Context, token string, page policies.Page) (policies.PolicyPage, error) {
	policypage, err := es.svc.ListPolicies(ctx, token, page)
	if err != nil {
		return policies.PolicyPage{}, err
	}

	event := policyEvent{
		groupID:  page.Object,
		clientID: page.Subject,
		actions:  []string{page.Action},
	}

	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       event.Encode(listPolicy),
	}
	if err := es.client.XAdd(ctx, record).Err(); err != nil {
		return policies.PolicyPage{}, err
	}

	return policypage, nil
}

func (es eventStore) DeletePolicy(ctx context.Context, token string, policy policies.Policy) error {
	if err := es.svc.DeletePolicy(ctx, token, policy); err != nil {
		return err
	}

	event := policyEvent{
		groupID:  policy.Object,
		clientID: policy.Subject,
		actions:  policy.Actions,
	}

	record := &redis.XAddArgs{
		Stream:       streamID,
		MaxLenApprox: streamLen,
		Values:       event.Encode(deletePolicy),
	}
	if err := es.client.XAdd(ctx, record).Err(); err != nil {
		return err
	}

	return nil
}
