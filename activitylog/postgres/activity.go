// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/absmach/magistrala/activitylog"
	"github.com/absmach/magistrala/internal/postgres"
	"github.com/absmach/magistrala/pkg/errors"
	repoerr "github.com/absmach/magistrala/pkg/errors/repository"
)

type repository struct {
	db postgres.Database
}

func NewRepository(db postgres.Database) activitylog.Repository {
	return &repository{db: db}
}

func (repo *repository) Save(ctx context.Context, activity activitylog.Activity) (err error) {
	q := `INSERT INTO activities (operation, occurred_at, payload)
		VALUES (:operation, :occurred_at, :payload)`

	dbActivity, err := toDBActivity(activity)
	if err != nil {
		return errors.Wrap(repoerr.ErrCreateEntity, err)
	}

	if _, err = repo.db.NamedExecContext(ctx, q, dbActivity); err != nil {
		return postgres.HandleError(repoerr.ErrCreateEntity, err)
	}

	return nil
}

func (repo *repository) RetrieveAll(ctx context.Context, page activitylog.Page) (activitylog.ActivitiesPage, error) {
	query := pageQuery(page)

	sq := "operation, occurred_at"
	if page.WithPayload {
		sq += ", payload"
	}
	q := fmt.Sprintf("SELECT %s FROM activities %s ORDER BY occurred_at %s LIMIT :limit OFFSET :offset;", sq, query, page.Direction)

	rows, err := repo.db.NamedQueryContext(ctx, q, page)
	if err != nil {
		return activitylog.ActivitiesPage{}, postgres.HandleError(repoerr.ErrViewEntity, err)
	}
	defer rows.Close()

	var items []activitylog.Activity
	for rows.Next() {
		var item dbActivity
		if err = rows.StructScan(&item); err != nil {
			return activitylog.ActivitiesPage{}, postgres.HandleError(repoerr.ErrViewEntity, err)
		}
		activity, err := toActivity(item)
		if err != nil {
			return activitylog.ActivitiesPage{}, err
		}
		items = append(items, activity)
	}

	tq := fmt.Sprintf(`SELECT COUNT(*) FROM activities %s`, query)

	total, err := postgres.Total(ctx, repo.db, tq, page)
	if err != nil {
		return activitylog.ActivitiesPage{}, postgres.HandleError(repoerr.ErrViewEntity, err)
	}

	activitiesPage := activitylog.ActivitiesPage{
		Total:      total,
		Offset:     page.Offset,
		Limit:      page.Limit,
		Activities: items,
	}

	return activitiesPage, nil
}

func pageQuery(pm activitylog.Page) string {
	var query []string
	var emq string
	if pm.Operation != "" {
		query = append(query, "operation = :operation")
	}
	if !pm.From.IsZero() {
		query = append(query, "occurred_at >= :from")
	}
	if !pm.To.IsZero() {
		query = append(query, "occurred_at <= :to")
	}

	if len(query) > 0 {
		emq = fmt.Sprintf("WHERE %s", strings.Join(query, " AND "))
	}

	return emq
}

type dbActivity struct {
	Operation  string    `db:"operation"`
	OccurredAt time.Time `db:"occurred_at"`
	Payload    []byte    `db:"payload"`
}

func toDBActivity(activity activitylog.Activity) (dbActivity, error) {
	data := []byte("{}")
	if len(activity.Payload) > 0 {
		b, err := json.Marshal(activity.Payload)
		if err != nil {
			return dbActivity{}, errors.Wrap(repoerr.ErrMalformedEntity, err)
		}
		data = b
	}

	return dbActivity{
		Operation:  activity.Operation,
		OccurredAt: activity.OccurredAt,
		Payload:    data,
	}, nil
}

func toActivity(activity dbActivity) (activitylog.Activity, error) {
	var payload map[string]interface{}
	if activity.Payload != nil {
		if err := json.Unmarshal(activity.Payload, &payload); err != nil {
			return activitylog.Activity{}, errors.Wrap(repoerr.ErrMalformedEntity, err)
		}
	}

	return activitylog.Activity{
		Operation:  activity.Operation,
		OccurredAt: activity.OccurredAt,
		Payload:    payload,
	}, nil
}
