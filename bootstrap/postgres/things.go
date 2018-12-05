package postgres

import (
	"database/sql"
	"fmt"
	"nov/bootstrap"

	"github.com/lib/pq" // required for DB access
	"github.com/mainflux/mainflux/logger"
)

var _ bootstrap.ThingRepository = (*thingRepository)(nil)

type thingRepository struct {
	db  *sql.DB
	log logger.Logger
}

// NewThingRepository instantiates a PostgreSQL implementation of thing
// repository.
func NewThingRepository(db *sql.DB, log logger.Logger) bootstrap.ThingRepository {
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

func (tr thingRepository) Save(thing bootstrap.Thing) (string, error) {
	q := `INSERT INTO things (mainflux_key, owner, mainflux_thing, external_id, mainflux_channels, external_config, status)
	VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	if err := tr.db.QueryRow(q, nullString(thing.MFKey), thing.Owner, nullString(thing.MFThing),
		thing.ExternalID, pq.Array(thing.MFChannels), nullString(thing.Config), thing.Status).Scan(&thing.ID); err != nil {
		return "", err
	}

	return thing.ID, nil
}

func (tr thingRepository) RetrieveByID(key, id string) (bootstrap.Thing, error) {
	q := `SELECT mainflux_key, mainflux_thing, external_id, mainflux_channels, external_config, status FROM things WHERE id = $1 AND owner = $2`
	thing := bootstrap.Thing{ID: id, Owner: key}
	var mfKey, mfThing, config sql.NullString

	err := tr.db.
		QueryRow(q, id, key).
		Scan(&mfKey, &mfThing, &thing.ExternalID, pq.Array(&thing.MFChannels), &config, &thing.Status)

	thing.MFKey = mfKey.String
	thing.MFThing = mfThing.String
	thing.Config = config.String

	if err != nil {
		empty := bootstrap.Thing{}
		if err == sql.ErrNoRows {
			return empty, bootstrap.ErrNotFound
		}
		return empty, err
	}

	return thing, nil
}

func (tr thingRepository) RetrieveAll(key string, offset, limit uint64) []bootstrap.Thing {
	q := `SELECT mainflux_key, mainflux_thing, external_id, mainflux_channels, external_config, status FROM things WHERE owner = $1 ORDER BY id LIMIT $2 OFFSET $3`
	items := []bootstrap.Thing{}

	rows, err := tr.db.Query(q, key, limit, offset)
	if err != nil {
		tr.log.Error(fmt.Sprintf("Failed to retrieve things due to %s", err))
		return []bootstrap.Thing{}
	}
	defer rows.Close()

	var mfKey, mfThing, config sql.NullString
	for rows.Next() {
		t := bootstrap.Thing{Owner: key}
		if err = rows.Scan(&mfKey, &mfThing, &t.ExternalID, pq.Array(&t.MFChannels), &config, &t.Status); err != nil {
			tr.log.Error(fmt.Sprintf("Failed to read retrieved thing due to %s", err))

			return []bootstrap.Thing{}
		}

		t.MFKey = mfKey.String
		t.MFThing = mfThing.String
		t.Config = config.String

		items = append(items, t)
	}

	return items
}

func (tr thingRepository) RetrieveByExternalID(externalID string) (bootstrap.Thing, error) {
	q := `SELECT id, owner, mainflux_key, mainflux_thing, mainflux_channels, external_config, status FROM things WHERE external_id = $1`

	var mfKey, mfThing, config sql.NullString
	thing := bootstrap.Thing{ExternalID: externalID}
	if err := tr.db.QueryRow(q, externalID).Scan(&thing.ID, &thing.Owner, &mfKey, &mfThing, pq.Array(&thing.MFChannels), &config, &thing.Status); err != nil {
		empty := bootstrap.Thing{}
		if err == sql.ErrNoRows {
			return empty, bootstrap.ErrNotFound
		}
		return empty, err
	}

	thing.MFKey = mfKey.String
	thing.MFThing = mfThing.String
	thing.Config = config.String

	return thing, nil
}

func (tr thingRepository) Update(thing bootstrap.Thing) error {
	q := `UPDATE things SET mainflux_key = $1, mainflux_thing = $2, external_id = $3, mainflux_channels = $4, external_config = $5, status = $6 WHERE id = $7 AND owner = $8`
	res, err := tr.db.Exec(q, nullString(thing.MFKey), nullString(thing.MFThing), thing.ExternalID, pq.Array(thing.MFChannels), thing.Config, thing.Status, thing.ID, thing.Owner)
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

func (tr thingRepository) ChangeStatus(key, id string, status bootstrap.Status) error {
	q := `UPDATE things SET status = $1 WHERE id = $2 AND owner = $3;`

	println("calling")
	res, err := tr.db.Exec(q, status, id, key)
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
