// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"encoding/json"
	"strings"

	apiutil "github.com/absmach/supermq/api/http/util"
)

// Role represents Client role.
type Role uint8

// Possible Client role values.
const (
	UserRole Role = iota
	AdminRole

	// AllRole is used for querying purposes to list clients irrespective
	// of their role - both admin and user. It is never stored in the
	// database as the actual Client role and should always be the largest
	// value in this enumeration.
	AllRole
)

// String representation of the possible role values.
const (
	Admin = "admin"
	User  = "user"
)

// String converts client role to string literal.
func (cs Role) String() string {
	switch cs {
	case AdminRole:
		return Admin
	case UserRole:
		return User
	case AllRole:
		return All
	default:
		return Unknown
	}
}

// ToRole converts string value to a valid Client role.
func ToRole(status string) (Role, error) {
	switch status {
	case "", User:
		return UserRole, nil
	case Admin:
		return AdminRole, nil
	case All:
		return AllRole, nil
	default:
		return Role(0), apiutil.ErrInvalidRole
	}
}

func (r Role) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

func (r *Role) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), "\"")
	val, err := ToRole(str)
	*r = val
	return err
}
