package nullable

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type Nullable[T any] struct {
	Set   bool
	Value T
}

// New creates a new Nullable with a value.
func New[T any](v T) Nullable[T] {
	return Nullable[T]{Set: true, Value: v}
}

func (n Nullable[T]) IsSet() bool {
	return n.Set
}

func (n Nullable[T]) Get() (T, bool) {
	return n.Value, n.Set
}

func (n Nullable[T]) OrElse(defaultVal T) T {
	if n.Set {
		return n.Value
	}
	return defaultVal
}

// MarshalJSON encodes the value if set, otherwise returns `null`.
func (n Nullable[T]) MarshalJSON() ([]byte, error) {
	if !n.Set {
		return []byte("null"), nil
	}
	return json.Marshal(n.Value)
}

// UnmarshalJSON decodes JSON and sets the value and Set flag.
func (n *Nullable[T]) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		n.Set = false
		var zero T
		n.Value = zero
		return nil
	}

	var val T
	if err := json.Unmarshal(data, &val); err != nil {
		return fmt.Errorf("nullable: failed to unmarshal: %w", err)
	}
	n.Value = val
	n.Set = true
	return nil
}
