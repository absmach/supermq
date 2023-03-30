package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgtype" // required for SQL access
	mfclients "github.com/mainflux/mainflux/internal/mainflux/clients"
	"github.com/mainflux/mainflux/internal/postgres"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/things/clients"
)

var _ clients.Repository = (*clientRepo)(nil)

type clientRepo struct {
	db postgres.Database
}

// NewRepository instantiates a PostgreSQL
// implementation of Clients repository.
func NewRepository(db postgres.Database) clients.Repository {
	return &clientRepo{
		db: db,
	}
}

func (repo clientRepo) Save(ctx context.Context, cs ...clients.Client) ([]clients.Client, error) {
	tx, err := repo.db.BeginTxx(ctx, nil)
	if err != nil {
		return []clients.Client{}, errors.Wrap(errors.ErrCreateEntity, err)
	}

	for _, cli := range cs {
		q := `INSERT INTO clients (id, name, tags, owner_id, identity, secret, metadata, created_at, updated_at, updated_by, status)
        VALUES (:id, :name, :tags, :owner_id, :identity, :secret, :metadata, :created_at, :updated_at, :updated_by, :status)
        RETURNING id, name, tags, identity, secret, metadata, COALESCE(owner_id, '') AS owner_id, status, created_at, updated_at, updated_by`

		dbcli, err := toDBClient(cli)
		if err != nil {
			return []clients.Client{}, errors.Wrap(errors.ErrCreateEntity, err)
		}

		if _, err := tx.NamedExecContext(ctx, q, dbcli); err != nil {
			if err := tx.Rollback(); err != nil {
				return []clients.Client{}, postgres.HandleError(err, errors.ErrCreateEntity)
			}
		}
	}
	if err = tx.Commit(); err != nil {
		return []clients.Client{}, errors.Wrap(errors.ErrCreateEntity, err)
	}

	return cs, nil
}

func (repo clientRepo) RetrieveByID(ctx context.Context, id string) (clients.Client, error) {
	q := `SELECT id, name, tags, COALESCE(owner_id, '') AS owner_id, identity, secret, metadata, created_at, updated_at, updated_by, status 
        FROM clients
        WHERE id = $1`

	dbc := dbClient{
		ID: id,
	}

	if err := repo.db.QueryRowxContext(ctx, q, id).StructScan(&dbc); err != nil {
		if err == sql.ErrNoRows {
			return clients.Client{}, errors.Wrap(errors.ErrNotFound, err)

		}
		return clients.Client{}, errors.Wrap(errors.ErrViewEntity, err)
	}

	return toClient(dbc)
}

func (repo clientRepo) RetrieveBySecret(ctx context.Context, key string) (clients.Client, error) {
	q := fmt.Sprintf(`SELECT id, name, tags, COALESCE(owner_id, '') AS owner_id, identity, secret, metadata, created_at, updated_at, updated_by, status
        FROM clients
        WHERE secret = $1 AND status = %d`, mfclients.EnabledStatus)

	dbc := dbClient{
		Secret: key,
	}

	if err := repo.db.QueryRowxContext(ctx, q, key).StructScan(&dbc); err != nil {
		if err == sql.ErrNoRows {
			return clients.Client{}, errors.Wrap(errors.ErrNotFound, err)

		}
		return clients.Client{}, errors.Wrap(errors.ErrViewEntity, err)
	}

	return toClient(dbc)
}

func (repo clientRepo) RetrieveAll(ctx context.Context, pm clients.Page) (clients.ClientsPage, error) {
	query, err := pageQuery(pm)
	if err != nil {
		return clients.ClientsPage{}, errors.Wrap(errors.ErrViewEntity, err)
	}

	q := fmt.Sprintf(`SELECT c.id, c.name, c.tags, c.identity, c.secret, c.metadata, COALESCE(c.owner_id, '') AS owner_id, c.status, c.created_at
						FROM clients c %s ORDER BY c.created_at LIMIT :limit OFFSET :offset;`, query)

	dbPage, err := toDBClientsPage(pm)
	if err != nil {
		return clients.ClientsPage{}, errors.Wrap(postgres.ErrFailedToRetrieveAll, err)
	}
	rows, err := repo.db.NamedQueryContext(ctx, q, dbPage)
	if err != nil {
		return clients.ClientsPage{}, errors.Wrap(postgres.ErrFailedToRetrieveAll, err)
	}
	defer rows.Close()

	var items []clients.Client
	for rows.Next() {
		dbc := dbClient{}
		if err := rows.StructScan(&dbc); err != nil {
			return clients.ClientsPage{}, errors.Wrap(errors.ErrViewEntity, err)
		}

		c, err := toClient(dbc)
		if err != nil {
			return clients.ClientsPage{}, err
		}

		items = append(items, c)
	}
	cq := fmt.Sprintf(`SELECT COUNT(*) FROM clients c %s;`, query)

	total, err := postgres.Total(ctx, repo.db, cq, dbPage)
	if err != nil {
		return clients.ClientsPage{}, errors.Wrap(errors.ErrViewEntity, err)
	}

	page := clients.ClientsPage{
		Clients: items,
		Page: clients.Page{
			Total:  total,
			Offset: pm.Offset,
			Limit:  pm.Limit,
		},
	}

	return page, nil
}

