// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package nullable

import (
	"errors"
	"net/url"
	"strconv"
)

var ErrInvalidQueryParams = errors.New("invalid query parameters")

func Parse[T any](q url.Values, key string, parser FromString[T]) (Nullable[T], error) {
	vals, ok := q[key]
	if !ok {
		return Nullable[T]{}, nil
	}
	if len(vals) > 1 {
		return Nullable[T]{}, ErrInvalidQueryParams
	}
	s := vals[0]
	if s == "" {
		// The actual value is sent in query, so nullable is set, but empty.
		return Nullable[T]{Set: true}, nil
	}
	return parser(s)
}

func ParseString(s string) (Nullable[string], error) {
	return Nullable[string]{Set: true, Value: s}, nil
}

func ParseInt(s string) (Nullable[int], error) {
	val, err := strconv.Atoi(s)
	if err != nil {
		return Nullable[int]{}, err
	}
	return Nullable[int]{Set: true, Value: val}, nil
}

func ParseFloat(s string) (Nullable[float64], error) {
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return Nullable[float64]{}, err
	}
	return Nullable[float64]{Set: true, Value: val}, err
}

func ParseBool(s string) (Nullable[bool], error) {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return Nullable[bool]{}, err
	}
	return Nullable[bool]{Set: true, Value: b}, nil
}

func ParseU16(s string) (Nullable[uint16], error) {
	val, err := strconv.ParseUint(s, 10, 16)
	if err != nil {
		return Nullable[uint16]{}, err
	}
	return Nullable[uint16]{Set: true, Value: uint16(val)}, nil
}

func ParseU64(s string) (Nullable[uint64], error) {
	val, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return Nullable[uint64]{}, err
	}
	return Nullable[uint64]{Set: true, Value: val}, nil
}
