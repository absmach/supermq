package clients

import (
	"encoding/json"
	"strings"

	"github.com/mainflux/mainflux/internal/apiutil"
)

// Status represents Client status.
type Status uint8

// Possible Client status values
const (
	EnabledStatus Status = iota
	DisabledStatus

	// AllStatus is used for querying purposes to list clients irrespective
	// of their status - both enabled and disabled. It is never stored in the
	// database as the actual Client status and should always be the largest
	// value in this enumeration.
	AllStatus
)

// String representation of the possible status values.
const (
	Disabled = "disabled"
	Enabled  = "enabled"
	All      = "all"
	Unknown  = "unknown"
)

// String converts client status to string literal.
func (s Status) String() string {
	switch s {
	case DisabledStatus:
		return Disabled
	case EnabledStatus:
		return Enabled
	case AllStatus:
		return All
	default:
		return Unknown
	}
}

// ToClientStatus converts string value to a valid Client status.
func ToStatus(status string) (Status, error) {
	switch status {
	case "", Enabled:
		return EnabledStatus, nil
	case Disabled:
		return DisabledStatus, nil
	case All:
		return AllStatus, nil
	}
	return Status(0), apiutil.ErrInvalidStatus
}

// Custom Marshaller for Client
func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// Custom Unmarshaler for Client
func (s *Status) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), "\"")
	val, err := ToStatus(str)
	*s = val
	return err
}