func (repo clientRepo) Members(ctx context.Context, groupID string, pm clients.Page) (clients.MembersPage, error) {
	emq, err := pageQuery(pm)
	if err != nil {
		return clients.MembersPage{}, err
	}

	aq := ""
	// If not admin, the client needs to have a g_list action on the group
	if pm.Subject != "" {
		aq = fmt.Sprintf("AND EXISTS (SELECT 1 FROM policies WHERE policies.subject = '%s' AND policies.object = :group_id AND '%s'=ANY(actions))", pm.Subject, pm.Action)
	}

	q := fmt.Sprintf(`SELECT c.id, c.name, c.tags, c.metadata, c.identity, c.secret, c.status, c.created_at FROM clients c
		INNER JOIN policies ON c.id=policies.subject %s AND policies.object = :group_id %s ORDER BY c.created_at LIMIT :limit OFFSET :offset;`, emq, aq)
	dbPage, err := toDBClientsPage(pm)
	if err != nil {
		return clients.MembersPage{}, errors.Wrap(postgres.ErrFailedToRetrieveAll, err)
	}
	dbPage.GroupID = groupID
	rows, err := repo.db.NamedQueryContext(ctx, q, dbPage)
	if err != nil {
		return clients.MembersPage{}, errors.Wrap(postgres.ErrFailedToRetrieveMembers, err)
	}
	defer rows.Close()

	var items []clients.Client
	for rows.Next() {
		dbc := dbClient{}
		if err := rows.StructScan(&dbc); err != nil {
			return clients.MembersPage{}, errors.Wrap(postgres.ErrFailedToRetrieveMembers, err)
		}

		c, err := toClient(dbc)
		if err != nil {
			return clients.MembersPage{}, err
		}

		items = append(items, c)
	}
	cq := fmt.Sprintf(`SELECT COUNT(*) FROM clients c INNER JOIN policies ON c.id=policies.subject %s AND policies.object = :group_id;`, emq)

	total, err := postgres.Total(ctx, repo.db, cq, dbPage)
	if err != nil {
		return clients.MembersPage{}, errors.Wrap(postgres.ErrFailedToRetrieveMembers, err)
	}

	page := clients.MembersPage{
		Members: items,
		Page: clients.Page{
			Total:  total,
			Offset: pm.Offset,
			Limit:  pm.Limit,
		},
	}
	return page, nil
}

func (repo clientRepo) Update(ctx context.Context, client clients.Client) (clients.Client, error) {
	var query []string
	var upq string
	if client.Name != "" {
		query = append(query, "name = :name,")
	}
	if client.Metadata != nil {
		query = append(query, "metadata = :metadata,")
	}
	if len(query) > 0 {
		upq = strings.Join(query, " ")
	}
	client.Status = clients.EnabledStatus
	q := fmt.Sprintf(`UPDATE clients SET %s updated_at = :updated_at, updated_by = :updated_by
        WHERE id = :id AND status = :status
        RETURNING id, name, tags, identity, secret,  metadata, COALESCE(owner_id, '') AS owner_id, status, created_at, updated_at, updated_by`,
		upq)

	return repo.update(ctx, client, q)
}

func (repo clientRepo) UpdateTags(ctx context.Context, client clients.Client) (clients.Client, error) {
	client.Status = clients.EnabledStatus
	q := `UPDATE clients SET tags = :tags, updated_at = :updated_at, updated_by = :updated_by
        WHERE id = :id AND status = :status
        RETURNING id, name, tags, identity, secret, metadata, COALESCE(owner_id, '') AS owner_id, status, created_at, updated_at, updated_by`

	return repo.update(ctx, client, q)
}

