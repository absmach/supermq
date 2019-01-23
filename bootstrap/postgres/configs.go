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

	"github.com/mainflux/mainflux/bootstrap"

	"github.com/lib/pq" // required for DB access
	"github.com/mainflux/mainflux/logger"
)

const (
	duplicateErr     = "unique_violation"
	uuidErr          = "invalid input syntax for type uuid"
	configFieldsNum  = 8
	channelFieldsNum = 3
)

var _ bootstrap.ConfigRepository = (*configRepository)(nil)

type configRepository struct {
	db  *sql.DB
	log logger.Logger
}

// NewConfigRepository instantiates a PostgreSQL implementation of thing
// repository.
func NewConfigRepository(db *sql.DB, log logger.Logger) bootstrap.ConfigRepository {
	return &configRepository{db: db, log: log}
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

func (cr configRepository) Save(cfg bootstrap.Config) (string, error) {
	q := `INSERT INTO configs (mainflux_thing, owner, mainflux_key, external_id, external_key, content, state)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	query, v := generateValues(q, cfg)

	if _, err := cr.db.Exec(query, v...); err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == duplicateErr {
			return "", bootstrap.ErrConflict
		}
		return "", err
	}

	return cfg.MFThing, nil
}

func (cr configRepository) RetrieveByID(key, id string) (bootstrap.Config, error) {
	q := `SELECT mainflux_thing, mainflux_key, external_id, external_key, content, state FROM configs WHERE mainflux_thing = $1 AND owner = $2`
	cfg := bootstrap.Config{MFThing: id, Owner: key, MFChannels: []bootstrap.Channel{}}
	var content sql.NullString

	err := cr.db.QueryRow(q, id, key).
		Scan(&cfg.MFThing, &cfg.MFKey, &cfg.ExternalID, &cfg.ExternalKey, &content, &cfg.State)
	if err != nil {
		empty := bootstrap.Config{}
		if err == sql.ErrNoRows {
			return empty, bootstrap.ErrNotFound
		}
		return empty, err
	}

	q = `SELECT channel_id, name, metadata FROM channels WHERE config_id = $1 AND owner = $2`
	rows, err := cr.db.Query(q, id, key)
	if err != nil {
		cr.log.Error(fmt.Sprintf("Failed to retrieve connected due to %s", err))
		return bootstrap.Config{}, err
	}
	defer rows.Close()

	for rows.Next() {
		c := bootstrap.Channel{}
		if err = rows.Scan(&c.ID, &c.Name, &c.Metadata); err != nil {
			cr.log.Error(fmt.Sprintf("Failed to read connected channels due to %s", err))
			return bootstrap.Config{}, err
		}
		cfg.MFChannels = append(cfg.MFChannels, c)
	}

	cfg.Content = content.String
	return cfg, nil
}

func (cr configRepository) RetrieveAll(key string, filter bootstrap.Filter, offset, limit uint64) []bootstrap.Config {
	rows, err := cr.retrieveAll(key, filter, offset, limit)
	if err != nil {
		cr.log.Error(fmt.Sprintf("Failed to retrieve configs due to %s", err))
		return []bootstrap.Config{}
	}
	defer rows.Close()

	configs := map[string]bootstrap.Config{}
	var content sql.NullString
	ids := []string{}
	for rows.Next() {
		c := bootstrap.Config{Owner: key}
		if err = rows.Scan(&c.MFThing, &c.MFKey, &c.ExternalID, &c.ExternalKey, &content, &c.State); err != nil {
			cr.log.Error(fmt.Sprintf("Failed to read retrieved config due to %s", err))
			return []bootstrap.Config{}
		}

		c.Content = content.String
		configs[c.MFThing] = c
		ids = append(ids, c.MFThing)
	}
	fmt.Println("IDS: ", ids)
	q := `SELECT channel_id, name, metadata, config_id FROM channels WHERE config_id = ANY($1);`
	fmt.Println("ARR: ", pq.Array(&ids))
	rows, err = cr.db.Query(q, pq.Array(ids))
	if err != nil {
		cr.log.Error(fmt.Sprintf("Failed to retrieve connected due to %s", err))
		return []bootstrap.Config{}
	}
	defer rows.Close()

	for rows.Next() {
		var configID string
		c := bootstrap.Channel{}
		if err = rows.Scan(&c.ID, &c.Name, &c.Metadata, &configID); err != nil {
			cr.log.Error(fmt.Sprintf("Failed to read connected channels due to %s", err))
			return []bootstrap.Config{}
		}
		if cfg, ok := configs[configID]; ok {
			cfg.MFChannels = append(cfg.MFChannels, c)
			configs[configID] = cfg
		}
	}

	items := []bootstrap.Config{}
	for _, v := range configs {
		items = append(items, v)
	}

	return items
}

func (cr configRepository) RetrieveByExternalID(externalKey, externalID string) (bootstrap.Config, error) {
	q := `SELECT mainflux_thing, owner, mainflux_key, content, state FROM configs WHERE external_key = $1 AND external_id = $2`
	var content sql.NullString
	cfg := bootstrap.Config{
		ExternalID:  externalID,
		ExternalKey: externalKey,
	}

	if err := cr.db.QueryRow(q, externalKey, externalID).
		Scan(&cfg.MFThing, &cfg.Owner, &cfg.MFKey, &content, &cfg.State); err != nil {
		empty := bootstrap.Config{}
		if err == sql.ErrNoRows {
			return empty, bootstrap.ErrNotFound
		}
		return empty, err
	}

	q = `SELECT channel_id, name, metadata FROM channels WHERE config_id = $1`
	rows, err := cr.db.Query(q, cfg.MFThing)
	if err != nil {
		cr.log.Error(fmt.Sprintf("Failed to retrieve connected due to %s", err))
		return bootstrap.Config{}, err
	}
	defer rows.Close()

	for rows.Next() {
		c := bootstrap.Channel{}
		if err = rows.Scan(&c.ID, &c.Name, &c.Metadata); err != nil {
			cr.log.Error(fmt.Sprintf("Failed to read connected channels due to %s", err))
			return bootstrap.Config{}, err
		}
		cfg.MFChannels = append(cfg.MFChannels, c)
	}

	cfg.Content = content.String

	return cfg, nil
}

func (cr configRepository) Update(cfg bootstrap.Config) error {
	q := `UPDATE configs SET mainflux_channels = $1, content = $2, state = $3 WHERE mainflux_thing = $4 AND owner = $5`
	arr := pq.Array(cfg.MFChannels)
	res, err := cr.db.Exec(q, arr, cfg.Content, cfg.State, cfg.MFThing, cfg.Owner)
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

func (cr configRepository) Remove(key, id string) error {
	q := `DELETE FROM configs WHERE mainflux_thing = $1 AND owner = $2`
	cr.db.Exec(q, id, key)
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

func (cr configRepository) RetrieveUnknown(offset, limit uint64) []bootstrap.Config {
	q := `SELECT external_id, external_key FROM unknown_configs LIMIT $1 OFFSET $2`
	rows, err := cr.db.Query(q, limit, offset)
	if err != nil {
		cr.log.Error(fmt.Sprintf("Failed to retrieve config due to %s", err))
		return []bootstrap.Config{}
	}
	defer rows.Close()

	items := []bootstrap.Config{}
	for rows.Next() {
		c := bootstrap.Config{}
		if err = rows.Scan(&c.ExternalID, &c.ExternalKey); err != nil {
			cr.log.Error(fmt.Sprintf("Failed to read retrieved config due to %s", err))
			return []bootstrap.Config{}
		}

		items = append(items, c)
	}

	return items
}

func (cr configRepository) RemoveUnknown(key, id string) error {
	q := `DELETE FROM unknown_configs WHERE external_id = $1 AND external_key = $2`
	_, err := cr.db.Exec(q, id, key)
	return err
}

func (cr configRepository) retrieveAll(key string, filter bootstrap.Filter, offset, limit uint64) (*sql.Rows, error) {
	template := `SELECT mainflux_thing, mainflux_key, external_id, external_key, content, state FROM configs WHERE owner = $1 %s ORDER BY mainflux_thing LIMIT $2 OFFSET $3`
	params := []interface{}{key, limit, offset}
	// One empty string so that strings Join works if only one filter is applied.
	queries := []string{""}
	// Since key = 1, limit = 2, offset = 3, the next one is 4.
	counter := 4
	for k, v := range filter {
		queries = append(queries, fmt.Sprintf("%s = $%d", k, counter))
		params = append(params, v)
		counter++
	}

	f := strings.Join(queries, " AND ")
	return cr.db.Query(fmt.Sprintf(template, f), params...)
}

func generateValues(query string, cfg bootstrap.Config) (string, []interface{}) {
	template := `($1, $2, $%d, $%d, $%d)`
	content := nullString(cfg.Content)
	q := []string{}
	v := []interface{}{cfg.MFThing, cfg.Owner, cfg.MFKey, cfg.ExternalID, cfg.ExternalKey, content, cfg.State}

	counter := configFieldsNum
	for _, c := range cfg.MFChannels {
		q = append(q, fmt.Sprintf(template, counter, counter+1, counter+2))
		v = append(v, c.ID, c.Name, c.Metadata)
		counter += channelFieldsNum
	}
	if len(v) > configFieldsNum {
		chInsert := fmt.Sprintf(`) INSERT INTO channels (config_id, owner, channel_id, name, metadata) VALUES %s;`, strings.Join(q, ", "))
		query = fmt.Sprintf("%s%s%s", "WITH save_thing AS (", query, chInsert)
	}

	return query, v
}
