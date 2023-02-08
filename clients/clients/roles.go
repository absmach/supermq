package clients

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
		return Unknown
	}
}
