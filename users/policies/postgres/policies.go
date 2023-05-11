package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgtype"
	"github.com/mainflux/mainflux/internal/postgres"
	"github.com/mainflux/mainflux/pkg/clients"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/users/policies"
)

var _ policies.PolicyRepository = (*policyRepository)(nil)

var (
	// ErrInvalidEntityType indicates that the entity type is invalid.
	ErrInvalidEntityType = errors.New("invalid entity type")
)

type policyRepository struct {
	db postgres.Database
}

// NewPolicyRepo instantiates a PostgreSQL implementclients.Serviceation of policy repository.
func NewPolicyRepo(db postgres.Database) policies.PolicyRepository {
	return &policyRepository{
		db: db,
	}
}

func (pr policyRepository) Save(ctx context.Context, policy policies.Policy) error {
	q := `INSERT INTO policies (owner_id, subject, object, actions, created_at)
		VALUES (:owner_id, :subject, :object, :actions, :created_at)`

	dbp, err := toDBPolicy(policy)
	if err != nil {
		return errors.Wrap(errors.ErrCreateEntity, err)
	}

	row, err := pr.db.NamedQueryContext(ctx, q, dbp)
	if err != nil {
		return postgres.HandleError(err, errors.ErrCreateEntity)
	}

	defer row.Close()

	return nil
}

func (pr policyRepository) CheckAdmin(ctx context.Context, id string) error {
	q := fmt.Sprintf(`SELECT id FROM clients WHERE id = '%s' AND role = '%d';`, id, clients.AdminRole)

	var clientID string
	if err := pr.db.QueryRowxContext(ctx, q).Scan(&clientID); err != nil {
		return errors.Wrap(errors.ErrAuthorization, err)
	}
	if clientID == "" {
		return errors.ErrAuthorization
	}

	return nil
}

func (pr policyRepository) CheckClientOwner(ctx context.Context, id, owner string) error {
	q := fmt.Sprintf(`SELECT id FROM clients WHERE id = '%s' AND owner = '%s';`, id, owner)

	var clientID string
	if err := pr.db.QueryRowxContext(ctx, q).Scan(&clientID); err != nil {
		return errors.Wrap(errors.ErrAuthorization, err)
	}
	if clientID == "" {
		return errors.ErrAuthorization
	}

	return nil
}

func (pr policyRepository) CheckGroupOwner(ctx context.Context, id, owner string) error {
	q := fmt.Sprintf(`SELECT id FROM group WHERE id = '%s' AND owner_id = '%s';`, id, owner)

	var clientID string
	if err := pr.db.QueryRowxContext(ctx, q).Scan(&clientID); err != nil {
		return errors.Wrap(errors.ErrAuthorization, err)
	}
	if clientID == "" {
		return errors.ErrAuthorization
	}

	return nil
}

