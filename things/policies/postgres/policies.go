package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgtype"
	"github.com/mainflux/mainflux/internal/postgres"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/things/policies"
)

var _ policies.Repository = (*prepo)(nil)

var (
	// ErrInvalidEntityType indicates that the entity type is invalid.
	ErrInvalidEntityType = errors.New("invalid entity type")
)

type prepo struct {
	db postgres.Database
}

// NewRepository instantiates a PostgreSQL implementation of policy repository.
func NewRepository(db postgres.Database) policies.Repository {
	return &prepo{
		db: db,
	}
}

func (pr prepo) Save(ctx context.Context, policy policies.Policy) (policies.Policy, error) {
	q := `INSERT INTO policies (owner_id, subject, object, actions, created_at, updated_at)
		VALUES (:owner_id, :subject, :object, :actions, :created_at, :updated_at)
		RETURNING owner_id, subject, object, actions, created_at, updated_at;`

	dbp, err := toDBPolicy(policy)
	if err != nil {
		return policies.Policy{}, errors.Wrap(errors.ErrCreateEntity, err)
	}

	row, err := pr.db.NamedQueryContext(ctx, q, dbp)
	if err != nil {
		return policies.Policy{}, postgres.HandleError(err, errors.ErrCreateEntity)
	}

	defer row.Close()
	row.Next()
	dbp = dbPolicy{}
	if err := row.StructScan(&dbp); err != nil {
		return policies.Policy{}, err
	}

	return toPolicy(dbp)
}

func (pr prepo) Evaluate(ctx context.Context, entityType string, policy policies.Policy) error {
	q := ""
	switch entityType {
	case "client":
		// Evaluates if two clients are connected to the same group and the subject has the specified action
		// or subject is the owner of the object
		q = fmt.Sprintf(`(SELECT subject FROM policies WHERE (subject = :subject AND object IN (
				SELECT object FROM policies WHERE subject = '%s') AND '%s'=ANY(actions))) UNION
				(SELECT id as subject FROM clients WHERE owner = :subject AND id = '%s') LIMIT 1`, policy.Object, policy.Actions[0], policy.Object)
	case "group":
		// Evaluates if client is connected to the specified group and has the required action
		q = fmt.Sprintf(`(SELECT subject FROM policies WHERE subject = :subject AND object = :object AND '%s'=ANY(actions))
				UNION (SELECT id as subject FROM groups WHERE owner_id = :subject AND id = :object) LIMIT 1`, policy.Actions[0])
	default:
		return ErrInvalidEntityType
	}

	dbu, err := toDBPolicy(policy)
	if err != nil {
		return errors.Wrap(errors.ErrAuthorization, err)
	}
	row, err := pr.db.NamedQueryContext(ctx, q, dbu)
	if err != nil {
		return postgres.HandleError(err, errors.ErrAuthorization)
	}

	defer row.Close()

	if ok := row.Next(); !ok {
		return errors.Wrap(errors.ErrAuthorization, row.Err())
	}
	var rPolicy dbPolicy
	if err := row.StructScan(&rPolicy); err != nil {
		return err
	}
	return nil
}

func (pr prepo) Update(ctx context.Context, policy policies.Policy) (policies.Policy, error) {
	q := `UPDATE policies SET actions = :actions, updated_at = :updated_at
		WHERE subject = :subject AND object = :object`

	dbp, err := toDBPolicy(policy)
	if err != nil {
		return policies.Policy{}, errors.Wrap(errors.ErrUpdateEntity, err)
	}

	row, err := pr.db.NamedQueryContext(ctx, q, dbp)
	if err != nil {
		return policies.Policy{}, postgres.HandleError(err, errors.ErrUpdateEntity)
	}

	defer row.Close()
	if ok := row.Next(); !ok {
		return policies.Policy{}, errors.Wrap(errors.ErrNotFound, row.Err())
	}
	dbp = dbPolicy{}
	if err := row.StructScan(&dbp); err != nil {
		return policies.Policy{}, errors.Wrap(err, errors.ErrUpdateEntity)
	}
	return toPolicy(dbp)
}

