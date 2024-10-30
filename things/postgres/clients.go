// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/absmach/magistrala/internal/api"
	"github.com/absmach/magistrala/pkg/apiutil"
	"github.com/absmach/magistrala/pkg/connections"
	"github.com/absmach/magistrala/pkg/errors"
	repoerr "github.com/absmach/magistrala/pkg/errors/repository"
	"github.com/absmach/magistrala/pkg/postgres"
	rolesPostgres "github.com/absmach/magistrala/pkg/roles/repo/postgres"
	"github.com/absmach/magistrala/things"
	"github.com/jackc/pgtype"
)

const (
	entityTableName      = "clients"
	entityIDColumnName   = "id"
	rolesTableNamePrefix = "things"
)

var _ things.Repository = (*clientRepo)(nil)

type clientRepo struct {
	DB postgres.Database
	rolesPostgres.Repository
}

// NewRepository instantiates a PostgreSQL
// implementation of Clients repository.
func NewRepository(db postgres.Database) things.Repository {
	repo := rolesPostgres.NewRepository(db, rolesTableNamePrefix, entityTableName, entityIDColumnName)

	return &clientRepo{
		DB:         db,
		Repository: repo,
	}
}

func (repo *clientRepo) Save(ctx context.Context, clients ...things.Client) ([]things.Client, error) {
	var dbClients []DBClient

	for _, client := range clients {
		dbcli, err := ToDBClient(client)
		if err != nil {
			return []things.Client{}, errors.Wrap(repoerr.ErrCreateEntity, err)
		}
		dbClients = append(dbClients, dbcli)
	}
	q := `INSERT INTO clients (id, name, tags, domain_id, parent_group_id, identity, secret, metadata, created_at, updated_at, updated_by, status)
	VALUES (:id, :name, :tags, :domain_id, :parent_group_id, :identity, :secret, :metadata, :created_at, :updated_at, :updated_by, :status)
	RETURNING id, name, tags, identity, secret, metadata, COALESCE(domain_id, '') AS domain_id, COALESCE(parent_group_id, '') AS  parent_group_id, status, created_at, updated_at, updated_by`

	row, err := repo.DB.NamedQueryContext(ctx, q, dbClients)
	if err != nil {
		return []things.Client{}, postgres.HandleError(repoerr.ErrCreateEntity, err)
	}

	defer row.Close()

	var reClients []things.Client
	for row.Next() {
		dbcli := DBClient{}
		if err := row.StructScan(&dbcli); err != nil {
			return []things.Client{}, errors.Wrap(repoerr.ErrFailedOpDB, err)
		}

		client, err := ToClient(dbcli)
		if err != nil {
			return []things.Client{}, errors.Wrap(repoerr.ErrFailedOpDB, err)
		}
		reClients = append(reClients, client)
	}
	return reClients, nil
}

func (repo *clientRepo) RetrieveBySecret(ctx context.Context, key string) (things.Client, error) {
	q := fmt.Sprintf(`SELECT id, name, tags, COALESCE(domain_id, '') AS domain_id,  COALESCE(parent_group_id, '') AS parent_group_id, identity, secret, metadata, created_at, updated_at, updated_by, status
        FROM clients
        WHERE secret = :secret AND status = %d`, things.EnabledStatus)

	dbt := DBClient{
		Secret: key,
	}

	rows, err := repo.DB.NamedQueryContext(ctx, q, dbt)
	if err != nil {
		return things.Client{}, postgres.HandleError(repoerr.ErrViewEntity, err)
	}
	defer rows.Close()

	dbt = DBClient{}
	if rows.Next() {
		if err = rows.StructScan(&dbt); err != nil {
			return things.Client{}, postgres.HandleError(repoerr.ErrViewEntity, err)
		}

		thing, err := ToClient(dbt)
		if err != nil {
			return things.Client{}, errors.Wrap(repoerr.ErrFailedOpDB, err)
		}

		return thing, nil
	}

	return things.Client{}, repoerr.ErrNotFound
}

