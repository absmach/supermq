// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lib/pq" // required for DB access
	"github.com/mainflux/mainflux/commands"
	"github.com/mainflux/mainflux/pkg/errors"
)

const (
	errDuplicate  = "unique_violation"
	errFK         = "foreign_key_violation"
	errInvalid    = "invalid_text_representation"
	errTruncation = "string_data_right_truncation"
)

var _ commands.CommandRepository = (*commandRepository)(nil)

type commandRepository struct {
	db Database
}

// NewCommandRepository instantiates a PostgreSQL implementation of command
// repository.
func NewCommandRepository(db database) commands.CommandRepository {
	return &commandRepository{
		db: db,
	}
}

func (cr commandRepository) Save(ctx context.Context, cmds ...commands.Command) ([]commands.Command, error) {
	tx, err := cr.db.BeginTxx(ctx, nil)
	if err != nil {
		return []commands.Command{}, errors.Wrap(commands.ErrCreateEntity, err)
	}

	q := `INSERT INTO commands (id, owner, name, key, metadata)
		  VALUES (:id, :owner, :name, :key, :metadata);`

	for _, command := range cmds {
		dbcmd, err := toDBCommand(command)
		if err != nil {
			return []commands.Command{}, errors.Wrap(commands.ErrCreateEntity, err)
		}

		if _, err := tx.NamedExecContext(ctx, q, dbcmd); err != nil {
			tx.Rollback()
			pqErr, ok := err.(*pq.Error)
			if ok {
				switch pqErr.Code.Name() {
				case errInvalid, errTruncation:
					return []commands.Command{}, errors.Wrap(commands.ErrMalformedEntity, err)
				case errDuplicate:
					return []commands.Command{}, errors.Wrap(commands.ErrConflict, err)
				}
			}

			return []commands.Command{}, errors.Wrap(commands.ErrCreateEntity, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return []commands.Command{}, errors.Wrap(commands.ErrCreateEntity, err)
	}

	return cmds, nil
}

func (cr commandRepository) Update(ctx context.Context, cmd commands.Command) error {
	q := `UPDATE commands SET name = :name, metadata = :metadata WHERE owner = :owner AND id = :id;`

	dbcmd, err := toDBCommand(cmd)
	if err != nil {
		return errors.Wrap(commands.ErrUpdateEntity, err)
	}

	res, errdb := cr.db.NamedExecContext(ctx, q, dbcmd)
	if errdb != nil {
		pqErr, ok := errdb.(*pq.Error)
		if ok {
			switch pqErr.Code.Name() {
			case errInvalid, errTruncation:
				return errors.Wrap(commands.ErrMalformedEntity, errdb)
			}
		}

		return errors.Wrap(commands.ErrUpdateEntity, errdb)
	}

	cnt, errdb := res.RowsAffected()
	if errdb != nil {
		return errors.Wrap(commands.ErrUpdateEntity, errdb)
	}

	if cnt == 0 {
		return commands.ErrNotFound
	}

	return nil
}

func (cr commandRepository) UpdateKey(ctx context.Context, owner, id, key string) error {
	q := `UPDATE commands SET key = :key WHERE owner = :owner AND id = :id;`

	dbcmd := dbCommand{
		ID:    id,
		Owner: owner,
		Key:   key,
	}

	res, err := cr.db.NamedExecContext(ctx, q, dbcmd)
	if err != nil {
		pqErr, ok := err.(*pq.Error)
		if ok {
			switch pqErr.Code.Name() {
			case errInvalid:
				return errors.Wrap(commands.ErrMalformedEntity, err)
			case errDuplicate:
				return errors.Wrap(commands.ErrConflict, err)
			}
		}

		return errors.Wrap(commands.ErrUpdateEntity, err)
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(commands.ErrUpdateEntity, err)
	}

	if cnt == 0 {
		return commands.ErrNotFound
	}

	return nil
}

func (cr commandRepository) RetrieveByID(ctx context.Context, owner, id string) (commands.Command, error) {
	q := `SELECT name, key, metadata FROM commands WHERE id = $1 AND owner = $2;`

	dbcmd := dbCommand{
		ID:    id,
		Owner: owner,
	}

	if err := cr.db.QueryRowxContext(ctx, q, id, owner).StructScan(&dbcmd); err != nil {
		pqErr, ok := err.(*pq.Error)
		if err == sql.ErrNoRows || ok && errInvalid == pqErr.Code.Name() {
			return commands.Command{}, errors.Wrap(commands.ErrNotFound, err)
		}
		return commands.Command{}, errors.Wrap(commands.ErrSelectEntity, err)
	}

	return toCommand(dbcmd)
}

func (cr commandRepository) RetrieveByKey(ctx context.Context, key string) (string, error) {
	q := `SELECT id FROM commands WHERE key = $1;`

	var id string
	if err := cr.db.QueryRowxContext(ctx, q, key).Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			return "", errors.Wrap(commands.ErrNotFound, err)
		}
		return "", errors.Wrap(commands.ErrSelectEntity, err)
	}

	return id, nil
}

