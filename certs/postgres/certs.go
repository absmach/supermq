// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/absmach/supermq/certs"
	"github.com/absmach/supermq/pkg/errors"
	repoerr "github.com/absmach/supermq/pkg/errors/repository"
	"github.com/absmach/supermq/pkg/postgres"
	"github.com/jmoiron/sqlx"
)

var _ certs.Repository = (*certsRepository)(nil)

type certsRepository struct {
	db  postgres.Database
	log *slog.Logger
}

// NewRepository instantiates a PostgreSQL implementation of certs
// repository.
func NewRepository(db postgres.Database, log *slog.Logger) certs.Repository {
	return &certsRepository{db: db, log: log}
}

func (cr certsRepository) RetrieveAll(ctx context.Context, offset, limit uint64) (certs.CertPage, error) {
	q := `SELECT client_id, serial_number, expiry_time FROM certs ORDER BY expiry_time LIMIT $1 OFFSET $2;`
	rows, err := cr.db.QueryContext(ctx, q, limit, offset)
	if err != nil {
		cr.log.Error(fmt.Sprintf("Failed to retrieve configs due to %s", err))
		return certs.CertPage{}, err
	}
	defer rows.Close()

	certificates := []certs.Cert{}
	for rows.Next() {
		c := certs.Cert{}
		if err := rows.Scan(&c.ClientID, &c.SerialNumber, &c.ExpiryTime); err != nil {
			cr.log.Error(fmt.Sprintf("Failed to read retrieved config due to %s", err))
			return certs.CertPage{}, err
		}
		certificates = append(certificates, c)
	}

	q = `SELECT COUNT(*) FROM certs`
	var total uint64
	if err := cr.db.QueryRowxContext(ctx, q).Scan(&total); err != nil {
		cr.log.Error(fmt.Sprintf("Failed to count certs due to %s", err))
		return certs.CertPage{}, err
	}

	return certs.CertPage{
		Total:        total,
		Limit:        limit,
		Offset:       offset,
		Certificates: certificates,
	}, nil
}

func (cr certsRepository) Save(ctx context.Context, cert certs.Cert) (string, error) {
	q := `INSERT INTO certs (client_id, serial_number, expiry_time) VALUES (:client_id, :serial_number, :expiry_time)`

	tx, err := cr.db.BeginTxx(ctx, nil)
	if err != nil {
		return "", errors.Wrap(repoerr.ErrCreateEntity, err)
	}

	dbcrt := toDBCert(cert)

	if _, err := tx.NamedExec(q, dbcrt); err != nil {
		cr.rollback("Failed to insert a Cert", tx, err)

		return "", errors.Wrap(repoerr.ErrCreateEntity, err)
	}

	if err := tx.Commit(); err != nil {
		cr.rollback("Failed to commit Config save", tx, err)
	}

	return cert.SerialNumber, nil
}

func (cr certsRepository) Remove(ctx context.Context, serial string) error {
	if _, err := cr.RetrieveBySerial(ctx, serial); err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}
	q := `DELETE FROM certs WHERE serial_number = :serial_number`
	var c certs.Cert
	c.SerialNumber = serial
	dbcrt := toDBCert(c)
	if _, err := cr.db.NamedExecContext(ctx, q, dbcrt); err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}
	return nil
}

func (cr certsRepository) RetrieveByClient(ctx context.Context, clientID string, offset, limit uint64) (certs.CertPage, error) {
	q := `SELECT client_id, serial_number, expiry_time FROM certs WHERE client_id = $1 ORDER BY expiry_time LIMIT $2 OFFSET $3;`
	rows, err := cr.db.QueryContext(ctx, q, clientID, limit, offset)
	if err != nil {
		cr.log.Error(fmt.Sprintf("Failed to retrieve configs due to %s", err))
		return certs.CertPage{}, err
	}
	defer rows.Close()

	certificates := []certs.Cert{}
	for rows.Next() {
		c := certs.Cert{}
		if err := rows.Scan(&c.ClientID, &c.SerialNumber, &c.ExpiryTime); err != nil {
			cr.log.Error(fmt.Sprintf("Failed to read retrieved config due to %s", err))
			return certs.CertPage{}, err
		}
		certificates = append(certificates, c)
	}

	q = `SELECT COUNT(*) FROM certs WHERE client_id = $1`
	var total uint64
	if err := cr.db.QueryRowxContext(ctx, q, clientID).Scan(&total); err != nil {
		cr.log.Error(fmt.Sprintf("Failed to count certs due to %s", err))
		return certs.CertPage{}, err
	}

	return certs.CertPage{
		Total:        total,
		Limit:        limit,
		Offset:       offset,
		Certificates: certificates,
	}, nil
}

func (cr certsRepository) RetrieveBySerial(ctx context.Context, serial string) (certs.Cert, error) {
	q := `SELECT client_id, serial_number, expiry_time FROM certs WHERE serial_number = $1`
	var dbcrt dbCert
	var c certs.Cert

	if err := cr.db.QueryRowxContext(ctx, q, serial).StructScan(&dbcrt); err != nil {
		if err == sql.ErrNoRows {
			return c, errors.Wrap(repoerr.ErrNotFound, err)
		}

		return c, errors.Wrap(repoerr.ErrViewEntity, err)
	}
	c = toCert(dbcrt)

	return c, nil
}

func (cr certsRepository) rollback(content string, tx *sqlx.Tx, err error) {
	cr.log.Error(fmt.Sprintf("%s %s", content, err))

	if err := tx.Rollback(); err != nil {
		cr.log.Error(fmt.Sprintf("Failed to rollback due to %s", err))
	}
}

type dbCert struct {
	ClientID     string    `db:"client_id"`
	SerialNumber string    `db:"serial_number"`
	ExpiryTime   time.Time `db:"expiry_time"`
}

func toDBCert(c certs.Cert) dbCert {
	return dbCert{
		ClientID:     c.ClientID,
		SerialNumber: c.SerialNumber,
		ExpiryTime:   c.ExpiryTime,
	}
}

func toCert(cdb dbCert) certs.Cert {
	var c certs.Cert
	c.ClientID = cdb.ClientID
	c.SerialNumber = cdb.SerialNumber
	c.ExpiryTime = cdb.ExpiryTime
	return c
}
