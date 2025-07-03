// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package nullable

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Value type is used to represent difference betweeen an
// intentionally omitted value and default type value.
type Value[T any] struct {
	Set   bool
	Value T
}

func New[T Nullable](v T, set bool) Value[T] {
	return Value[T]{
		Set:   set,
		Value: v,
	}
}

// Parser[T any] represents a parser function. It is used to avoid
// a single parser for all nullables for improved readability and performance.
// Parser should always return Nullable with Set=true, error otherwise.
type Parser[T any] func(string) (Value[T], error)

// MarshalJSON encodes the value if set, otherwise returns `null`.
func (n Value[T]) MarshalJSON() ([]byte, error) {
	if !n.Set {
		return []byte("null"), nil
	}
	return json.Marshal(n.Value)
}

// UnmarshalJSON decodes JSON and sets the value and Set flag.
func (n *Value[T]) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		n.Set = false
		var empty T
		n.Value = empty
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
