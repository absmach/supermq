// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/absmach/supermq/auditlogs"
	"github.com/absmach/supermq/pkg/errors"
	repoerr "github.com/absmach/supermq/pkg/errors/repository"
	"github.com/absmach/supermq/pkg/postgres"
	"github.com/jmoiron/sqlx"
)

type auditLogsRepo struct {
	db postgres.Database
}

func NewRepository(db postgres.Database) auditlogs.Repository {
	return &auditLogsRepo{
		db: db,
	}
}

func (ar *auditLogsRepo) Save(ctx context.Context, l auditlogs.AuditLog) error {
	q := `INSERT INTO audit_logs (id, request_id, domain_id, occurred_at, actor_id, current_state, previous_state, state_attributes, entity_id, entity_type, metadata)
	VALUES (:id, :request_id, :domain_id, :occurred_at, :actor_id, :current_state, :previous_state, :state_attributes, :entity_id, :entity_type, :metadata)`

	dbal, err := toDBAuditLog(l)
	if err != nil {
		return errors.Wrap(repoerr.ErrCreateEntity, err)
	}
	if _, err := ar.db.NamedQueryContext(ctx, q, dbal); err != nil {
		return errors.Wrap(repoerr.ErrCreateEntity, err)
	}

	return nil
}

func (ar *auditLogsRepo) RetrieveByID(ctx context.Context, id string) (auditlogs.AuditLog, error) {
	q := `SELECT id, request_id, domain_id, occurred_at, actor_id, current_state, previous_state, state_attributes, entity_id, entity_type, metadata
	FROM audit logs WHERE id = :id`

	dbal := dbAuditLog{
		ID: id,
	}

	rows, err := ar.db.NamedQueryContext(ctx, q, dbal)
	if err != nil {
		return auditlogs.AuditLog{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}
	defer rows.Close()

	dbal = dbAuditLog{}
	if rows.Next() {
		if err := rows.StructScan(&dbal); err != nil {
			return auditlogs.AuditLog{}, errors.Wrap(repoerr.ErrViewEntity, err)
		}

		log, err := toAuditLog(dbal)
		if err != nil {
			return auditlogs.AuditLog{}, errors.Wrap(repoerr.ErrViewEntity, err)
		}

		return log, nil
	}

	return auditlogs.AuditLog{}, repoerr.ErrNotFound
}

func (ar *auditLogsRepo) RetrieveAll(ctx context.Context, pm auditlogs.Page) (auditlogs.AuditLogPage, error) {
	query, err := buildPageQuery(pm)
	if err != nil {
		return auditlogs.AuditLogPage{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}

	q := fmt.Sprintf(`SELECT l.id as id, l.request_id as request_id, l.domain_id as domain_id, l.occurred_at as occurred_at, l.actor_id as actor_id, l.current_state as current_state, l.previous_state as previous_state, l.state_attributes as state_attributes, l.entity_id as entity_id, l.entity_type as entity_type, l.metadata as metadata
	FROM audit_logs l %s ORDER BY l.occurred_at %s OFFSET :offset LIMIT :limit`, query, pm.Order)

	dbpm, err := toDBAuditLogsPage(pm)
	if err != nil {
		return auditlogs.AuditLogPage{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}
	rows, err := ar.db.NamedQueryContext(ctx, q, dbpm)
	if err != nil {
		return auditlogs.AuditLogPage{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}
	defer rows.Close()

	items, err := ar.processRows(rows)
	if err != nil {
		return auditlogs.AuditLogPage{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}

	cq := fmt.Sprintf(`SELECT COUNT(*) as total_count
	FROM (
	SELECT DISTINCT l.id, l.request_id, l.domain_id, l.occurred_at, l.actor_id, l.current_state, l.previous_state, l.state_attributes, l.entity_id, l.entity_type, l.metadata
	FROM audit_logs l %s) as subquery`, query)

	total, err := postgres.Total(ctx, ar.db, cq, dbpm)
	if err != nil {
		return auditlogs.AuditLogPage{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}

	page := auditlogs.AuditLogPage{
		Page: pm,
	}
	page.Total = total
	page.Logs = items

	return page, nil
}

type dbAuditLog struct {
	ID              string    `db:"id"`
	RequestID       string    `db:"request_id"`
	DomainID        string    `db:"domain_id"`
	OccurredAt      time.Time `db:"occurred_at"`
	ActorID         string    `db:"actor_id"`
	CurrentState    string    `db:"current_state"`
	PreviousState   string    `db:"previous_state"`
	StateAttributes []byte    `db:"state_attributes"`
	EntityID        string    `db:"entity_id"`
	EntityType      string    `db:"entity_type"`
	Metadata        []byte    `db:"metadata"`
}

func toDBAuditLog(log auditlogs.AuditLog) (dbAuditLog, error) {
	stateAttr := []byte("{}")
	if len(log.StateAttributes) > 0 {
		b, err := json.Marshal(log.StateAttributes)
		if err != nil {
			return dbAuditLog{}, errors.Wrap(errors.ErrMalformedEntity, err)
		}
		stateAttr = b
	}
	metadata := []byte("{}")
	if len(log.Metadata) > 0 {
		b, err := json.Marshal(log.Metadata)
		if err != nil {
			return dbAuditLog{}, errors.Wrap(errors.ErrMalformedEntity, err)
		}
		metadata = b
	}
	return dbAuditLog{
		ID:              log.ID,
		RequestID:       log.RequestID,
		DomainID:        log.DomainID,
		OccurredAt:      log.OccurredAt,
		ActorID:         log.ActorID,
		CurrentState:    log.CurrentState.String(),
		PreviousState:   log.PreviousState.String(),
		StateAttributes: stateAttr,
		EntityID:        log.EntityID,
		EntityType:      log.EntityType.String(),
		Metadata:        metadata,
	}, nil
}

func toAuditLog(log dbAuditLog) (auditlogs.AuditLog, error) {
	var stateAttr auditlogs.Metadata
	if log.StateAttributes != nil {
		if err := json.Unmarshal(log.StateAttributes, &stateAttr); err != nil {
			return auditlogs.AuditLog{}, errors.Wrap(errors.ErrMalformedEntity, err)
		}
	}
	var metadata auditlogs.Metadata
	if log.Metadata != nil {
		if err := json.Unmarshal(log.Metadata, &metadata); err != nil {
			return auditlogs.AuditLog{}, errors.Wrap(errors.ErrMalformedEntity, err)
		}

	}
	cs, err := auditlogs.ToEntityState(log.CurrentState)
	if err != nil {
		return auditlogs.AuditLog{}, errors.Wrap(errors.ErrMalformedEntity, err)
	}
	ps, err := auditlogs.ToEntityState(log.PreviousState)
	if err != nil {
		return auditlogs.AuditLog{}, errors.Wrap(errors.ErrMalformedEntity, err)
	}
	et, err := auditlogs.ToEntityType(log.EntityType)
	if err != nil {
		return auditlogs.AuditLog{}, errors.Wrap(errors.ErrMalformedEntity, err)
	}
	return auditlogs.AuditLog{
		ID:              log.ID,
		RequestID:       log.RequestID,
		DomainID:        log.DomainID,
		OccurredAt:      log.OccurredAt,
		ActorID:         log.ActorID,
		CurrentState:    cs,
		PreviousState:   ps,
		StateAttributes: stateAttr,
		EntityID:        log.EntityID,
		EntityType:      et,
		Metadata:        metadata,
	}, nil
}

func buildPageQuery(pm auditlogs.Page) (string, error) {
	var query []string
	var emq string

	if pm.ID != "" {
		query = append(query, "l.id = :id")
	}
	if pm.RequestID != "" {
		query = append(query, "l.request_id = :request_id")
	}
	if pm.ActorID != "" {
		query = append(query, "l.actor_id = :actor_id")
	}
	if pm.EntityType != "" {
		query = append(query, "l.entity_type = :entity_type")
	}
	if pm.EntityID != "" {
		query = append(query, "l.entity_id = :entity_id")
	}

	if len(query) > 0 {
		emq = fmt.Sprintf("WHERE %s", strings.Join(query, " AND "))
	}

	return emq, nil
}

type dbAuditLogsPage struct {
	Total      uint64 `db:"total"`
	Limit      uint64 `db:"limit"`
	Offset     uint64 `db:"offset"`
	Order      string `db:"order"`
	Dir        string `db:"dir"`
	ID         string `db:"id"`
	RequestID  string `db:"request_id"`
	OccuredAt  string `db:"occurred_at"`
	ActorID    string `db:"actor_id"`
	EntityType string `db:"entity_type"`
	EntityID   string `db:"entity_id"`
}

func toDBAuditLogsPage(pm auditlogs.Page) (dbAuditLogsPage, error) {
	return dbAuditLogsPage{
		Total:      pm.Total,
		Limit:      pm.Limit,
		Offset:     pm.Offset,
		Order:      pm.Order,
		Dir:        pm.Dir,
		ID:         pm.ID,
		RequestID:  pm.RequestID,
		OccuredAt:  pm.OccuredAt,
		ActorID:    pm.ActorID,
		EntityType: pm.EntityType,
		EntityID:   pm.EntityID,
	}, nil
}

func (ar *auditLogsRepo) processRows(rows *sqlx.Rows) ([]auditlogs.AuditLog, error) {
	var items []auditlogs.AuditLog
	for rows.Next() {
		dbal := dbAuditLog{}
		if err := rows.StructScan(&dbal); err != nil {
			return items, err
		}
		al, err := toAuditLog(dbal)
		if err != nil {
			return items, err
		}
		items = append(items, al)
	}
	return items, nil
}