func (repo clientRepo) UpdateIdentity(ctx context.Context, client clients.Client) (clients.Client, error) {
	client.Status = clients.EnabledStatus
	q := `UPDATE clients SET identity = :identity, updated_at = :updated_at, updated_by = :updated_by
        WHERE id = :id AND status = :status
        RETURNING id, name, tags, identity, secret, metadata, COALESCE(owner_id, '') AS owner_id, status, created_at, updated_at, updated_by`

	return repo.update(ctx, client, q)
}

func (repo clientRepo) UpdateSecret(ctx context.Context, client clients.Client) (clients.Client, error) {
	client.Status = clients.EnabledStatus
	q := `UPDATE clients SET secret = :secret, updated_at = :updated_at, updated_by = :updated_by
        WHERE id = :id AND status = :status
        RETURNING id, name, tags, identity, secret, metadata, COALESCE(owner_id, '') AS owner_id, status, created_at, updated_at, updated_by`

	return repo.update(ctx, client, q)
}

func (repo clientRepo) UpdateOwner(ctx context.Context, client clients.Client) (clients.Client, error) {
	client.Status = clients.EnabledStatus
	q := `UPDATE clients SET owner_id = :owner_id, updated_at = :updated_at, updated_by = :updated_by
        WHERE id = :id AND status = :status
        RETURNING id, name, tags, identity, secret, metadata, COALESCE(owner_id, '') AS owner_id, status, created_at, updated_at, updated_by`

	return repo.update(ctx, client, q)
}

func (repo clientRepo) ChangeStatus(ctx context.Context, client clients.Client) (clients.Client, error) {
	q := `UPDATE clients SET status = :status WHERE id = :id
        RETURNING id, name, tags, identity, secret, metadata, COALESCE(owner_id, '') AS owner_id, status, created_at, updated_at, updated_by`

	return repo.update(ctx, client, q)
}

// generic update function
func (repo clientRepo) update(ctx context.Context, client clients.Client, query string) (clients.Client, error) {
	dbc, err := toDBClient(client)
	if err != nil {
		return clients.Client{}, errors.Wrap(errors.ErrUpdateEntity, err)
	}

	row, err := repo.db.NamedQueryContext(ctx, query, dbc)
	if err != nil {
		return clients.Client{}, postgres.HandleError(err, errors.ErrUpdateEntity)
	}

	defer row.Close()
	if ok := row.Next(); !ok {
		return clients.Client{}, errors.Wrap(errors.ErrNotFound, row.Err())
	}
	dbc = dbClient{}
	if err := row.StructScan(&dbc); err != nil {
		return clients.Client{}, err
	}

	return toClient(dbc)
}

type dbClient struct {
	ID        string           `db:"id"`
	Name      string           `db:"name,omitempty"`
	Tags      pgtype.TextArray `db:"tags,omitempty"`
	Identity  string           `db:"identity"`
	Owner     string           `db:"owner_id,omitempty"` // nullable
	Secret    string           `db:"secret"`
	Metadata  []byte           `db:"metadata,omitempty"`
	CreatedAt time.Time        `db:"created_at"`
	UpdatedAt sql.NullTime     `db:"updated_at,omitempty"`
	UpdatedBy *string          `db:"updated_by,omitempty"`
	Groups    []groups.Group   `db:"groups"`
	Status    mfclients.Status `db:"status"`
}

func toDBClient(c clients.Client) (dbClient, error) {
	data := []byte("{}")
	if len(c.Metadata) > 0 {
		b, err := json.Marshal(c.Metadata)
		if err != nil {
			return dbClient{}, errors.Wrap(errors.ErrMalformedEntity, err)
		}
		data = b
	}
	var tags pgtype.TextArray
	if err := tags.Set(c.Tags); err != nil {
		return dbClient{}, err
	}
	var updatedBy *string
	if c.UpdatedBy != "" {
		updatedBy = &c.UpdatedBy
	}
	var updatedAt sql.NullTime
	if c.UpdatedAt != (time.Time{}) {
		updatedAt = sql.NullTime{Time: c.UpdatedAt, Valid: true}
	}

	return dbClient{
		ID:        c.ID,
		Name:      c.Name,
		Tags:      tags,
		Owner:     c.Owner,
		Identity:  c.Credentials.Identity,
		Secret:    c.Credentials.Secret,
		Metadata:  data,
		CreatedAt: c.CreatedAt,
		UpdatedAt: updatedAt,
		UpdatedBy: updatedBy,
		Status:    c.Status,
	}, nil
}

