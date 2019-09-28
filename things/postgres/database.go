package postgres

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/opentracing/opentracing-go"
)

var _ DatabaseMiddleware = (*databaseMiddleware)(nil)

type databaseMiddleware struct {
	db *sqlx.DB
}

// DatabaseMiddleware provides a database interface
type DatabaseMiddleware interface {
	NamedExec(context.Context, string, interface{}) (sql.Result, error)
	QueryRowx(context.Context, string, ...interface{}) *sqlx.Row
	NamedQuery(context.Context, string, interface{}) (*sqlx.Rows, error)
	Get(context.Context, interface{}, string, ...interface{}) error
}

// NewDatabaseMiddleware creates a ThingDatabase instance
func NewDatabaseMiddleware(db *sqlx.DB) DatabaseMiddleware {
	return &databaseMiddleware{
		db: db,
	}
}

func (dm databaseMiddleware) NamedExec(ctx context.Context, query string, args interface{}) (sql.Result, error) {
	addSpanTags(ctx, query)
	return dm.db.NamedExec(query, args)
}

func (dm databaseMiddleware) QueryRowx(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	addSpanTags(ctx, query)
	return dm.db.QueryRowx(query, args...)
}

func (dm databaseMiddleware) NamedQuery(ctx context.Context, query string, args interface{}) (*sqlx.Rows, error) {
	addSpanTags(ctx, query)
	return dm.db.NamedQuery(query, args)
}

func (dm databaseMiddleware) Get(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	addSpanTags(ctx, query)
	return dm.db.Get(dest, query, args...)
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
