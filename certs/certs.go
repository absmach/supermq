// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package certs

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"os"
	"time"

	"github.com/absmach/supermq/pkg/errors"
)

type Cert struct {
	SerialNumber string    `json:"serial_number"`
	CAChain      []string  `json:"ca_chain,omitempty"`
	IssuingCA    string    `json:"issuing_ca,omitempty"`
	Certificate  string    `json:"certificate,omitempty"`
	Key          string    `json:"key,omitempty"`
	ExpiryTime   time.Time `json:"expiry_time"`
	ClientID     string    `json:"entity_id"`
	Revoked      bool      `json:"revoked"`
}

type CertPage struct {
	Total        uint64 `json:"total"`
	Offset       uint64 `json:"offset"`
	Limit        uint64 `json:"limit"`
	Certificates []Cert `json:"certificates,omitempty"`
}

// Repository specifies a Config persistence API.
type Repository interface {
	// Save saves cert for client into database
	Save(ctx context.Context, cert Cert) (string, error)

	// Update updates an existing certificate in the database
	Update(ctx context.Context, cert Cert) error

	// RetrieveAll retrieve issued certificates
	RetrieveAll(ctx context.Context, offset, limit uint64) (CertPage, error)

	// Remove removes certificate from DB for a given client ID
	Remove(ctx context.Context, clientID string) error

	// RemoveBySerial removes certificate from DB for a given serial number
	RemoveBySerial(ctx context.Context, serialID string) error

	// RetrieveByClient retrieves issued certificates for a given client ID
	RetrieveByClient(ctx context.Context, clientID string, pm PageMetadata) (CertPage, error)

	// RetrieveBySerial retrieves a certificate for a given serial ID
	RetrieveBySerial(ctx context.Context, serialID string) (Cert, error)
}

type PageMetadata struct {
	Total      uint64 `json:"total,omitempty"`
	Offset     uint64 `json:"offset,omitempty"`
	Limit      uint64 `json:"limit,omitempty"`
	CommonName string `json:"common_name,omitempty"`
	Revoked    string `json:"revoked,omitempty"`
}

var ErrMissingCerts = errors.New("CA path or CA key path not set")

func LoadCertificates(caPath, caKeyPath string) (tls.Certificate, *x509.Certificate, error) {
	if caPath == "" || caKeyPath == "" {
		return tls.Certificate{}, &x509.Certificate{}, ErrMissingCerts
	}

	_, err := os.Stat(caPath)
	if os.IsNotExist(err) || os.IsPermission(err) {
		return tls.Certificate{}, &x509.Certificate{}, err
	}

	_, err = os.Stat(caKeyPath)
	if os.IsNotExist(err) || os.IsPermission(err) {
		return tls.Certificate{}, &x509.Certificate{}, err
	}

	tlsCert, err := tls.LoadX509KeyPair(caPath, caKeyPath)
	if err != nil {
		return tlsCert, &x509.Certificate{}, err
	}

	b, err := os.ReadFile(caPath)
	if err != nil {
		return tlsCert, &x509.Certificate{}, err
	}

	caCert, err := ReadCert(b)
	if err != nil {
		return tlsCert, &x509.Certificate{}, err
	}

	return tlsCert, caCert, nil
}

func ReadCert(b []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, errors.New("failed to decode PEM data")
	}

	return x509.ParseCertificate(block.Bytes)
}
