package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mainflux/mainflux/internal/postgres"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/users/groups"
)

var _ groups.GroupRepository = (*groupRepository)(nil)

type groupRepository struct {
	db postgres.Database
}

// NewGroupRepo instantiates a PostgreSQL implementation of group
// repository.
func NewGroupRepo(db postgres.Database) groups.GroupRepository {
	return &groupRepository{
		db: db,
	}
}

// TODO - check parent group write access.
func (repo groupRepository) Save(ctx context.Context, g groups.Group) (groups.Group, error) {
	q := `INSERT INTO groups (name, description, id, owner_id, metadata, created_at, updated_at, status)
		VALUES (:name, :description, :id, :owner_id, :metadata, :created_at, :updated_at, :status)
		RETURNING id, name, description, owner_id, COALESCE(parent_id, '') AS parent_id, metadata, created_at, updated_at, status;`
	if g.ParentID != "" {
		q = `INSERT INTO groups (name, description, id, owner_id, parent_id, metadata, created_at, updated_at, status)
		VALUES (:name, :description, :id, :owner_id, :parent_id, :metadata, :created_at, :updated_at, :status)
		RETURNING id, name, description, owner_id, COALESCE(parent_id, '') AS parent_id, metadata, created_at, updated_at, status;`
	}
	dbg, err := toDBGroup(g)
	if err != nil {
		return groups.Group{}, err
	}
	row, err := repo.db.NamedQueryContext(ctx, q, dbg)
	if err != nil {
		return groups.Group{}, postgres.HandleError(err, errors.ErrCreateEntity)
	}

	defer row.Close()
	row.Next()
	dbg = dbGroup{}
	if err := row.StructScan(&dbg); err != nil {
		return groups.Group{}, err
	}

	return toGroup(dbg)
}

func (repo groupRepository) RetrieveByID(ctx context.Context, id string) (groups.Group, error) {
	dbu := dbGroup{
		ID: id,
	}
	q := `SELECT id, name, owner_id, COALESCE(parent_id, '') AS parent_id, description, metadata, created_at, updated_at, status FROM groups
	    WHERE id = $1`
	if err := repo.db.QueryRowxContext(ctx, q, dbu.ID).StructScan(&dbu); err != nil {
		if err == sql.ErrNoRows {
			return groups.Group{}, errors.Wrap(errors.ErrNotFound, err)

		}
		return groups.Group{}, errors.Wrap(errors.ErrViewEntity, err)
	}
	return toGroup(dbu)
}

func (repo groupRepository) RetrieveAll(ctx context.Context, gm groups.GroupsPage) (groups.GroupsPage, error) {
	var q string
	query, err := buildQuery(gm)
	if err != nil {
		return groups.GroupsPage{}, err
	}

	if gm.ID != "" {
		q = buildHierachy(gm)
	}
	if gm.ID == "" {
		q = `SELECT DISTINCT g.id, g.owner_id, COALESCE(g.parent_id, '') AS parent_id, g.name, g.description,
		g.metadata, g.created_at, g.updated_at, g.status FROM groups g`
	}
	q = fmt.Sprintf("%s %s ORDER BY g.updated_at LIMIT :limit OFFSET :offset;", q, query)

	dbPage, err := toDBGroupPage(gm)
	if err != nil {
		return groups.GroupsPage{}, errors.Wrap(postgres.ErrFailedToRetrieveAll, err)
	}
	rows, err := repo.db.NamedQueryContext(ctx, q, dbPage)
	if err != nil {
		return groups.GroupsPage{}, errors.Wrap(postgres.ErrFailedToRetrieveAll, err)
	}
	defer rows.Close()

	items, err := repo.processRows(rows)
	if err != nil {
		return groups.GroupsPage{}, errors.Wrap(postgres.ErrFailedToRetrieveAll, err)
	}

	cq := "SELECT COUNT(*) FROM groups g"
	if query != "" {
		cq = fmt.Sprintf(" %s %s", cq, query)
	}

	total, err := postgres.Total(ctx, repo.db, cq, dbPage)
	if err != nil {
		return groups.GroupsPage{}, errors.Wrap(postgres.ErrFailedToRetrieveAll, err)
	}

	page := gm
	page.Groups = items
	page.Total = total

	return page, nil
}