func (pr policyRepository) Evaluate(ctx context.Context, entityType string, policy policies.Policy) error {
	q := ""
	switch entityType {
	case "client":
		// Evaluates if two clients are connected to the same group and the subject has the specified action
		// or subject is the owner of the object
		q = fmt.Sprintf(`SELECT COALESCE(p.subject, c.id) as subject FROM policies p
		JOIN policies p2 ON p.object = p2.object LEFT JOIN clients c ON c.owner_id = :subject AND c.id = :object
		WHERE (p.subject = :subject AND p2.subject = :object AND '%s' = ANY(p.actions)) OR (c.id IS NOT NULL) LIMIT 1;`,
			policy.Actions[0])
	case "group":
		// Evaluates if client is connected to the specified group and has the required action
		q = fmt.Sprintf(`SELECT DISTINCT policies.subject FROM policies
		LEFT JOIN groups ON groups.owner_id = policies.subject AND groups.id = policies.object
		WHERE policies.subject = :subject AND policies.object = :object AND '%s' = ANY(policies.actions)
		LIMIT 1`, policy.Actions[0])
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

func (pr policyRepository) Update(ctx context.Context, policy policies.Policy) error {
	if err := policy.Validate(); err != nil {
		return errors.Wrap(errors.ErrCreateEntity, err)
	}
	q := `UPDATE policies SET actions = :actions, updated_at = :updated_at, updated_by = :updated_by
		WHERE subject = :subject AND object = :object`

	dbu, err := toDBPolicy(policy)
	if err != nil {
		return errors.Wrap(errors.ErrUpdateEntity, err)
	}

	if _, err := pr.db.NamedExecContext(ctx, q, dbu); err != nil {
		return errors.Wrap(errors.ErrUpdateEntity, err)
	}

	return nil
}

func (pr policyRepository) Retrieve(ctx context.Context, pm policies.Page) (policies.PolicyPage, error) {
	var query []string
	var emq string

	if pm.OwnerID != "" {
		query = append(query, "owner_id = :owner_id")
	}
	if pm.Subject != "" {
		query = append(query, "subject = :subject")
	}
	if pm.Object != "" {
		query = append(query, "object = :object")
	}
	if pm.Action != "" {
		query = append(query, ":action = ANY (actions)")
	}

	if len(query) > 0 {
		emq = fmt.Sprintf(" WHERE %s", strings.Join(query, " AND "))
	}

	q := fmt.Sprintf(`SELECT owner_id, subject, object, actions, created_at, updated_at, updated_by
		FROM policies %s ORDER BY updated_at LIMIT :limit OFFSET :offset;`, emq)

	dbPage, err := toDBPoliciesPage(pm)
	if err != nil {
		return policies.PolicyPage{}, errors.Wrap(errors.ErrViewEntity, err)
	}
	rows, err := pr.db.NamedQueryContext(ctx, q, dbPage)
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

	total, err := postgres.Total(ctx, pr.db, cq, dbPage)
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

func (pr policyRepository) Delete(ctx context.Context, p policies.Policy) error {
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
	UpdatedAt sql.NullTime     `db:"updated_at,omitempty"`
	UpdatedBy *string          `db:"updated_by,omitempty"`
}

func toDBPolicy(p policies.Policy) (dbPolicy, error) {
	var ps pgtype.TextArray
	if err := ps.Set(p.Actions); err != nil {
		return dbPolicy{}, err
	}
	var updatedAt sql.NullTime
	if !p.UpdatedAt.IsZero() {
		updatedAt = sql.NullTime{Time: p.UpdatedAt, Valid: true}
	}
	var updatedBy *string
	if p.UpdatedBy != "" {
		updatedBy = &p.UpdatedBy
	}
	return dbPolicy{
		OwnerID:   p.OwnerID,
		Subject:   p.Subject,
		Object:    p.Object,
		Actions:   ps,
		CreatedAt: p.CreatedAt,
		UpdatedAt: updatedAt,
		UpdatedBy: updatedBy,
	}, nil
}

func toPolicy(dbp dbPolicy) (policies.Policy, error) {
	var ps []string
	for _, e := range dbp.Actions.Elements {
		ps = append(ps, e.String)
	}
	var updatedAt time.Time
	if dbp.UpdatedAt.Valid {
		updatedAt = dbp.UpdatedAt.Time
	}
	var updatedBy string
	if dbp.UpdatedBy != nil {
		updatedBy = *dbp.UpdatedBy
	}
	return policies.Policy{
		OwnerID:   dbp.OwnerID,
		Subject:   dbp.Subject,
		Object:    dbp.Object,
		Actions:   ps,
		CreatedAt: dbp.CreatedAt,
		UpdatedAt: updatedAt,
		UpdatedBy: updatedBy,
	}, nil
}

func toDBPoliciesPage(pm policies.Page) (dbPoliciesPage, error) {
	return dbPoliciesPage{
		Total:   pm.Total,
		Offset:  pm.Offset,
		Limit:   pm.Limit,
		OwnerID: pm.OwnerID,
		Subject: pm.Subject,
		Object:  pm.Object,
		Action:  pm.Action,
	}, nil
}

type dbPoliciesPage struct {
	Total   uint64 `db:"total"`
	Limit   uint64 `db:"limit"`
	Offset  uint64 `db:"offset"`
	OwnerID string `db:"owner_id"`
	Subject string `db:"subject"`
	Object  string `db:"object"`
	Action  string `db:"action"`
}
