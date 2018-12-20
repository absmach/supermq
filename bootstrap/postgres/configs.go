package postgres

import (
	"database/sql"
	"fmt"
	"nov/bootstrap"
	"strings"

	"github.com/lib/pq" // required for DB access
	"github.com/mainflux/mainflux/logger"
)

const duplicateErr = "unique_violation"

var _ bootstrap.ConfigRepository = (*thingRepository)(nil)

type thingRepository struct {
	db  *sql.DB
	log logger.Logger
}

// NewThingRepository instantiates a PostgreSQL implementation of thing
// repository.
func NewThingRepository(db *sql.DB, log logger.Logger) bootstrap.ConfigRepository {
	return &thingRepository{db: db, log: log}
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

func (tr thingRepository) Save(thing bootstrap.Config) (string, error) {
	q := `INSERT INTO things (mainflux_key, owner, mainflux_thing, external_id, external_key, mainflux_channels, content, state)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

	if err := tr.db.QueryRow(q, thing.MFKey, thing.Owner, thing.MFThing, thing.ExternalID, thing.ExternalKey,
		pq.Array(thing.MFChannels), nullString(thing.Content), thing.State).Scan(&thing.ID); err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code.Name() == duplicateErr {
			return "", bootstrap.ErrConflict
		}
		return "", err
	}

	return thing.ID, nil
}

func (tr thingRepository) RetrieveByID(key, id string) (bootstrap.Config, error) {
	q := `SELECT mainflux_key, mainflux_thing, external_id, external_key, mainflux_channels, content, state FROM things WHERE id = $1 AND owner = $2`
	config := bootstrap.Config{ID: id, Owner: key}
	var content sql.NullString

	err := tr.db.
		QueryRow(q, id, key).
		Scan(&config.MFKey, &config.MFThing, &config.ExternalID, &config.ExternalKey, pq.Array(&config.MFChannels), &content, &config.State)

	config.Content = content.String

	if err != nil {
		empty := bootstrap.Config{}
		if err == sql.ErrNoRows {
			return empty, bootstrap.ErrNotFound
		}
		return empty, err
	}

	return config, nil
}

func (tr thingRepository) RetrieveAll(key string, filter bootstrap.Filter, offset, limit uint64) []bootstrap.Config {
	rows, err := tr.retrieveAll(key, filter, offset, limit)
	if err != nil {
		tr.log.Error(fmt.Sprintf("Failed to retrieve things due to %s", err))
		return []bootstrap.Config{}
	}
	defer rows.Close()

	items := []bootstrap.Config{}
	var content sql.NullString
	for rows.Next() {
		c := bootstrap.Config{Owner: key}
		if err = rows.Scan(&c.ID, &c.MFKey, &c.MFThing, &c.ExternalID, &c.ExternalKey, pq.Array(&c.MFChannels), &content, &c.State); err != nil {
			tr.log.Error(fmt.Sprintf("Failed to read retrieved thing due to %s", err))
			return []bootstrap.Config{}
		}

		c.Content = content.String
		items = append(items, c)
	}

	return items
}

func (tr thingRepository) RetrieveByExternalID(externalKey, externalID string) (bootstrap.Config, error) {
	q := `SELECT id, owner, mainflux_key, mainflux_thing, mainflux_channels, content, state FROM things WHERE external_key = $1 AND external_id = $2`

	var content sql.NullString
	config := bootstrap.Config{
		ExternalID:  externalID,
		ExternalKey: externalKey,
	}

	if err := tr.db.QueryRow(q, externalKey, externalID).Scan(&config.ID, &config.Owner, &config.MFKey, &config.MFThing, pq.Array(&config.MFChannels), &content, &config.State); err != nil {
		empty := bootstrap.Config{}
		if err == sql.ErrNoRows {
			return empty, bootstrap.ErrNotFound
		}
		return empty, err
	}

	config.Content = content.String

	return config, nil
}

func (tr thingRepository) Update(config bootstrap.Config) error {
	q := `UPDATE things SET mainflux_channels = $1, content = $2, state = $3 WHERE id = $4 AND owner = $5`
	res, err := tr.db.Exec(q, pq.Array(config.MFChannels), config.Content, config.State, config.ID, config.Owner)
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

func (tr thingRepository) Remove(key, id string) error {
	q := `DELETE FROM things WHERE id = $1 AND owner = $2`
	tr.db.Exec(q, id, key)
	return nil
}

func (tr thingRepository) ChangeState(key, id string, state bootstrap.State) error {
	q := `UPDATE things SET state = $1 WHERE id = $2 AND owner = $3;`

	res, err := tr.db.Exec(q, state, id, key)
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

func (tr thingRepository) retrieveAll(key string, filter bootstrap.Filter, offset, limit uint64) (*sql.Rows, error) {
	template := `SELECT id, mainflux_key, mainflux_thing, external_id, external_key, mainflux_channels, content, state FROM things WHERE owner = $1 %s ORDER BY id LIMIT $2 OFFSET $3`
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
	return tr.db.Query(fmt.Sprintf(template, f), params...)
}