func toClient(c dbClient) (clients.Client, error) {
	var metadata clients.Metadata
	if c.Metadata != nil {
		if err := json.Unmarshal([]byte(c.Metadata), &metadata); err != nil {
			return clients.Client{}, errors.Wrap(errors.ErrMalformedEntity, err)
		}
	}
	var tags []string
	for _, e := range c.Tags.Elements {
		tags = append(tags, e.String)
	}
	var updatedBy string
	if c.UpdatedBy != nil {
		updatedBy = *c.UpdatedBy
	}
	var updatedAt time.Time
	if c.UpdatedAt.Valid {
		updatedAt = c.UpdatedAt.Time
	}

	return clients.Client{
		ID:    c.ID,
		Name:  c.Name,
		Tags:  tags,
		Owner: c.Owner,
		Credentials: clients.Credentials{
			Identity: c.Identity,
			Secret:   c.Secret,
		},
		Metadata:  metadata,
		CreatedAt: c.CreatedAt,
		UpdatedAt: updatedAt,
		UpdatedBy: updatedBy,
		Status:    c.Status,
	}, nil
}

func pageQuery(pm clients.Page) (string, error) {
	mq, _, err := postgres.CreateMetadataQuery("", pm.Metadata)
	if err != nil {
		return "", errors.Wrap(errors.ErrViewEntity, err)
	}
	var query []string
	var emq string
	if mq != "" {
		query = append(query, mq)
	}
	if len(pm.IDs) != 0 {
		query = append(query, fmt.Sprintf("id IN ('%s')", strings.Join(pm.IDs, "','")))
	}
	if pm.Name != "" {
		query = append(query, fmt.Sprintf("c.name = '%s'", pm.Name))
	}
	if pm.Tag != "" {
		query = append(query, fmt.Sprintf("'%s' = ANY(c.tags)", pm.Tag))
	}
	if pm.Status != mfclients.AllStatus {
		query = append(query, fmt.Sprintf("c.status = %d", pm.Status))
	}
	// For listing clients that the specified client owns but not sharedby
	if pm.Owner != "" && pm.SharedBy == "" {
		query = append(query, fmt.Sprintf("c.owner_id = '%s'", pm.Owner))
	}

	// For listing clients that the specified client owns and that are shared with the specified client
	if pm.Owner != "" && pm.SharedBy != "" {
		query = append(query, fmt.Sprintf("(c.owner_id = '%s' OR c.id IN (SELECT subject FROM policies WHERE object IN (SELECT object FROM policies WHERE subject = '%s' AND '%s'=ANY(actions))))", pm.Owner, pm.SharedBy, pm.Action))
	}
	// For listing clients that the specified client is shared with
	if pm.SharedBy != "" && pm.Owner == "" {
		query = append(query, fmt.Sprintf("c.owner_id != '%s' AND (c.id IN (SELECT subject FROM policies WHERE object IN (SELECT object FROM policies WHERE subject = '%s' AND '%s'=ANY(actions))))", pm.SharedBy, pm.SharedBy, pm.Action))
	}
	if len(query) > 0 {
		emq = fmt.Sprintf("WHERE %s", strings.Join(query, " AND "))
	}
	return emq, nil

}

func toDBClientsPage(pm clients.Page) (dbClientsPage, error) {
	_, data, err := postgres.CreateMetadataQuery("", pm.Metadata)
	if err != nil {
		return dbClientsPage{}, errors.Wrap(errors.ErrViewEntity, err)
	}
	return dbClientsPage{
		Name:     pm.Name,
		Metadata: data,
		Owner:    pm.Owner,
		Total:    pm.Total,
		Offset:   pm.Offset,
		Limit:    pm.Limit,
		Status:   pm.Status,
		Tag:      pm.Tag,
	}, nil
}

type dbClientsPage struct {
	GroupID  string           `db:"group_id"`
	Name     string           `db:"name"`
	Owner    string           `db:"owner_id"`
	Identity string           `db:"identity"`
	Metadata []byte           `db:"metadata"`
	Tag      string           `db:"tag"`
	Status   mfclients.Status `db:"status"`
	Total    uint64           `db:"total"`
	Limit    uint64           `db:"limit"`
	Offset   uint64           `db:"offset"`
}
