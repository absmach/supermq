// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"context"

	"github.com/absmach/magistrala/pkg/domains"
	entityRolesEvents "github.com/absmach/magistrala/pkg/entityroles/events"
	"github.com/absmach/magistrala/pkg/events"
	"github.com/absmach/magistrala/pkg/events/store"
)

const streamID = "magistrala.domains"

var _ domains.Service = (*eventStore)(nil)

type eventStore struct {
	events.Publisher
	svc domains.Service
	entityRolesEvents.RolesSvcEventStoreMiddleware
}

// NewEventStoreMiddleware returns wrapper around auth service that sends
// events to event store.
func NewEventStoreMiddleware(ctx context.Context, svc domains.Service, url string) (domains.Service, error) {
	publisher, err := store.NewPublisher(ctx, url, streamID)
	if err != nil {
		return nil, err
	}

	rolesSvcEventStoreMiddleware := entityRolesEvents.NewRolesSvcEventStoreMiddleware("domains", svc, publisher)

	return &eventStore{
		svc:                          svc,
		Publisher:                    publisher,
		RolesSvcEventStoreMiddleware: rolesSvcEventStoreMiddleware,
	}, nil
}

func (es *eventStore) CreateDomain(ctx context.Context, token string, domain domains.Domain) (domains.Domain, error) {
	domain, err := es.svc.CreateDomain(ctx, token, domain)
	if err != nil {
		return domain, err
	}

	event := createDomainEvent{
		domain,
	}

	if err := es.Publish(ctx, event); err != nil {
		return domain, err
	}

	return domain, nil
}

func (es *eventStore) RetrieveDomain(ctx context.Context, token, id string) (domains.Domain, error) {
	domain, err := es.svc.RetrieveDomain(ctx, token, id)
	if err != nil {
		return domain, err
	}

	event := retrieveDomainEvent{
		domain,
	}

	if err := es.Publish(ctx, event); err != nil {
		return domain, err
	}

	return domain, nil
}

func (es *eventStore) UpdateDomain(ctx context.Context, token, id string, d domains.DomainReq) (domains.Domain, error) {
	domain, err := es.svc.UpdateDomain(ctx, token, id, d)
	if err != nil {
		return domain, err
	}

	event := updateDomainEvent{
		domain,
	}

	if err := es.Publish(ctx, event); err != nil {
		return domain, err
	}

	return domain, nil
}

func (es *eventStore) EnableDomain(ctx context.Context, token, id string) (domains.Domain, error) {
	domain, err := es.svc.EnableDomain(ctx, token, id)
	if err != nil {
		return domain, err
	}

	event := enableDomainEvent{
		domainID:  id,
		updatedAt: domain.UpdatedAt,
		updatedBy: domain.UpdatedBy,
	}

	if err := es.Publish(ctx, event); err != nil {
		return domain, err
	}

	return domain, nil
}

func (es *eventStore) DisableDomain(ctx context.Context, token, id string) (domains.Domain, error) {
	domain, err := es.svc.DisableDomain(ctx, token, id)
	if err != nil {
		return domain, err
	}

	event := disableDomainEvent{
		domainID:  id,
		updatedAt: domain.UpdatedAt,
		updatedBy: domain.UpdatedBy,
	}

	if err := es.Publish(ctx, event); err != nil {
		return domain, err
	}

	return domain, nil
}

func (es *eventStore) FreezeDomain(ctx context.Context, token, id string) (domains.Domain, error) {
	domain, err := es.svc.FreezeDomain(ctx, token, id)
	if err != nil {
		return domain, err
	}

	event := freezeDomainEvent{
		domainID:  id,
		updatedAt: domain.UpdatedAt,
		updatedBy: domain.UpdatedBy,
	}

	if err := es.Publish(ctx, event); err != nil {
		return domain, err
	}

	return domain, nil
}

func (es *eventStore) ListDomains(ctx context.Context, token string, p domains.Page) (domains.DomainsPage, error) {
	dp, err := es.svc.ListDomains(ctx, token, p)
	if err != nil {
		return dp, err
	}

	event := listDomainsEvent{
		p, dp.Total,
	}

	if err := es.Publish(ctx, event); err != nil {
		return dp, err
	}

	return dp, nil
}
