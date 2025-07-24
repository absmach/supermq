// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package certs

import (
	"context"
	"time"

	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	mgsdk "github.com/absmach/supermq/pkg/sdk"
)

var (
	// ErrFailedCertCreation failed to create certificate.
	ErrFailedCertCreation = errors.New("failed to create client certificate")

	// ErrFailedCertRevocation failed to revoke certificate.
	ErrFailedCertRevocation = errors.New("failed to revoke certificate")

	ErrFailedToRemoveCertFromDB = errors.New("failed to remove cert serial from db")

	ErrFailedReadFromPKI = errors.New("failed to read certificate from PKI")

	ErrFailedReadFromDB = errors.New("failed to read certificate from database")
)

var _ Service = (*certsService)(nil)

// Service specifies an API that must be fulfilled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	// IssueCert issues certificate for given client id if access is granted with token
	IssueCert(ctx context.Context, domainID, token, clientID, ttl string) (Cert, error)

	// ListCerts lists certificates issued for a given client ID
	ListCerts(ctx context.Context, clientID string, pm PageMetadata) (CertPage, error)

	// ListSerials lists certificate serial IDs issued for a given client ID
	ListSerials(ctx context.Context, clientID string, pm PageMetadata) (CertPage, error)

	// ViewCert retrieves the certificate issued for a given serial ID
	ViewCert(ctx context.Context, serialID string) (Cert, error)

	// RevokeCert revokes a certificate for a given client ID
	RevokeCert(ctx context.Context, domainID, token, clientID string) (Revoke, error)

	// RevokeBySerial revokes a certificate by its serial number from both PKI and database
	RevokeBySerial(ctx context.Context, serialID string) (Revoke, error)
}

// Revoke defines the conditions to revoke a certificate.
type Revoke struct {
	RevocationTime time.Time `json:"revocation_time"`
}
type certsService struct {
	sdk       mgsdk.SDK
	certsRepo Repository
	pki       Agent
}

// New returns new Certs service.
func New(sdk mgsdk.SDK, certsRepo Repository, pkiAgent Agent) Service {
	return &certsService{
		sdk:       sdk,
		pki:       pkiAgent,
		certsRepo: certsRepo,
	}
}

func (cs *certsService) IssueCert(ctx context.Context, domainID, token, clientID, ttl string) (Cert, error) {
	var err error

	client, err := cs.sdk.Client(ctx, clientID, domainID, token)
	if err != nil {
		return Cert{}, errors.Wrap(ErrFailedCertCreation, err)
	}

	cert, err := cs.pki.Issue(client.ID, ttl, []string{})
	if err != nil {
		return Cert{}, errors.Wrap(ErrFailedCertCreation, err)
	}

	_, err = cs.certsRepo.Save(ctx, cert)
	if err != nil {
		return Cert{}, errors.Wrap(ErrFailedCertCreation, err)
	}

	return Cert{
		SerialNumber: cert.SerialNumber,
		Certificate:  cert.Certificate,
		Key:          cert.Key,
		ExpiryTime:   cert.ExpiryTime,
		IssuingCA:    cert.IssuingCA,
		CAChain:      cert.CAChain,
		ClientID:     cert.ClientID,
		Revoked:      cert.Revoked,
	}, err
}

func (cs *certsService) RevokeCert(ctx context.Context, domainID, token, clientID string) (Revoke, error) {
	var revoke Revoke
	var err error

	cp, err := cs.certsRepo.RetrieveByClient(ctx, clientID, PageMetadata{Offset: 0, Limit: 10000})
	if err != nil {
		return revoke, errors.Wrap(ErrFailedCertRevocation, err)
	}

	for _, c := range cp.Certificates {
		err := cs.pki.Revoke(c.SerialNumber)
		if err != nil {
			return revoke, errors.Wrap(ErrFailedCertRevocation, err)
		}

		c.Revoked = true
		err = cs.certsRepo.Update(ctx, c)
		if err != nil {
			return revoke, errors.Wrap(ErrFailedReadFromDB, err)
		}

		revoke.RevocationTime = time.Now().UTC()
	}

	return revoke, nil
}

func (cs *certsService) RevokeBySerial(ctx context.Context, serialID string) (Revoke, error) {
	var revoke Revoke

	cert, err := cs.certsRepo.RetrieveBySerial(ctx, serialID)
	if err != nil {
		return revoke, errors.Wrap(ErrFailedReadFromDB, err)
	}

	err = cs.pki.Revoke(serialID)
	if err != nil {
		return revoke, errors.Wrap(ErrFailedCertRevocation, err)
	}

	cert.Revoked = true
	err = cs.certsRepo.Update(ctx, cert)
	if err != nil {
		return revoke, errors.Wrap(ErrFailedReadFromDB, err)
	}

	revoke.RevocationTime = time.Now().UTC()
	return revoke, nil
}

func (cs *certsService) ListCerts(ctx context.Context, clientID string, pm PageMetadata) (CertPage, error) {
	cp, err := cs.certsRepo.RetrieveByClient(ctx, clientID, pm)
	if err != nil {
		return CertPage{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}

	for i, cert := range cp.Certificates {
		vcert, err := cs.pki.View(cert.SerialNumber)
		if err != nil {
			return CertPage{}, errors.Wrap(svcerr.ErrViewEntity, err)
		}
		cp.Certificates[i].Certificate = vcert.Certificate
		cp.Certificates[i].Key = vcert.Key
	}

	return cp, nil
}

func (cs *certsService) ListSerials(ctx context.Context, clientID string, pm PageMetadata) (CertPage, error) {
	cp, err := cs.certsRepo.RetrieveByClient(ctx, clientID, pm)
	if err != nil {
		return CertPage{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}

	return cp, nil
}

func (cs *certsService) ViewCert(ctx context.Context, serialID string) (Cert, error) {
	cert, err := cs.certsRepo.RetrieveBySerial(ctx, serialID)
	if err != nil {
		return Cert{}, errors.Wrap(ErrFailedReadFromDB, err)
	}

	vcert, err := cs.pki.View(serialID)
	if err != nil {
		return Cert{}, errors.Wrap(ErrFailedReadFromPKI, err)
	}

	return Cert{
		SerialNumber: cert.SerialNumber,
		Certificate:  vcert.Certificate,
		Key:          vcert.Key,
		ExpiryTime:   vcert.ExpiryTime,
		ClientID:     cert.ClientID,
		Revoked:      cert.Revoked,
	}, nil
}