func (pr prepo) Retrieve(ctx context.Context, pm policies.Page) (policies.PolicyPage, error) {
	var query []string
	var emq string

	if pm.OwnerID != "" {
		query = append(query, fmt.Sprintf("owner_id = '%s'", pm.OwnerID))
	}
	if pm.Subject != "" {
		query = append(query, fmt.Sprintf("subject = '%s'", pm.Subject))
	}
	if pm.Object != "" {
		query = append(query, fmt.Sprintf("object = '%s'", pm.Object))
	}
	if pm.Action != "" {
		query = append(query, fmt.Sprintf("'%s' = ANY (actions)", pm.Action))
	}

	if len(query) > 0 {
		emq = fmt.Sprintf(" WHERE %s", strings.Join(query, " AND "))
	}

	q := fmt.Sprintf(`SELECT owner_id, subject, object, actions
		FROM policies %s ORDER BY updated_at LIMIT :limit OFFSET :offset;`, emq)
	params := map[string]interface{}{
		"limit":  pm.Limit,
		"offset": pm.Offset,
	}
	rows, err := pr.db.NamedQueryContext(ctx, q, params)
	if err != nil {
		return policies.PolicyPage{}, errors.Wrap(errors.ErrViewEntity, err)
	}
	defer rows.Close()

	var items []policies.Policy
	for rows.Next() {
		dbp := dbPolicy{}
		if err := rows.StructScan(&dbp); err != nil {
			return policies.PolicyPage{}, errors.Wrap(errors.ErrViewEntity, err)
		}

		policy, err := toPolicy(dbp)
		if err != nil {
			return policies.PolicyPage{}, err
		}

		items = append(items, policy)
	}

	cq := fmt.Sprintf(`SELECT COUNT(*) FROM policies %s;`, emq)

	total, err := postgres.Total(ctx, pr.db, cq, params)
	if err != nil {
		return policies.PolicyPage{}, errors.Wrap(errors.ErrViewEntity, err)
	}

	page := policies.PolicyPage{
		Policies: items,
		Page: policies.Page{
			Total:  total,
			Offset: pm.Offset,
			Limit:  pm.Limit,
		},
	}

	return page, nil
}

func (pr prepo) Delete(ctx context.Context, p policies.Policy) error {
	dbp := dbPolicy{
		Subject: p.Subject,
		Object:  p.Object,
	}
	q := `DELETE FROM policies WHERE subject = :subject AND object = :object`
	if _, err := pr.db.NamedExecContext(ctx, q, dbp); err != nil {
		return errors.Wrap(errors.ErrRemoveEntity, err)
	}
	return nil
}

type dbPolicy struct {
	OwnerID   string           `db:"owner_id"`
	Subject   string           `db:"subject"`
	Object    string           `db:"object"`
	Actions   pgtype.TextArray `db:"actions"`
	CreatedAt time.Time        `db:"created_at"`
	UpdatedAt time.Time        `db:"updated_at"`
}

func toDBPolicy(p policies.Policy) (dbPolicy, error) {
	var ps pgtype.TextArray
	if err := ps.Set(p.Actions); err != nil {
		return dbPolicy{}, err
	}

	return dbPolicy{
		OwnerID:   p.OwnerID,
		Subject:   p.Subject,
		Object:    p.Object,
		Actions:   ps,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}, nil
}

func toPolicy(dbp dbPolicy) (policies.Policy, error) {
	var ps []string
	for _, e := range dbp.Actions.Elements {
		ps = append(ps, e.String)
	}

	return policies.Policy{
		OwnerID:   dbp.OwnerID,
		Subject:   dbp.Subject,
		Object:    dbp.Object,
		Actions:   ps,
		CreatedAt: dbp.CreatedAt,
		UpdatedAt: dbp.UpdatedAt,
	}, nil
}
