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

func (tr thingRepository) Save(thing bootstrap.Thing) (string, error) {
	const q = `INSERT INTO things (key, owner, mainflux_id, external_id, channel_id, status) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	if err := tr.db.QueryRow(q, thing.MFKey, thing.Owner, thing.MFThing, thing.ExternalID, thing.MFChan, thing.Status).Scan(&thing.ID); err != nil {
		return "", err
	}

	return thing.ID, nil
}

func (tr thingRepository) RetrieveByID(id, owner string) (bootstrap.Thing, error) {
	const q = `SELECT key, mainflux_id, external_id, channel_id, status FROM things WHERE id = $1 AND owner = $2`
	thing := bootstrap.Thing{ID: id, Owner: owner}
	err := tr.db.
		QueryRow(q, id, owner).
		Scan(&thing.MFKey, &thing.MFThing, &thing.ExternalID, &thing.MFChan, &thing.Status)

	if err != nil {
		empty := bootstrap.Thing{}
		if err == sql.ErrNoRows {
			return empty, bootstrap.ErrNotFound
		}
		return empty, err
	}

	return thing, nil
}

func (tr thingRepository) RetrieveByExternalID(externalID string) (bootstrap.Thing, error) {
	const q = `SELECT id, owner, key, mainflux_id, channel_id, status FROM things WHERE external_id = $1`
	thing := bootstrap.Thing{ExternalID: externalID}
	if err := tr.db.QueryRow(q, externalID).Scan(&thing.ID, &thing.Owner, &thing.MFKey, &thing.MFThing, &thing.MFChan, &thing.Status); err != nil {
		empty := bootstrap.Thing{}
		if err == sql.ErrNoRows {
			return empty, bootstrap.ErrNotFound
		}
		return empty, err
	}

	return thing, nil
}

func (tr thingRepository) Update(thing bootstrap.Thing) error {
	const q = `UPDATE things SET key = $1, mainflux_id = $2, external_id = $3, channel_id = $4, status = $5 WHERE id = $6 AND owner = $7`
	fmt.Printf("%v\n", thing)
	res, err := tr.db.Exec(q, thing.MFKey, thing.MFThing, thing.ExternalID, thing.MFChan, thing.Status, thing.ID, thing.Owner)
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
	const q = `UPDATE things SET status = $1 WHERE id = $2 AND owner = $3;`

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