func (cr commandRepository) RetrieveByIDs(ctx context.Context, CommandIDs []string, pm commands.PageMetadata) (commands.CommandPage, error) {
	if len(CommandIDs) == 0 {
		return commands.CommandPage{}, nil
	}

	nq, name := getNameQuery(pm.Name)
	oq := getOrderQuery(pm.Order)
	dq := getDirQuery(pm.Dir)
	idq := fmt.Sprintf("WHERE id IN ('%s') ", strings.Join(CommandIDs, "','"))

	m, mq, err := getMetadataQuery(pm.Metadata)
	if err != nil {
		return commands.CommandPage{}, errors.Wrap(commands.ErrSelectEntity, err)
	}

	q := fmt.Sprintf(`SELECT id, owner, name, key, metadata FROM commands
					   %s%s%s ORDER BY %s %s LIMIT :limit OFFSET :offset;`, idq, mq, nq, oq, dq)

	params := map[string]interface{}{
		"limit":    pm.Limit,
		"offset":   pm.Offset,
		"name":     name,
		"metadata": m,
	}

	rows, err := cr.db.NamedQueryContext(ctx, q, params)
	if err != nil {
		return commands.CommandPage{}, errors.Wrap(commands.ErrSelectEntity, err)
	}
	defer rows.Close()

	var items []commands.Command
	for rows.Next() {
		dbcmd := dbCommand{}
		if err := rows.StructScan(&dbcmd); err != nil {
			return commands.CommandPage{}, errors.Wrap(commands.ErrSelectEntity, err)
		}

		cmd, err := toCommand(dbcmd)
		if err != nil {
			return commands.CommandPage{}, errors.Wrap(commands.ErrViewEntity, err)
		}

		items = append(items, cmd)
	}

	cq := fmt.Sprintf(`SELECT COUNT(*) FROM commands %s%s%s;`, idq, mq, nq)

	total, err := total(ctx, cr.db, cq, params)
	if err != nil {
		return commands.CommandPage{}, errors.Wrap(commands.ErrSelectEntity, err)
	}

	page := commands.CommandPage{
		Commands: items,
		PageMetadata: commands.PageMetadata{
			Total:  total,
			Offset: pm.Offset,
			Limit:  pm.Limit,
			Order:  pm.Order,
			Dir:    pm.Dir,
		},
	}

	return page, nil
}

