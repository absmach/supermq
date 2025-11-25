// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package notifications

import (
	"context"
)

// User represents minimal user information needed for notifications.
type User struct {
	ID       string
	Email    string
	Username string
}

// Domain represents minimal domain information needed for notifications.
type Domain struct {
	ID   string
	Name string
}

// UserRepository provides access to user information.
type UserRepository interface {
	// RetrieveByID retrieves a user by ID.
	RetrieveByID(ctx context.Context, id string) (User, error)
}

// DomainRepository provides access to domain information.
type DomainRepository interface {
	// RetrieveByID retrieves a domain by ID.
	RetrieveByID(ctx context.Context, id string) (Domain, error)
}
