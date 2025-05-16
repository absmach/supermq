// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package nullable

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Nullable type is used to represent difference betweeen an
// intentionally omitted value and default type falue.
type Nullable[T any] struct {
	Set   bool
	Value T
}

// OrElse returns the dedault value if n is not set.
func (n Nullable[T]) OrElse(defaultVal T) T {
	if n.Set {
		return n.Value
	}
	return defaultVal
}

// FromString[T any] represents a parser fucntion. It is used to avoid
// a single parser for all nullables to improve readability and performance.
// FromString should always return Nullable with Set=true, error otherwise.
type FromString[T any] func(string) (Nullable[T], error)

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
