package clients

import (
	"github.com/mainflux/mainflux/internal/apiutil"
	mfclients "github.com/mainflux/mainflux/pkg/clients"
)

// Role represents Client role.
type Role uint8

// Possible Client role values
const (
	UserRole Role = iota
	AdminRole
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
	default:
		return mfclients.Unknown
	}
}

// ToRole converts string value to a valid Client role.
func ToRole(status string) (Role, error) {
	switch status {
	case "", User:
		return UserRole, nil
	case Admin:
		return AdminRole, nil
	}
	return Role(0), apiutil.ErrInvalidRole
}
