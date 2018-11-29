package postgres

import (
	"database/sql"
	"fmt"

	"github.com/mainflux/mainflux/logger"

	"nov/bootstrap"

	_ "github.com/lib/pq" // required for DB access
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
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

func (tr thingRepository) Save(thing bootstrap.Thing) (string, error) {
	q := `INSERT INTO things (mainflux_key, owner, mainflux_thing, external_id, mainflux_channel, status) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	if err := tr.db.QueryRow(q, nullString(thing.MFKey), thing.Owner, nullString(thing.MFThing),
		thing.ExternalID, nullString(thing.MFChan), thing.Status).Scan(&thing.ID); err != nil {
		return "", err
	}

	return thing.ID, nil
}

func (tr thingRepository) RetrieveByID(id, owner string) (bootstrap.Thing, error) {
	q := `SELECT mainflux_key, mainflux_thing, external_id, mainflux_channel, status FROM things WHERE id = $1 AND owner = $2`
	thing := bootstrap.Thing{ID: id, Owner: owner}
	var mfKey, mfThing, mfChan sql.NullString

	err := tr.db.
		QueryRow(q, id, owner).
		Scan(&mfKey, &mfThing, &thing.ExternalID, &mfChan, &thing.Status)

	thing.MFKey = mfKey.String
	thing.MFThing = mfThing.String
	thing.MFChan = mfChan.String

	if err != nil {
		empty := bootstrap.Thing{}
		if err == sql.ErrNoRows {
			return empty, bootstrap.ErrNotFound
		}
		return empty, err
	}

	return thing, nil
}

func (tr thingRepository) RetrieveAll(owner string, offset, limit uint64) []bootstrap.Thing {
	q := `SELECT mainflux_key, mainflux_thing, external_id, mainflux_channel, status FROM things WHERE owner = $1 ORDER BY id LIMIT $2 OFFSET $3`
	items := []bootstrap.Thing{}

	rows, err := tr.db.Query(q, owner, limit, offset)
	if err != nil {
		tr.log.Error(fmt.Sprintf("Failed to retrieve things due to %s", err))
		return []bootstrap.Thing{}
	}
	defer rows.Close()

	var mfKey, mfThing, mfChan sql.NullString
	for rows.Next() {
		t := bootstrap.Thing{Owner: owner}
		if err = rows.Scan(&mfKey, &mfThing, &t.ExternalID, &mfChan, &t.Status); err != nil {
			tr.log.Error(fmt.Sprintf("Failed to read retrieved thing due to %s", err))
			return []bootstrap.Thing{}
		}
		t.MFKey = mfKey.String
		t.MFThing = mfThing.String
		t.MFChan = mfChan.String
		items = append(items, t)
	}

	return items
}

func (tr thingRepository) RetrieveByExternalID(externalID string) (bootstrap.Thing, error) {
	q := `SELECT id, owner, mainflux_key, mainflux_thing, mainflux_channel, status FROM things WHERE external_id = $1`

	var mfKey, mfThing, mfChan sql.NullString
	thing := bootstrap.Thing{ExternalID: externalID}
	if err := tr.db.QueryRow(q, externalID).Scan(&thing.ID, &thing.Owner, &mfKey, &mfThing, &mfChan, &thing.Status); err != nil {
		empty := bootstrap.Thing{}
		if err == sql.ErrNoRows {
			return empty, bootstrap.ErrNotFound
		}
		return empty, err
	}

	thing.MFKey = mfKey.String
	thing.MFThing = mfThing.String
	thing.MFChan = mfChan.String
	return thing, nil
}

func (tr thingRepository) Update(thing bootstrap.Thing) error {
	q := `UPDATE things SET mainflux_key = $1, mainflux_thing = $2, external_id = $3, mainflux_channel = $4, status = $5 WHERE id = $6 AND owner = $7`
	fmt.Printf("%v\n", thing)
	res, err := tr.db.Exec(q, nullString(thing.MFKey), nullString(thing.MFThing), thing.ExternalID, nullString(thing.MFChan), thing.Status, thing.ID, thing.Owner)
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

func (tr thingRepository) Remove(id, owner string) error {
	q := `DELETE FROM things WHERE id = $1 AND owner = $2`
	tr.db.Exec(q, id, owner)
	return nil
}

func (tr thingRepository) ChangeStatus(id, owner string, status bootstrap.Status) error {
	q := `UPDATE things SET status = $1 WHERE id = $2 AND owner = $3;`

	res, err := tr.db.Exec(q, status, id, owner)
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
