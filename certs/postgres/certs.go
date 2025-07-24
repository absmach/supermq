// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/absmach/supermq/certs"
	"github.com/absmach/supermq/pkg/errors"
	repoerr "github.com/absmach/supermq/pkg/errors/repository"
	"github.com/absmach/supermq/pkg/postgres"
)

var _ certs.Repository = (*certsRepository)(nil)

type PageMetadata struct {
	Offset   uint64 `db:"offset,omitempty"`
	Limit    uint64 `db:"limit,omitempty"`
	ClientID string `db:"client_id,omitempty"`
}

type certsRepository struct {
	db postgres.Database
}

// NewRepository instantiates a PostgreSQL implementation of certs
// repository.
func NewRepository(db postgres.Database) certs.Repository {
	return &certsRepository{db: db}
}

func (cr certsRepository) RetrieveAll(ctx context.Context, offset, limit uint64) (certs.CertPage, error) {
	pm := certs.PageMetadata{
		Offset: offset,
		Limit:  limit,
	}

	return cr.retrieveCertificates(ctx, "", pm)
}

func (cr certsRepository) RetrieveByClient(ctx context.Context, clientID string, pm certs.PageMetadata) (certs.CertPage, error) {
	return cr.retrieveCertificates(ctx, clientID, pm)
}

func (cr certsRepository) Save(ctx context.Context, cert certs.Cert) (string, error) {
	dbcrt := toDBCert(cert)

	q := `INSERT INTO certs (client_id, serial_number, expiry_time, revoked) 
	VALUES (:client_id, :serial_number, :expiry_time, :revoked)
	RETURNING serial_number`

	row, err := cr.db.NamedQueryContext(ctx, q, dbcrt)
	if err != nil {
		return "", postgres.HandleError(repoerr.ErrCreateEntity, err)
	}
	defer row.Close()

	var serialNumber string
	if row.Next() {
		if err := row.Scan(&serialNumber); err != nil {
			return "", errors.Wrap(repoerr.ErrFailedOpDB, err)
		}
	}

	return serialNumber, nil
}

func (cr certsRepository) Update(ctx context.Context, cert certs.Cert) error {
	dbcrt := toDBCert(cert)

	q := `UPDATE certs SET 
		client_id = :client_id,
		expiry_time = :expiry_time,
		revoked = :revoked
		WHERE serial_number = :serial_number`

	result, err := cr.db.NamedExecContext(ctx, q, dbcrt)
	if err != nil {
		return postgres.HandleError(repoerr.ErrUpdateEntity, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(repoerr.ErrFailedOpDB, err)
	}

	if rowsAffected == 0 {
		return errors.Wrap(repoerr.ErrNotFound, errors.New("certificate not found"))
	}

	return nil
}

func (cr certsRepository) Remove(ctx context.Context, clientID string) error {
	q := `DELETE FROM certs WHERE client_id = :client_id`
	var c certs.Cert
	c.ClientID = clientID
	dbcrt := toDBCert(c)
	if _, err := cr.db.NamedExecContext(ctx, q, dbcrt); err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}
	return nil
}

func (cr certsRepository) RemoveBySerial(ctx context.Context, serialID string) error {
	q := `DELETE FROM certs WHERE serial_number = :serial_number`
	var c certs.Cert
	c.SerialNumber = serialID
	dbcrt := toDBCert(c)
	if _, err := cr.db.NamedExecContext(ctx, q, dbcrt); err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}
	return nil
}

func PageQuery(pm certs.PageMetadata) (string, error) {
	var query []string

	if pm.Revoked != "all" {
		switch pm.Revoked {
		case "true":
			query = append(query, "revoked = true")
		case "false":
			query = append(query, "revoked = false")
		}
	}

	if pm.CommonName != "" {
		query = append(query, "client_id ILIKE '%' || :client_id || '%'")
	}

	var emq string
	if len(query) > 0 {
		emq = fmt.Sprintf("WHERE %s", strings.Join(query, " AND "))
	}
	return emq, nil
}

func (cr certsRepository) retrieveCertificates(ctx context.Context, clientID string, pm certs.PageMetadata) (certs.CertPage, error) {
	pageQuery, err := PageQuery(pm)
	if err != nil {
		return certs.CertPage{}, err
	}

	q := fmt.Sprintf(`SELECT client_id, serial_number, expiry_time, revoked FROM certs %s`,
		pageQuery)

	q = applyLimitOffset(q)

	param := PageMetadata{
		Offset:   pm.Offset,
		Limit:    pm.Limit,
		ClientID: clientID,
	}

	rows, err := cr.db.NamedQueryContext(ctx, q, param)
	if err != nil {
		return certs.CertPage{}, err
	}
	defer rows.Close()

	certificates := []certs.Cert{}
	for rows.Next() {
		c := certs.Cert{}
		if err := rows.Scan(&c.ClientID, &c.SerialNumber, &c.ExpiryTime, &c.Revoked); err != nil {
			return certs.CertPage{}, err
		}
		certificates = append(certificates, c)
	}

	cq := fmt.Sprintf(`SELECT COUNT(*) AS total_count
			FROM certs %s`, pageQuery)

	total, err := postgres.Total(ctx, cr.db, cq, param)
	if err != nil {
		return certs.CertPage{}, errors.Wrap(repoerr.ErrFailedOpDB, err)
	}

	return certs.CertPage{
		Total:        total,
		Limit:        pm.Limit,
		Offset:       pm.Offset,
		Certificates: certificates,
	}, nil
}

func (cr certsRepository) RetrieveBySerial(ctx context.Context, serial string) (certs.Cert, error) {
	q := `SELECT client_id, serial_number, expiry_time, revoked FROM certs WHERE serial_number = $1`
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

type dbCert struct {
	ClientID     string    `db:"client_id"`
	SerialNumber string    `db:"serial_number"`
	ExpiryTime   time.Time `db:"expiry_time"`
	Revoked      bool      `db:"revoked"`
}

func toDBCert(c certs.Cert) dbCert {
	return dbCert{
		ClientID:     c.ClientID,
		SerialNumber: c.SerialNumber,
		ExpiryTime:   c.ExpiryTime,
		Revoked:      c.Revoked,
	}
}

func toCert(cdb dbCert) certs.Cert {
	var c certs.Cert
	c.ClientID = cdb.ClientID
	c.SerialNumber = cdb.SerialNumber
	c.ExpiryTime = cdb.ExpiryTime
	c.Revoked = cdb.Revoked
	return c
}

func applyLimitOffset(query string) string {
	return fmt.Sprintf(`%s
			LIMIT :limit OFFSET :offset`, query)
}
