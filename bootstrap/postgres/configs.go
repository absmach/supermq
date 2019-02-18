//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
	"github.com/mainflux/mainflux/bootstrap"
	"github.com/mainflux/mainflux/logger"
)

const (
	duplicateErr    = "unique_violation"
	uuidErr         = "invalid input syntax for type uuid"
	configFieldsNum = 8
	chanFieldsNum   = 3
	connFieldsNum   = 2
)

var _ bootstrap.ConfigRepository = (*configRepository)(nil)

type configRepository struct {
	db  *sql.DB
	log logger.Logger
}

type dbChannel struct {
	ID       string      `json:"id"`
	Name     string      `json:"name"`
	Metadata interface{} `json:"metadata"`
}

// NewConfigRepository instantiates a PostgreSQL implementation of thing
// repository.
func NewConfigRepository(db *sql.DB, log logger.Logger) bootstrap.ConfigRepository {
	return &configRepository{db: db, log: log}
}

func (cr configRepository) Save(cfg bootstrap.Config, connections []string) (string, error) {
	q := `INSERT INTO configs (mainflux_thing, owner, name, mainflux_key, external_id, external_key, content, state)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	content := nullString(cfg.Content)
	name := nullString(cfg.Name)
	tx, err := cr.db.Begin()

	if err != nil {
		return "", err
	}

	if _, err := tx.Exec(q, cfg.MFThing, cfg.Owner, name, cfg.MFKey, cfg.ExternalID, cfg.ExternalKey, content, cfg.State); err != nil {
		e := err
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == duplicateErr {
			e = bootstrap.ErrConflict
		}

		cr.rollback("Failed to insert a Config", tx, err)

		return "", e
	}

	if err := insertChannels(cfg, tx); err != nil {
		cr.rollback("Failed to insert Channels", tx, err)

		return "", err
	}

	if err := insertConnections(cfg, connections, tx); err != nil {
		cr.rollback("Failed to insert connections", tx, err)

		return "", err
	}

	q = "DELETE FROM unknown_configs WHERE external_id = $1 AND external_key = $2"

	if _, err := tx.Exec(q, cfg.ExternalID, cfg.ExternalKey); err != nil {
		cr.rollback("Failed to remove from unknown", tx, err)

		return "", err
	}

	if err := tx.Commit(); err != nil {
		cr.rollback("Failed to commit Config save", tx, err)
	}

	return cfg.MFThing, nil
}

func (cr configRepository) RetrieveByID(key, id string) (bootstrap.Config, error) {
	q := `SELECT mainflux_thing, mainflux_key, external_id, external_key, name, content, state FROM configs WHERE mainflux_thing = $1 AND owner = $2`
	cfg := bootstrap.Config{MFThing: id, Owner: key, MFChannels: []bootstrap.Channel{}}
	var name, content sql.NullString
	if err := cr.db.QueryRow(q, id, key).
		Scan(&cfg.MFThing, &cfg.MFKey, &cfg.ExternalID, &cfg.ExternalKey, &name, &content, &cfg.State); err != nil {
		empty := bootstrap.Config{}
		if err == sql.ErrNoRows {
			return empty, bootstrap.ErrNotFound
		}
		return empty, err
	}

	q = `SELECT mainflux_channel, name, metadata FROM channels ch
	INNER JOIN connections conn
	ON ch.mainflux_channel = conn.channel_id AND ch.owner = conn.config_owner
	WHERE conn.config_id = $1 AND conn.config_owner = $2`

	rows, err := cr.db.Query(q, cfg.MFThing, cfg.Owner)
	if err != nil {
		cr.log.Error(fmt.Sprintf("Failed to retrieve connected due to %s", err))
		return bootstrap.Config{}, err
	}
	defer rows.Close()

	for rows.Next() {
		c := bootstrap.Channel{}
		if err = rows.Scan(&c.ID, &c.Name, &c.Metadata); err != nil {
			cr.log.Error(fmt.Sprintf("Failed to read connected thing due to %s", err))
			return bootstrap.Config{}, err
		}

		cfg.MFChannels = append(cfg.MFChannels, c)
	}

	cfg.Content = content.String
	cfg.Name = name.String

	return cfg, nil
}

func (cr configRepository) RetrieveAll(key string, filter bootstrap.Filter, offset, limit uint64) bootstrap.ConfigsPage {
	search, params := cr.retrieveAll(key, filter)
	n := len(params)

	q := `SELECT mainflux_thing, mainflux_key, external_id, external_key, name, content, state
	FROM configs %s ORDER BY mainflux_thing LIMIT $%d OFFSET $%d`
	q = fmt.Sprintf(q, search, n+1, n+2)

	rows, err := cr.db.Query(q, append(params, limit, offset)...)
	if err != nil {
		cr.log.Error(fmt.Sprintf("Failed to retrieve configs due to %s", err))
		return bootstrap.ConfigsPage{}
	}
	defer rows.Close()

	var name, content sql.NullString
	configs := []bootstrap.Config{}

	for rows.Next() {
		c := bootstrap.Config{Owner: key}
		if err := rows.Scan(&c.MFThing, &c.MFKey, &c.ExternalID, &c.ExternalKey, &name, &content, &c.State); err != nil {
			cr.log.Error(fmt.Sprintf("Failed to read retrieved config due to %s", err))
			return bootstrap.ConfigsPage{}
		}

		c.Name = name.String
		c.Content = content.String
		configs = append(configs, c)
	}

	q = fmt.Sprintf(`SELECT COUNT(*) FROM configs %s`, search)

	var total uint64
	if err := cr.db.QueryRow(q, params...).Scan(&total); err != nil {
		cr.log.Error(fmt.Sprintf("Failed to count configs due to %s", err))
		return bootstrap.ConfigsPage{}
	}

	return bootstrap.ConfigsPage{
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		Configs: configs,
	}
}

func (cr configRepository) RetrieveByExternalID(externalKey, externalID string) (bootstrap.Config, error) {
	q := `SELECT mainflux_thing, mainflux_key, owner, name, content, state FROM configs WHERE external_key = $1 AND external_id = $2`
	cfg := bootstrap.Config{ExternalID: externalID, ExternalKey: externalKey, MFChannels: []bootstrap.Channel{}}
	var name, content sql.NullString
	if err := cr.db.QueryRow(q, externalKey, externalID).
		Scan(&cfg.MFThing, &cfg.MFKey, &cfg.Owner, &name, &content, &cfg.State); err != nil {
		empty := bootstrap.Config{}
		if err == sql.ErrNoRows {
			return empty, bootstrap.ErrNotFound
		}
		return empty, err
	}

	q = `SELECT mainflux_channel, name, metadata FROM channels ch
	INNER JOIN connections conn
	ON ch.mainflux_channel = conn.channel_id AND ch.owner = conn.config_owner
	WHERE conn.config_id = $1 AND conn.config_owner = $2`

	rows, err := cr.db.Query(q, cfg.MFThing, cfg.Owner)
	if err != nil {
		cr.log.Error(fmt.Sprintf("Failed to retrieve connected due to %s", err))
		return bootstrap.Config{}, err
	}
	defer rows.Close()

	for rows.Next() {
		c := bootstrap.Channel{}
		if err = rows.Scan(&c.ID, &c.Name, &c.Metadata); err != nil {
			cr.log.Error(fmt.Sprintf("Failed to read connected thing due to %s", err))
			return bootstrap.Config{}, err
		}

		cfg.MFChannels = append(cfg.MFChannels, c)
	}

	cfg.Content = content.String
	cfg.Name = name.String

	return cfg, nil
}

func (cr configRepository) Update(cfg bootstrap.Config, connections []string) error {
	q := `UPDATE configs SET name = $1, content = $2, state = $3 WHERE mainflux_thing = $4 AND owner = $5`

	content := nullString(cfg.Content)
	name := nullString(cfg.Name)
	tx, err := cr.db.Begin()

	if err != nil {
		return err
	}

	if _, err := tx.Exec(q, name, content, cfg.State, cfg.MFThing, cfg.Owner); err != nil {
		e := err
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == duplicateErr {
			e = bootstrap.ErrConflict
		}

		cr.rollback("Failed to update a Config", tx, err)

		return e
	}

	if err = insertChannels(cfg, tx); err != nil {
		cr.rollback("Failed to insert Channels during the update", tx, err)

		return err
	}

	if err = updateConnections(cfg, connections, tx); err != nil {
		cr.rollback("Failed to update connections during the update", tx, err)

		return err
	}

	if err := tx.Commit(); err != nil {
		cr.rollback("Failed to commit Config update", tx, err)
	}

	return nil
}

func (cr configRepository) Remove(key, id string) error {
	params := []interface{}{id}
	q := `DELETE FROM configs WHERE mainflux_thing = $1`
	if key != "" {
		q = fmt.Sprintf("%s%s", q, " AND owner = $2")
		params = append(params, key)
	}

	cr.db.Exec(q, params...)

	return nil
}

func (cr configRepository) ChangeState(key, id string, state bootstrap.State) error {
	q := `UPDATE configs SET state = $1 WHERE mainflux_thing = $2 AND owner = $3;`

	res, err := cr.db.Exec(q, state, id, key)
	if err != nil {
		return err
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if cnt == 0 {
		return bootstrap.ErrNotFound
	}

	return nil
}

func (cr configRepository) Exist(key string, ids []string) ([]string, error) {
	q := "SELECT mainflux_channel FROM channels WHERE owner = $1 AND mainflux_channel IN ($2)"

	rows, err := cr.db.Query(q, key, pq.Array(ids))
	if err != nil {
		return []string{}, err
	}

	var connections []string
	for rows.Next() {
		var ch string
		if err = rows.Scan(&ch); err != nil {
			cr.log.Error(fmt.Sprintf("Failed to read retrieved channels due to %s", err))
			return []string{}, nil
		}

		connections = append(connections, ch)
	}

	return connections, nil
}

func (cr configRepository) SaveUnknown(key, id string) error {
	q := `INSERT INTO unknown_configs (external_id, external_key) VALUES ($1, $2)`

	if _, err := cr.db.Exec(q, id, key); err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == duplicateErr {
			return nil
		}
		return err
	}

	return nil
}

func (cr configRepository) RetrieveUnknown(offset, limit uint64) bootstrap.ConfigsPage {
	q := `SELECT external_id, external_key FROM unknown_configs LIMIT $1 OFFSET $2`
	rows, err := cr.db.Query(q, limit, offset)
	if err != nil {
		cr.log.Error(fmt.Sprintf("Failed to retrieve config due to %s", err))
		return bootstrap.ConfigsPage{}
	}
	defer rows.Close()

	items := []bootstrap.Config{}
	for rows.Next() {
		c := bootstrap.Config{}
		if err = rows.Scan(&c.ExternalID, &c.ExternalKey); err != nil {
			cr.log.Error(fmt.Sprintf("Failed to read retrieved config due to %s", err))
			return bootstrap.ConfigsPage{}
		}

		items = append(items, c)
	}

	q = fmt.Sprintf(`SELECT COUNT(*) FROM unknown_configs`)

	var total uint64
	if err := cr.db.QueryRow(q).Scan(&total); err != nil {
		cr.log.Error(fmt.Sprintf("Failed to count unknown configs due to %s", err))
		return bootstrap.ConfigsPage{}
	}

	return bootstrap.ConfigsPage{
		Total:   total,
		Offset:  offset,
		Limit:   limit,
		Configs: items,
	}
}

func (cr configRepository) retrieveAll(key string, filter bootstrap.Filter) (string, []interface{}) {
	template := `WHERE owner = $1 %s`
	params := []interface{}{key}
	// One empty string so that strings Join works if only one filter is applied.
	queries := []string{""}
	// Since key is the first param, start from 2.
	counter := 2
	for k, v := range filter.FullMatch {
		queries = append(queries, fmt.Sprintf("%s = $%d", k, counter))
		params = append(params, v)
		counter++
	}
	for k, v := range filter.PartialMatch {
		queries = append(queries, fmt.Sprintf("LOWER(%s) LIKE '%%' || $%d || '%%'", k, counter))
		params = append(params, v)
		counter++
	}

	f := strings.Join(queries, " AND ")

	return fmt.Sprintf(template, f), params
}

func insertChannels(cfg bootstrap.Config, tx *sql.Tx) error {
	if len(cfg.MFChannels) == 0 {
		return nil
	}

	q := `INSERT INTO channels (mainflux_channel, owner, name, metadata) VALUES `
	v := []interface{}{cfg.Owner}
	var vals []string
	// Since the first value is owner, start with the second one.
	count := 2
	for _, ch := range cfg.MFChannels {
		vals = append(vals, fmt.Sprintf("($%d, $1, $%d, $%d)", count, count+1, count+2))
		v = append(v, ch.ID, ch.Name, ch.Metadata)
		count += chanFieldsNum
	}

	q = fmt.Sprintf("%s%s%s", q, strings.Join(vals, ","), "ON CONFLICT (mainflux_channel) DO NOTHING")
	_, err := tx.Exec(q, v...)

	return err
}

func insertConnections(cfg bootstrap.Config, connections []string, tx *sql.Tx) error {
	if len(connections) == 0 {
		return nil
	}

	q := `INSERT INTO connections (config_id, channel_id, config_owner, channel_owner) VALUES`
	v := []interface{}{cfg.MFThing, cfg.Owner}
	var vals []string

	// Since the first value is Config ID and the second and third
	// are Config owner, start with  the second one.
	count := 3
	for _, id := range connections {
		vals = append(vals, fmt.Sprintf("($1, $%d, $2, $2)", count))
		v = append(v, id)
		count++
	}

	q = fmt.Sprintf("%s%s", q, strings.Join(vals, ","))
	_, err := tx.Exec(q, v...)

	return err
}

// Updating connections is removing old and adding new ones.
func updateConnections(cfg bootstrap.Config, connections []string, tx *sql.Tx) error {
	if len(connections) == 0 {
		return nil
	}

	q := `DELETE FROM connections
	WHERE config_id = $1 AND config_owner = $2 AND channel_owner = $2
	AND channel_id NOT IN ($3)`

	v := []interface{}{cfg.MFThing, cfg.Owner}
	v = append(v, pq.Array(connections))

	res, err := tx.Exec(q, v...)

	if err != nil {
		return err
	}

	cnt, err := res.RowsAffected()
	if err != nil {
		return err
	}

	q = `INSERT INTO connections (config_id, external_id, channel_id, config_owner, channel_owner) VALUES`
	v = []interface{}{cfg.MFThing, cfg.ExternalID, cfg.Owner}
	var vals []string

	// Since the first value is Config ID and the second is Config
	// owner, start with  the second one.
	count := 4
	for _, id := range connections {
		vals = append(vals, fmt.Sprintf("($1, $2, $%d, $3, $3)", count))
		v = append(v, id)
		count++
	}

	// Add connections for current list of channels. Ignore if already exists.
	q = fmt.Sprintf("%s%s%s", q, strings.Join(vals, ","), "ON CONFLICT (config_id, config_owner, channel_id, channel_owner) DO NOTHING")
	if _, err := tx.Exec(q, v...); err != nil {
		return err
	}

	if cnt == 0 {
		return nil
	}

	q = `DELETE FROM channels ch WHERE ch.mainflux_channel NOT IN (
    SELECT channel_id FROM connections)`

	_, err = tx.Exec(q)

	return err
}

func (cr configRepository) UpdateChannel(channel bootstrap.Channel) error {
	q := `UPDATE channels SET name = $1, metadata = $2 WHERE mainflux_channel = $3`
	if _, err := cr.db.Exec(q, channel.Name, channel.Metadata, channel.ID); err != nil {
		return err
	}

	return nil
}

func (cr configRepository) rollback(content string, tx *sql.Tx, err error) {
	cr.log.Error(fmt.Sprintf("%s %s", content, err))
	if err := tx.Rollback(); err != nil {
		cr.log.Error(fmt.Sprintf("Failed to rollback due to %s", err))
	}
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}

	return sql.NullString{
		String: s,
		Valid:  true,
	}
}