func (repo *clientRepo) Update(ctx context.Context, thing things.Client) (things.Client, error) {
	var query []string
	var upq string
	if thing.Name != "" {
		query = append(query, "name = :name,")
	}
	if thing.Metadata != nil {
		query = append(query, "metadata = :metadata,")
	}
	if len(query) > 0 {
		upq = strings.Join(query, " ")
	}

	q := fmt.Sprintf(`UPDATE clients SET %s updated_at = :updated_at, updated_by = :updated_by
        WHERE id = :id AND status = :status
        RETURNING id, name, tags, identity, secret,  metadata, COALESCE(domain_id, '') AS domain_id, COALESCE(parent_group_id, '') AS parent_group_id, status, created_at, updated_at, updated_by`,
		upq)
	thing.Status = things.EnabledStatus
	return repo.update(ctx, thing, q)
}

func (repo *clientRepo) UpdateTags(ctx context.Context, thing things.Client) (things.Client, error) {
	q := `UPDATE clients SET tags = :tags, updated_at = :updated_at, updated_by = :updated_by
        WHERE id = :id AND status = :status
        RETURNING id, name, tags, identity, metadata, COALESCE(domain_id, '') AS domain_id, COALESCE(parent_group_id, '') AS parent_group_id, status, created_at, updated_at, updated_by`
	thing.Status = things.EnabledStatus
	return repo.update(ctx, thing, q)
}

func (repo *clientRepo) UpdateIdentity(ctx context.Context, thing things.Client) (things.Client, error) {
	q := `UPDATE clients SET identity = :identity, updated_at = :updated_at, updated_by = :updated_by
        WHERE id = :id AND status = :status
        RETURNING id, name, tags, identity, metadata, COALESCE(domain_id, '') AS domain_id, status, COALESCE(parent_group_id, '') AS parent_group_id, created_at, updated_at, updated_by`
	thing.Status = things.EnabledStatus
	return repo.update(ctx, thing, q)
}

func (repo *clientRepo) UpdateSecret(ctx context.Context, thing things.Client) (things.Client, error) {
	q := `UPDATE clients SET secret = :secret, updated_at = :updated_at, updated_by = :updated_by
        WHERE id = :id AND status = :status
        RETURNING id, name, tags, identity, metadata, COALESCE(domain_id, '') AS domain_id, COALESCE(parent_group_id, '') AS parent_group_id, status, created_at, updated_at, updated_by`
	thing.Status = things.EnabledStatus
	return repo.update(ctx, thing, q)
}

func (repo *clientRepo) ChangeStatus(ctx context.Context, thing things.Client) (things.Client, error) {
	q := `UPDATE clients SET status = :status, updated_at = :updated_at, updated_by = :updated_by
		WHERE id = :id
        RETURNING id, name, tags, identity, metadata, COALESCE(domain_id, '') AS domain_id, COALESCE(parent_group_id, '') AS parent_group_id, status, created_at, updated_at, updated_by`

	return repo.update(ctx, thing, q)
}

