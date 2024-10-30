// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"context"
	"time"

	"github.com/absmach/magistrala/pat"
	"github.com/absmach/magistrala/pkg/authn"
	"github.com/absmach/magistrala/pkg/events"
	"github.com/absmach/magistrala/pkg/events/store"
)

const streamID = "magistrala.pat"

var _ pat.Service = (*eventStore)(nil)


type eventStore struct {
	events.Publisher
	svc pat.Service
}

// NewEventStoreMiddleware returns wrapper around pat service that sends
// events to event store.
func NewEventStoreMiddleware(ctx context.Context, svc pat.Service, url string) (pat.Service, error) {
	publisher, err := store.NewPublisher(ctx, url, streamID)
	if err != nil {
		return nil, err
	}

	return &eventStore{
		svc:       svc,
		Publisher: publisher,
	}, nil
}

func (es *eventStore) CreatePAT(ctx context.Context, session authn.Session, name, description string, duration time.Duration, scope pat.Scope) (pat.PAT, error) {
	return es.svc.CreatePAT(ctx, session, name, description, duration, scope)
}

func (es *eventStore) UpdatePATName(ctx context.Context, session authn.Session, patID, name string) (pat.PAT, error) {
	return es.svc.UpdatePATName(ctx, session, patID, name)
}

func (es *eventStore) UpdatePATDescription(ctx context.Context, session authn.Session, patID, description string) (pat.PAT, error) {
	return es.svc.UpdatePATDescription(ctx, session, patID, description)
}

func (es *eventStore) RetrievePAT(ctx context.Context, session authn.Session, patID string) (pat.PAT, error) {
	return es.svc.RetrievePAT(ctx, session, patID)
}

func (es *eventStore) ListPATS(ctx context.Context, session authn.Session, pm pat.PATSPageMeta) (pat.PATSPage, error) {
	return es.svc.ListPATS(ctx, session, pm)
}

func (es *eventStore) DeletePAT(ctx context.Context, session authn.Session, patID string) error {
	return es.svc.DeletePAT(ctx, session, patID)
}

func (es *eventStore) ResetPATSecret(ctx context.Context, session authn.Session, patID string, duration time.Duration) (pat.PAT, error) {
	return es.svc.ResetPATSecret(ctx, session, patID, duration)
}

func (es *eventStore) RevokePATSecret(ctx context.Context, session authn.Session, patID string) error {
	return es.svc.RevokePATSecret(ctx, session, patID)
}

func (es *eventStore) AddPATScopeEntry(ctx context.Context, session authn.Session, patID string, platformEntityType pat.PlatformEntityType, optionalDomainID string, optionalDomainEntityType pat.DomainEntityType, operation pat.OperationType, entityIDs ...string) (pat.Scope, error) {
	return es.svc.AddPATScopeEntry(ctx, session, patID, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
}

func (es *eventStore) RemovePATScopeEntry(ctx context.Context, session authn.Session, patID string, platformEntityType pat.PlatformEntityType, optionalDomainID string, optionalDomainEntityType pat.DomainEntityType, operation pat.OperationType, entityIDs ...string) (pat.Scope, error) {
	return es.svc.RemovePATScopeEntry(ctx, session, patID, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
}

func (es *eventStore) ClearPATAllScopeEntry(ctx context.Context, session authn.Session, patID string) error {
	return es.svc.ClearPATAllScopeEntry(ctx, session, patID)
}

func (es *eventStore) IdentifyPAT(ctx context.Context, paToken string) (pat.PAT, error) {
	return es.svc.IdentifyPAT(ctx, paToken)
}

func (es *eventStore) AuthorizePAT(ctx context.Context, paToken string, platformEntityType pat.PlatformEntityType, optionalDomainID string, optionalDomainEntityType pat.DomainEntityType, operation pat.OperationType, entityIDs ...string) error {
	return es.svc.AuthorizePAT(ctx, paToken, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
}

func (es *eventStore) CheckPAT(ctx context.Context, userID, patID string, platformEntityType pat.PlatformEntityType, optionalDomainID string, optionalDomainEntityType pat.DomainEntityType, operation pat.OperationType, entityIDs ...string) error {
	return es.svc.CheckPAT(ctx, userID, patID, platformEntityType, optionalDomainID, optionalDomainEntityType, operation, entityIDs...)
}