func (repo groupRepository) Memberships(ctx context.Context, clientID string, gm groups.GroupsPage) (groups.MembershipsPage, error) {
	var q string
	query, err := buildQuery(gm)
	if err != nil {
		return groups.MembershipsPage{}, err
	}
	if gm.ID != "" {
		q = buildHierachy(gm)
	}
	if gm.ID == "" {
		q = `SELECT DISTINCT g.id, g.owner_id, COALESCE(g.parent_id, '') AS parent_id, g.name, g.description,
		g.metadata, g.created_at, g.updated_at, g.status FROM groups g`
	}
	q = fmt.Sprintf(`%s INNER JOIN policies ON g.id=policies.object %s AND policies.subject = :client_id
			AND policies.object IN (SELECT object FROM policies WHERE subject = '%s' AND '%s'=ANY(actions))
			ORDER BY g.updated_at LIMIT :limit OFFSET :offset;`, q, query, gm.Subject, gm.Action)

	dbPage, err := toDBGroupPage(gm)
	if err != nil {
		return groups.MembershipsPage{}, errors.Wrap(postgres.ErrFailedToRetrieveMembership, err)
	}
	dbPage.ClientID = clientID
	rows, err := repo.db.NamedQueryContext(ctx, q, dbPage)
	if err != nil {
		return groups.MembershipsPage{}, errors.Wrap(postgres.ErrFailedToRetrieveMembership, err)
	}
	defer rows.Close()

	var items []groups.Group
	for rows.Next() {
		dbg := dbGroup{}
		if err := rows.StructScan(&dbg); err != nil {
			return groups.MembershipsPage{}, errors.Wrap(postgres.ErrFailedToRetrieveMembership, err)
		}
		group, err := toGroup(dbg)
		if err != nil {
			return groups.MembershipsPage{}, errors.Wrap(postgres.ErrFailedToRetrieveMembership, err)
		}
		items = append(items, group)
	}

	cq := fmt.Sprintf(`SELECT COUNT(*) FROM groups g INNER JOIN policies
        ON g.id=policies.object %s AND policies.subject = :client_id`, query)

	total, err := postgres.Total(ctx, repo.db, cq, dbPage)
	if err != nil {
		return groups.MembershipsPage{}, errors.Wrap(postgres.ErrFailedToRetrieveMembership, err)
	}
	page := groups.MembershipsPage{
		Memberships: items,
		Page: groups.Page{
			Total: total,
		},
	}

	return page, nil
}

func (repo groupRepository) Update(ctx context.Context, g groups.Group) (groups.Group, error) {
	var query []string
	var upq string
	if g.Name != "" {
		query = append(query, "name = :name,")
	}
	if g.Description != "" {
		query = append(query, "description = :description,")
	}
	if g.Metadata != nil {
		query = append(query, "metadata = :metadata,")
	}
	if len(query) > 0 {
		upq = strings.Join(query, " ")
	}
	q := fmt.Sprintf(`UPDATE groups SET %s updated_at = :updated_at
		WHERE id = :id AND status = %d
		RETURNING id, name, description, owner_id, COALESCE(parent_id, '') AS parent_id, metadata, created_at, updated_at, status`, upq, groups.EnabledStatus)

	dbu, err := toDBGroup(g)
	if err != nil {
		return groups.Group{}, errors.Wrap(errors.ErrUpdateEntity, err)
	}

	row, err := repo.db.NamedQueryContext(ctx, q, dbu)
	if err != nil {
		return groups.Group{}, postgres.HandleError(err, errors.ErrUpdateEntity)
	}

	defer row.Close()
	if ok := row.Next(); !ok {
		return groups.Group{}, errors.Wrap(errors.ErrNotFound, row.Err())
	}
	dbu = dbGroup{}
	if err := row.StructScan(&dbu); err != nil {
		return groups.Group{}, errors.Wrap(err, errors.ErrUpdateEntity)
	}
	return toGroup(dbu)
}

func (repo groupRepository) ChangeStatus(ctx context.Context, id string, status groups.Status) (groups.Group, error) {
	qc := fmt.Sprintf(`UPDATE groups SET status = %d WHERE id = :id RETURNING id, name, description, owner_id, COALESCE(parent_id, '') AS parent_id, metadata, created_at, updated_at, status`, status)

	dbg := dbGroup{
		ID: id,
	}
	row, err := repo.db.NamedQueryContext(ctx, qc, dbg)
	if err != nil {
		return groups.Group{}, postgres.HandleError(err, errors.ErrUpdateEntity)
	}

	defer row.Close()
	if ok := row.Next(); !ok {
		return groups.Group{}, errors.Wrap(errors.ErrNotFound, row.Err())
	}
	dbg = dbGroup{}
	if err := row.StructScan(&dbg); err != nil {
		return groups.Group{}, errors.Wrap(err, errors.ErrUpdateEntity)
	}

	return toGroup(dbg)
}