func (cr commandRepository) RetrieveAll(ctx context.Context, owner string, pm commands.PageMetadata) (commands.CommandPage, error) {
	nq, name := getNameQuery(pm.Name)
	oq := getOrderQuery(pm.Order)
	dq := getDirQuery(pm.Dir)
	m, mq, err := getMetadataQuery(pm.Metadata)
	if err != nil {
		return commands.CommandPage{}, errors.Wrap(commands.ErrSelectEntity, err)
	}

	q := fmt.Sprintf(`SELECT id, name, key, metadata FROM commands
	      WHERE owner = :owner %s%s ORDER BY %s %s LIMIT :limit OFFSET :offset;`, mq, nq, oq, dq)
	params := map[string]interface{}{
		"owner":    owner,
		"limit":    pm.Limit,
		"offset":   pm.Offset,
		"name":     name,
		"metadata": m,
	}

	rows, err := cr.db.NamedQueryContext(ctx, q, params)
	if err != nil {
		return commands.CommandPage{}, errors.Wrap(commands.ErrSelectEntity, err)
	}
	defer rows.Close()

	var items []commands.Command
	for rows.Next() {
		dbcmd := dbCommand{Owner: owner}
		if err := rows.StructScan(&dbcmd); err != nil {
			return commands.CommandPage{}, errors.Wrap(commands.ErrSelectEntity, err)
		}

		cmd, err := toCommand(dbcmd)
		if err != nil {
			return commands.CommandPage{}, errors.Wrap(commands.ErrViewEntity, err)
		}

		items = append(items, cmd)
	}

	cq := fmt.Sprintf(`SELECT COUNT(*) FROM commands WHERE owner = :owner %s%s;`, nq, mq)

	total, err := total(ctx, cr.db, cq, params)
	if err != nil {
		return commands.CommandPage{}, errors.Wrap(commands.ErrSelectEntity, err)
	}

	page := commands.CommandPage{
		Commands: items,
		PageMetadata: commands.PageMetadata{
			Total:  total,
			Offset: pm.Offset,
			Limit:  pm.Limit,
			Order:  pm.Order,
			Dir:    pm.Dir,
		},
	}

	return page, nil
}

func (cr commandRepository) Remove(ctx context.Context, owner, id string) error {
	dbcmd := dbCommand{
		ID:    id,
		Owner: owner,
	}
	q := `DELETE FROM commands WHERE id = :id AND owner = :owner;`
	if _, err := cr.db.NamedExecContext(ctx, q, dbcmd); err != nil {
		return errors.Wrap(commands.ErrRemoveEntity, err)
	}
	return nil
}

type dbCommand struct {
	ID       string `db:"id"`
	Owner    string `db:"owner"`
	Name     string `db:"name"`
	Key      string `db:"key"`
	Metadata []byte `db:"metadata"`
}

func toDBCommand(cmd commands.Command) (dbCommand, error) {
	data := []byte("{}")
	if len(cmd.Metadata) > 0 {
		b, err := json.Marshal(cmd.Metadata)
		if err != nil {
			return dbCommand{}, errors.Wrap(commands.ErrMalformedEntity, err)
		}
		data = b
	}

	return dbCommand{
		ID:       cmd.ID,
		Owner:    cmd.Owner,
		Name:     cmd.Name,
		Metadata: data,
	}, nil
}

func toCommand(dbcmd dbCommand) (commands.Command, error) {
	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(dbcmd.Metadata), &metadata); err != nil {
		return commands.Command{}, errors.Wrap(commands.ErrMalformedEntity, err)
	}

	return commands.Command{
		ID:       dbcmd.ID,
		Owner:    dbcmd.Owner,
		Name:     dbcmd.Name,
		Metadata: metadata,
	}, nil
}

func getNameQuery(name string) (string, string) {
	if name == "" {
		return "", ""
	}
	name = fmt.Sprintf(`%%%s%%`, strings.ToLower(name))
	nq := ` AND LOWER(name) LIKE :name`
	return nq, name
}

func getOrderQuery(order string) string {
	switch order {
	case "name":
		return "name"
	default:
		return "id"
	}
}

func getConnOrderQuery(order string, level string) string {
	switch order {
	case "name":
		return level + ".name"
	default:
		return level + ".id"
	}
}

func getDirQuery(dir string) string {
	switch dir {
	case "asc":
		return "ASC"
	default:
		return "DESC"
	}
}

func getMetadataQuery(m commands.Metadata) ([]byte, string, error) {
	mq := ""
	mb := []byte("{}")
	if len(m) > 0 {
		mq = ` AND metadata @> :metadata`

		b, err := json.Marshal(m)
		if err != nil {
			return nil, "", err
		}
		mb = b
	}
	return mb, mq, nil
}

func total(ctx context.Context, db Database, query string, params interface{}) (uint64, error) {
	rows, err := db.NamedQueryContext(ctx, query, params)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	total := uint64(0)
	if rows.Next() {
		if err := rows.Scan(&total); err != nil {
			return 0, err
		}
	}
	return total, nil
}
