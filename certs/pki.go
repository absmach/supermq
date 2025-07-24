// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package certs

// Agent represents the PKI interface that all PKI implementations must satisfy.
type Agent interface {
	Issue(entityId, ttl string, ipAddrs []string) (Cert, error)
	View(serialNumber string) (Cert, error)
	Revoke(serialNumber string) error
	ListCerts(pm PageMetadata) (CertPage, error)
}