func buildHierachy(gm groups.GroupsPage) string {
	query := ""
	switch {
	case gm.Direction >= 0: // ancestors
		query = `WITH RECURSIVE groups_cte as (
			SELECT id, COALESCE(parent_id, '') AS parent_id, owner_id, name, description, metadata, created_at, updated_at, status, 1 as level from groups WHERE id = :id
			UNION SELECT x.id, COALESCE(x.parent_id, '') AS parent_id, x.owner_id, x.name, x.description, x.metadata, x.created_at, x.updated_at, x.status, level - 1 from groups x
			INNER JOIN groups_cte a ON a.parent_id = x.id
		) SELECT * FROM groups_cte g`

	case gm.Direction < 0: // descendants
		query = `WITH RECURSIVE groups_cte as (
			SELECT id, COALESCE(parent_id, '') AS parent_id, owner_id, name, description, metadata, created_at, updated_at, status, 1 as level, CONCAT('', '', id) as path from groups WHERE id = :id
			UNION SELECT x.id, COALESCE(x.parent_id, '') AS parent_id, x.owner_id, x.name, x.description, x.metadata, x.created_at, x.updated_at, x.status, level + 1, CONCAT(path, '.', x.id) as path from groups x
			INNER JOIN groups_cte d ON d.id = x.parent_id
		) SELECT * FROM groups_cte g`
	}
	return query
}
func buildQuery(gm groups.GroupsPage) (string, error) {
	queries := []string{}

	if gm.Name != "" {
		queries = append(queries, "g.name = :name")
	}
	if gm.Status != groups.AllStatus {
		queries = append(queries, fmt.Sprintf("g.status = %d", gm.Status))
	}

	if gm.Subject != "" {
		queries = append(queries, fmt.Sprintf("(g.owner_id = '%s' OR id IN (SELECT object as id FROM policies WHERE subject = '%s' AND '%s'=ANY(actions)))", gm.OwnerID, gm.Subject, gm.Action))
	}
	if len(gm.Metadata) > 0 {
		queries = append(queries, "'g.metadata @> :metadata'")
	}
	if len(queries) > 0 {
		return fmt.Sprintf("WHERE %s", strings.Join(queries, " AND ")), nil
	}
	return "", nil
}

type dbGroup struct {
	ID          string        `db:"id"`
	ParentID    string        `db:"parent_id"`
	OwnerID     string        `db:"owner_id"`
	Name        string        `db:"name"`
	Description string        `db:"description"`
	Level       int           `db:"level"`
	Path        string        `db:"path,omitempty"`
	Metadata    []byte        `db:"metadata"`
	CreatedAt   time.Time     `db:"created_at"`
	UpdatedAt   time.Time     `db:"updated_at"`
	Status      groups.Status `db:"status"`
}

func toDBGroup(g groups.Group) (dbGroup, error) {
	data := []byte("{}")
	if len(g.Metadata) > 0 {
		b, err := json.Marshal(g.Metadata)
		if err != nil {
			return dbGroup{}, errors.Wrap(errors.ErrMalformedEntity, err)
		}
		data = b
	}
	return dbGroup{
		ID:          g.ID,
		Name:        g.Name,
		ParentID:    g.ParentID,
		OwnerID:     g.OwnerID,
		Description: g.Description,
		Metadata:    data,
		Path:        g.Path,
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
		Status:      g.Status,
	}, nil
}

func toGroup(g dbGroup) (groups.Group, error) {
	var metadata groups.Metadata
	if g.Metadata != nil {
		if err := json.Unmarshal([]byte(g.Metadata), &metadata); err != nil {
			return groups.Group{}, errors.Wrap(errors.ErrMalformedEntity, err)
		}
	}
	return groups.Group{
		ID:          g.ID,
		Name:        g.Name,
		ParentID:    g.ParentID,
		OwnerID:     g.OwnerID,
		Description: g.Description,
		Metadata:    metadata,
		Level:       g.Level,
		Path:        g.Path,
		UpdatedAt:   g.UpdatedAt,
		CreatedAt:   g.CreatedAt,
		Status:      g.Status,
	}, nil
}

func (gr groupRepository) processRows(rows *sqlx.Rows) ([]groups.Group, error) {
	var items []groups.Group
	for rows.Next() {
		dbg := dbGroup{}
		if err := rows.StructScan(&dbg); err != nil {
			return items, err
		}
		group, err := toGroup(dbg)
		if err != nil {
			return items, err
		}
		items = append(items, group)
	}
	return items, nil
}

func toDBGroupPage(pm groups.GroupsPage) (dbGroupPage, error) {
	level := groups.MaxLevel
	if pm.Level < groups.MaxLevel {
		level = pm.Level
	}
	data := []byte("{}")
	if len(pm.Metadata) > 0 {
		b, err := json.Marshal(pm.Metadata)
		if err != nil {
			return dbGroupPage{}, errors.Wrap(errors.ErrMalformedEntity, err)
		}
		data = b
	}
	return dbGroupPage{
		ID:       pm.ID,
		Name:     pm.Name,
		Metadata: data,
		Path:     pm.Path,
		Level:    level,
		Total:    pm.Total,
		Offset:   pm.Offset,
		Limit:    pm.Limit,
		ParentID: pm.ID,
		OwnerID:  pm.OwnerID,
	}, nil
}

type dbGroupPage struct {
	ClientID string `db:"client_id"`
	ID       string `db:"id"`
	Name     string `db:"name"`
	ParentID string `db:"parent_id"`
	OwnerID  string `db:"owner_id"`
	Metadata []byte `db:"metadata"`
	Path     string `db:"path"`
	Level    uint64 `db:"level"`
	Total    uint64 `db:"total"`
	Limit    uint64 `db:"limit"`
	Offset   uint64 `db:"offset"`
}