func (repo *clientRepo) RetrieveByID(ctx context.Context, id string) (things.Client, error) {
	q := `SELECT id, name, tags, COALESCE(domain_id, '') AS domain_id, COALESCE(parent_group_id, '') AS parent_group_id, identity, secret, metadata, created_at, updated_at, updated_by, status
        FROM clients WHERE id = :id`

	dbt := DBClient{
		ID: id,
	}

	row, err := repo.DB.NamedQueryContext(ctx, q, dbt)
	if err != nil {
		return things.Client{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}
	defer row.Close()

	dbt = DBClient{}
	if row.Next() {
		if err := row.StructScan(&dbt); err != nil {
			return things.Client{}, errors.Wrap(repoerr.ErrViewEntity, err)
		}

		return ToClient(dbt)
	}

	return things.Client{}, repoerr.ErrNotFound
}

func (repo *clientRepo) RetrieveAll(ctx context.Context, pm things.Page) (things.ClientsPage, error) {
	query, err := PageQuery(pm)
	if err != nil {
		return things.ClientsPage{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}
	query = applyOrdering(query, pm)

	q := fmt.Sprintf(`SELECT c.id, c.name, c.tags, c.identity, c.metadata, COALESCE(c.domain_id, '') AS domain_id, COALESCE(parent_group_id, '') AS parent_group_id, c.status,
					c.created_at, c.updated_at, COALESCE(c.updated_by, '') AS updated_by FROM clients c %s ORDER BY c.created_at LIMIT :limit OFFSET :offset;`, query)

	dbPage, err := ToDBClientsPage(pm)
	if err != nil {
		return things.ClientsPage{}, errors.Wrap(repoerr.ErrFailedToRetrieveAllGroups, err)
	}
	rows, err := repo.DB.NamedQueryContext(ctx, q, dbPage)
	if err != nil {
		return things.ClientsPage{}, errors.Wrap(repoerr.ErrFailedToRetrieveAllGroups, err)
	}
	defer rows.Close()

	var items []things.Client
	for rows.Next() {
		dbt := DBClient{}
		if err := rows.StructScan(&dbt); err != nil {
			return things.ClientsPage{}, errors.Wrap(repoerr.ErrViewEntity, err)
		}

		c, err := ToClient(dbt)
		if err != nil {
			return things.ClientsPage{}, err
		}

		items = append(items, c)
	}
	cq := fmt.Sprintf(`SELECT COUNT(*) FROM clients c %s;`, query)

	total, err := postgres.Total(ctx, repo.DB, cq, dbPage)
	if err != nil {
		return things.ClientsPage{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}

	page := things.ClientsPage{
		Clients: items,
		Page: things.Page{
			Total:  total,
			Offset: pm.Offset,
			Limit:  pm.Limit,
		},
	}

	return page, nil
}

func (repo *clientRepo) SearchClients(ctx context.Context, pm things.Page) (things.ClientsPage, error) {
	query, err := PageQuery(pm)
	if err != nil {
		return things.ClientsPage{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}

	tq := query
	query = applyOrdering(query, pm)

	q := fmt.Sprintf(`SELECT c.id, c.name, c.created_at, c.updated_at FROM clients c %s LIMIT :limit OFFSET :offset;`, query)

	dbPage, err := ToDBClientsPage(pm)
	if err != nil {
		return things.ClientsPage{}, errors.Wrap(repoerr.ErrFailedToRetrieveAllGroups, err)
	}

	rows, err := repo.DB.NamedQueryContext(ctx, q, dbPage)
	if err != nil {
		return things.ClientsPage{}, errors.Wrap(repoerr.ErrFailedToRetrieveAllGroups, err)
	}
	defer rows.Close()

	var items []things.Client
	for rows.Next() {
		dbt := DBClient{}
		if err := rows.StructScan(&dbt); err != nil {
			return things.ClientsPage{}, errors.Wrap(repoerr.ErrViewEntity, err)
		}

		c, err := ToClient(dbt)
		if err != nil {
			return things.ClientsPage{}, err
		}

		items = append(items, c)
	}

	cq := fmt.Sprintf(`SELECT COUNT(*) FROM clients c %s;`, tq)
	total, err := postgres.Total(ctx, repo.DB, cq, dbPage)
	if err != nil {
		return things.ClientsPage{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}

	page := things.ClientsPage{
		Clients: items,
		Page: things.Page{
			Total:  total,
			Offset: pm.Offset,
			Limit:  pm.Limit,
		},
	}

	return page, nil
}

func (repo *clientRepo) RetrieveAllByIDs(ctx context.Context, pm things.Page) (things.ClientsPage, error) {
	if (len(pm.IDs) == 0) && (pm.Domain == "") {
		return things.ClientsPage{
			Page: things.Page{Total: pm.Total, Offset: pm.Offset, Limit: pm.Limit},
		}, nil
	}
	query, err := PageQuery(pm)
	if err != nil {
		return things.ClientsPage{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}
	query = applyOrdering(query, pm)

	q := fmt.Sprintf(`SELECT c.id, c.name, c.tags, c.identity, c.metadata, COALESCE(c.domain_id, '') AS domain_id, COALESCE(parent_group_id, '') AS parent_group_id, c.status,
					c.created_at, c.updated_at, COALESCE(c.updated_by, '') AS updated_by FROM clients c %s ORDER BY c.created_at LIMIT :limit OFFSET :offset;`, query)

	dbPage, err := ToDBClientsPage(pm)
	if err != nil {
		return things.ClientsPage{}, errors.Wrap(repoerr.ErrFailedToRetrieveAllGroups, err)
	}
	rows, err := repo.DB.NamedQueryContext(ctx, q, dbPage)
	if err != nil {
		return things.ClientsPage{}, errors.Wrap(repoerr.ErrFailedToRetrieveAllGroups, err)
	}
	defer rows.Close()

	var items []things.Client
	for rows.Next() {
		dbt := DBClient{}
		if err := rows.StructScan(&dbt); err != nil {
			return things.ClientsPage{}, errors.Wrap(repoerr.ErrViewEntity, err)
		}

		c, err := ToClient(dbt)
		if err != nil {
			return things.ClientsPage{}, err
		}

		items = append(items, c)
	}
	cq := fmt.Sprintf(`SELECT COUNT(*) FROM clients c %s;`, query)

	total, err := postgres.Total(ctx, repo.DB, cq, dbPage)
	if err != nil {
		return things.ClientsPage{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}

	page := things.ClientsPage{
		Clients: items,
		Page: things.Page{
			Total:  total,
			Offset: pm.Offset,
			Limit:  pm.Limit,
		},
	}

	return page, nil
}

func (repo *clientRepo) update(ctx context.Context, thing things.Client, query string) (things.Client, error) {
	dbc, err := ToDBClient(thing)
	if err != nil {
		return things.Client{}, errors.Wrap(repoerr.ErrUpdateEntity, err)
	}

	row, err := repo.DB.NamedQueryContext(ctx, query, dbc)
	if err != nil {
		return things.Client{}, postgres.HandleError(repoerr.ErrUpdateEntity, err)
	}
	defer row.Close()

	dbc = DBClient{}
	if row.Next() {
		if err := row.StructScan(&dbc); err != nil {
			return things.Client{}, errors.Wrap(repoerr.ErrUpdateEntity, err)
		}

		return ToClient(dbc)
	}

	return things.Client{}, repoerr.ErrNotFound
}

func (repo *clientRepo) Delete(ctx context.Context, clientIDs ...string) error {
	q := "DELETE FROM clients AS c  WHERE c.id = ANY(:client_ids) ;"

	params := map[string]interface{}{
		"client_ids": clientIDs,
	}
	result, err := repo.DB.NamedExecContext(ctx, q, params)
	if err != nil {
		return postgres.HandleError(repoerr.ErrRemoveEntity, err)
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return repoerr.ErrNotFound
	}

	return nil
}

type DBClient struct {
	ID          string           `db:"id"`
	Name        string           `db:"name,omitempty"`
	Tags        pgtype.TextArray `db:"tags,omitempty"`
	Identity    string           `db:"identity"`
	Domain      string           `db:"domain_id"`
	ParentGroup sql.NullString   `db:"parent_group_id,omitempty"`
	Secret      string           `db:"secret"`
	Metadata    []byte           `db:"metadata,omitempty"`
	CreatedAt   time.Time        `db:"created_at,omitempty"`
	UpdatedAt   sql.NullTime     `db:"updated_at,omitempty"`
	UpdatedBy   *string          `db:"updated_by,omitempty"`
	Status      things.Status    `db:"status,omitempty"`
}

func ToDBClient(c things.Client) (DBClient, error) {
	data := []byte("{}")
	if len(c.Metadata) > 0 {
		b, err := json.Marshal(c.Metadata)
		if err != nil {
			return DBClient{}, errors.Wrap(repoerr.ErrMalformedEntity, err)
		}
		data = b
	}
	var tags pgtype.TextArray
	if err := tags.Set(c.Tags); err != nil {
		return DBClient{}, err
	}
	var updatedBy *string
	if c.UpdatedBy != "" {
		updatedBy = &c.UpdatedBy
	}
	var updatedAt sql.NullTime
	if c.UpdatedAt != (time.Time{}) {
		updatedAt = sql.NullTime{Time: c.UpdatedAt, Valid: true}
	}

	return DBClient{
		ID:          c.ID,
		Name:        c.Name,
		Tags:        tags,
		Domain:      c.Domain,
		ParentGroup: toNullString(c.ParentGroup),
		Identity:    c.Credentials.Identity,
		Secret:      c.Credentials.Secret,
		Metadata:    data,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   updatedAt,
		UpdatedBy:   updatedBy,
		Status:      c.Status,
	}, nil
}

func ToClient(t DBClient) (things.Client, error) {
	var metadata things.Metadata
	if t.Metadata != nil {
		if err := json.Unmarshal([]byte(t.Metadata), &metadata); err != nil {
			return things.Client{}, errors.Wrap(errors.ErrMalformedEntity, err)
		}
	}
	var tags []string
	for _, e := range t.Tags.Elements {
		tags = append(tags, e.String)
	}
	var updatedBy string
	if t.UpdatedBy != nil {
		updatedBy = *t.UpdatedBy
	}
	var updatedAt time.Time
	if t.UpdatedAt.Valid {
		updatedAt = t.UpdatedAt.Time
	}

	thg := things.Client{
		ID:          t.ID,
		Name:        t.Name,
		Tags:        tags,
		Domain:      t.Domain,
		ParentGroup: toString(t.ParentGroup),
		Credentials: things.Credentials{
			Identity: t.Identity,
			Secret:   t.Secret,
		},
		Metadata:  metadata,
		CreatedAt: t.CreatedAt,
		UpdatedAt: updatedAt,
		UpdatedBy: updatedBy,
		Status:    t.Status,
	}
	return thg, nil
}

func ToDBClientsPage(pm things.Page) (dbClientsPage, error) {
	_, data, err := postgres.CreateMetadataQuery("", pm.Metadata)
	if err != nil {
		return dbClientsPage{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}
	return dbClientsPage{
		Name:     pm.Name,
		Identity: pm.Identity,
		Id:       pm.Id,
		Metadata: data,
		Domain:   pm.Domain,
		Total:    pm.Total,
		Offset:   pm.Offset,
		Limit:    pm.Limit,
		Status:   pm.Status,
		Tag:      pm.Tag,
	}, nil
}

type dbClientsPage struct {
	Total    uint64        `db:"total"`
	Limit    uint64        `db:"limit"`
	Offset   uint64        `db:"offset"`
	Name     string        `db:"name"`
	Id       string        `db:"id"`
	Domain   string        `db:"domain_id"`
	Identity string        `db:"identity"`
	Metadata []byte        `db:"metadata"`
	Tag      string        `db:"tag"`
	Status   things.Status `db:"status"`
	GroupID  string        `db:"group_id"`
}

func PageQuery(pm things.Page) (string, error) {
	mq, _, err := postgres.CreateMetadataQuery("", pm.Metadata)
	if err != nil {
		return "", errors.Wrap(errors.ErrMalformedEntity, err)
	}

	var query []string
	if pm.Name != "" {
		query = append(query, "name ILIKE '%' || :name || '%'")
	}
	if pm.Identity != "" {
		query = append(query, "identity ILIKE '%' || :identity || '%'")
	}
	if pm.Id != "" {
		query = append(query, "id ILIKE '%' || :id || '%'")
	}
	if pm.Tag != "" {
		query = append(query, "EXISTS (SELECT 1 FROM unnest(tags) AS tag WHERE tag ILIKE '%' || :tag || '%')")
	}
	// If there are search params presents, use search and ignore other options.
	// Always combine role with search params, so len(query) > 1.
	if len(query) > 1 {
		return fmt.Sprintf("WHERE %s", strings.Join(query, " AND ")), nil
	}

	if mq != "" {
		query = append(query, mq)
	}

	if len(pm.IDs) != 0 {
		query = append(query, fmt.Sprintf("id IN ('%s')", strings.Join(pm.IDs, "','")))
	}
	if pm.Status != things.AllStatus {
		query = append(query, "c.status = :status")
	}
	if pm.Domain != "" {
		query = append(query, "c.domain_id = :domain_id")
	}
	var emq string
	if len(query) > 0 {
		emq = fmt.Sprintf("WHERE %s", strings.Join(query, " AND "))
	}
	return emq, nil
}

func applyOrdering(emq string, pm things.Page) string {
	switch pm.Order {
	case "name", "identity", "created_at", "updated_at":
		emq = fmt.Sprintf("%s ORDER BY %s", emq, pm.Order)
		if pm.Dir == api.AscDir || pm.Dir == api.DescDir {
			emq = fmt.Sprintf("%s %s", emq, pm.Dir)
		}
	}
	return emq
}

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}

	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

func toString(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

func (repo *clientRepo) RetrieveByIds(ctx context.Context, ids []string) (things.ClientsPage, error) {
	if len(ids) == 0 {
		return things.ClientsPage{}, nil
	}

	pm := things.Page{IDs: ids}
	query, err := PageQuery(pm)
	if err != nil {
		return things.ClientsPage{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}

	q := fmt.Sprintf(`SELECT c.id, c.name, c.tags, c.identity, c.metadata, COALESCE(c.domain_id, '') AS domain_id,  COALESCE(parent_group_id, '') AS parent_group_id, c.status,
					c.created_at, c.updated_at, COALESCE(c.updated_by, '') AS updated_by FROM clients c %s ORDER BY c.created_at`, query)

	dbPage, err := ToDBClientsPage(pm)
	if err != nil {
		return things.ClientsPage{}, errors.Wrap(repoerr.ErrFailedToRetrieveAllGroups, err)
	}
	rows, err := repo.DB.NamedQueryContext(ctx, q, dbPage)
	if err != nil {
		return things.ClientsPage{}, errors.Wrap(repoerr.ErrFailedToRetrieveAllGroups, err)
	}
	defer rows.Close()

	var items []things.Client
	for rows.Next() {
		dbc := DBClient{}
		if err := rows.StructScan(&dbc); err != nil {
			return things.ClientsPage{}, errors.Wrap(repoerr.ErrViewEntity, err)
		}

		c, err := ToClient(dbc)
		if err != nil {
			return things.ClientsPage{}, err
		}

		items = append(items, c)
	}
	cq := fmt.Sprintf(`SELECT COUNT(*) FROM clients c %s;`, query)

	total, err := postgres.Total(ctx, repo.DB, cq, dbPage)
	if err != nil {
		return things.ClientsPage{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}

	page := things.ClientsPage{
		Clients: items,
		Page: things.Page{
			Total:  total,
			Offset: pm.Offset,
			Limit:  total,
		},
	}

	return page, nil
}

func (repo *clientRepo) AddConnections(ctx context.Context, conns []things.Connection) error {

	dbConns := toDBConnections(conns)

	q := `INSERT INTO connections (channel_id, domain_id, thing_id, type)
			VALUES (:channel_id, :domain_id, :thing_id, :type);`

	if _, err := repo.DB.NamedExecContext(ctx, q, dbConns); err != nil {
		return postgres.HandleError(repoerr.ErrCreateEntity, err)
	}

	return nil

}

func (repo *clientRepo) RemoveConnections(ctx context.Context, conns []things.Connection) (retErr error) {
	tx, err := repo.DB.BeginTxx(ctx, nil)
	if err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}
	defer func() {
		if retErr != nil {
			if errRollBack := tx.Rollback(); errRollBack != nil {
				retErr = errors.Wrap(retErr, errors.Wrap(apiutil.ErrRollbackTx, errRollBack))
			}
		}
	}()

	query := `DELETE FROM connections WHERE channel_id = :channel_id AND domain_id = :domain_id AND thing_id = :thing_id`

	for _, conn := range conns {
		if uint8(conn.Type) > 0 {
			query = query + " AND type = :type "
		}
		dbConn := toDBConnection(conn)
		if _, err := tx.NamedExec(query, dbConn); err != nil {
			return errors.Wrap(repoerr.ErrRemoveEntity, errors.Wrap(fmt.Errorf("failed to delete connection for channel_id: %s, domain_id: %s thing_id %s", conn.ChannelID, conn.DomainID, conn.ThingID), err))
		}
	}
	if err := tx.Commit(); err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}
	return nil
}

func (repo *clientRepo) SetParentGroup(ctx context.Context, th things.Client) error {
	q := "UPDATE clients SET parent_group_id = :parent_group_id, updated_at = :updated_at, updated_by = :updated_by WHERE id = :id"

	dbcli, err := ToDBClient(th)
	if err != nil {
		return errors.Wrap(repoerr.ErrUpdateEntity, err)
	}
	result, err := repo.DB.NamedExecContext(ctx, q, dbcli)
	if err != nil {
		return postgres.HandleError(repoerr.ErrUpdateEntity, err)
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return repoerr.ErrNotFound
	}
	return nil
}

func (repo *clientRepo) RemoveParentGroup(ctx context.Context, th things.Client) error {
	q := "UPDATE clients SET parent_group_id = NULL, updated_at = :updated_at, updated_by = :updated_by WHERE id = :id"
	dbcli, err := ToDBClient(th)
	if err != nil {
		return errors.Wrap(repoerr.ErrUpdateEntity, err)
	}
	result, err := repo.DB.NamedExecContext(ctx, q, dbcli)
	if err != nil {
		return postgres.HandleError(repoerr.ErrRemoveEntity, err)
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return repoerr.ErrNotFound
	}
	return nil
}

func (repo *clientRepo) ThingConnectionsCount(ctx context.Context, id string) (uint64, error) {
	query := `SELECT COUNT(*) FROM connections WHERE thing_id = :thing_id`
	dbConn := dbConnection{ThingID: id}

	total, err := postgres.Total(ctx, repo.DB, query, dbConn)
	if err != nil {
		return 0, postgres.HandleError(repoerr.ErrViewEntity, err)
	}
	return total, nil
}

func (repo *clientRepo) DoesThingHaveConnections(ctx context.Context, id string) (bool, error) {
	query := `SELECT 1 FROM connections WHERE thing_id = :thing_id`
	dbConn := dbConnection{ThingID: id}

	rows, err := repo.DB.NamedQueryContext(ctx, query, dbConn)
	if err != nil {
		return false, postgres.HandleError(repoerr.ErrViewEntity, err)
	}
	defer rows.Close()

	return rows.Next(), nil
}

func (repo *clientRepo) RemoveChannelConnections(ctx context.Context, channelID string) error {
	query := `DELETE FROM connections WHERE channel_id = :channel_id`

	dbConn := dbConnection{ChannelID: channelID}
	if _, err := repo.DB.NamedExecContext(ctx, query, dbConn); err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}
	return nil
}

func (repo *clientRepo) RemoveThingConnections(ctx context.Context, thingID string) error {
	query := `DELETE FROM connections WHERE thing_id = :thing_id`

	dbConn := dbConnection{ThingID: thingID}
	if _, err := repo.DB.NamedExecContext(ctx, query, dbConn); err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}
	return nil
}

func (repo *clientRepo) RetrieveParentGroupThings(ctx context.Context, parentGroupID string) ([]things.Client, error) {
	query := `SELECT c.id, c.name, c.tags,  c.metadata, COALESCE(c.domain_id, '') AS domain_id, COALESCE(parent_group_id, '') AS parent_group_id, c.status,
					c.created_at, c.updated_at, COALESCE(c.updated_by, '') AS updated_by FROM clients c WHERE c.parent_group_id = :parent_group_id ;`

	rows, err := repo.DB.NamedQueryContext(ctx, query, DBClient{ParentGroup: toNullString(parentGroupID)})
	if err != nil {
		return []things.Client{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}
	defer rows.Close()

	var ths []things.Client
	for rows.Next() {
		dbTh := DBClient{}
		if err := rows.StructScan(&dbTh); err != nil {
			return []things.Client{}, errors.Wrap(repoerr.ErrViewEntity, err)
		}

		th, err := ToClient(dbTh)
		if err != nil {
			return []things.Client{}, err
		}

		ths = append(ths, th)
	}
	return ths, nil
}

func (repo *clientRepo) UnsetParentGroupFromThings(ctx context.Context, parentGroupID string) error {
	query := "UPDATE clients SET parent_group_id = NULL WHERE parent_group_id = :parent_group_id"

	if _, err := repo.DB.NamedExecContext(ctx, query, DBClient{ParentGroup: toNullString(parentGroupID)}); err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}
	return nil
}

type dbConnection struct {
	ThingID   string               `db:"thing_id"`
	ChannelID string               `db:"channel_id"`
	DomainID  string               `db:"domain_id"`
	Type      connections.ConnType `db:"type"`
}

func toDBConnections(conns []things.Connection) []dbConnection {
	var dbconns []dbConnection
	for _, conn := range conns {
		dbconns = append(dbconns, toDBConnection(conn))
	}
	return dbconns
}

func toDBConnection(conn things.Connection) dbConnection {
	return dbConnection{
		ThingID:   conn.ThingID,
		ChannelID: conn.ChannelID,
		DomainID:  conn.DomainID,
		Type:      conn.Type,
	}
}
