package postgres

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/mainflux/mainflux/errors"
	"github.com/mainflux/mainflux/users"
	"github.com/opentracing/opentracing-go"
)

var _ Database = (*database)(nil)

type database struct {
	db *sqlx.DB
}

// Database provides a database interface
type Database interface {
	NamedExecContext(context.Context, string, interface{}) (sql.Result, errors.Error)
	QueryRowxContext(context.Context, string, ...interface{}) *sqlx.Row
	NamedQueryContext(context.Context, string, interface{}) (*sqlx.Rows, errors.Error)
	GetContext(context.Context, interface{}, string, ...interface{}) errors.Error
}

// NewDatabase creates a ThingDatabase instance
func NewDatabase(db *sqlx.DB) Database {
	return &database{
		db: db,
	}
}

func (dm database) NamedExecContext(ctx context.Context, query string, args interface{}) (sql.Result, errors.Error) {
	addSpanTags(ctx, query)
	println(query)
	result, err := dm.db.NamedExecContext(ctx, query, args)
	if pqErr, ok := err.(*pq.Error); ok && errDuplicate == pqErr.Code.Name() {
		return result, errors.Wrap(users.ErrConflict, errors.Cast(err))
	}
	return result, errors.Cast(err)
}

func (dm database) QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	addSpanTags(ctx, query)
	println(query)
	return dm.db.QueryRowxContext(ctx, query, args...)
}

func (dm database) NamedQueryContext(ctx context.Context, query string, args interface{}) (*sqlx.Rows, errors.Error) {
	addSpanTags(ctx, query)
	println(query)
	result, err := dm.db.NamedQueryContext(ctx, query, args)
	return result, errors.Cast(err)
}

func (dm database) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) errors.Error {
	addSpanTags(ctx, query)
	println(query)
	return errors.Cast((dm.db.GetContext(ctx, dest, query, args...)))
}

func addSpanTags(ctx context.Context, query string) {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		span.SetTag("sql.statement", query)
		span.SetTag("span.kind", "client")
		span.SetTag("peer.service", "postgres")
		span.SetTag("db.type", "sql")
	}
}
