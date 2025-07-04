// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package nullable

import (
	"errors"
	"net/url"
	"strconv"
)

var ErrInvalidQueryParams = errors.New("invalid query parameters")

type nullable interface {
	~string | ~int | ~uint | ~float64
}

func Parse[T nullable](q url.Values, key string, parser Parser[T]) (Value[T], error) {
	vals, ok := q[key]
	if !ok {
		return Value[T]{}, nil
	}
	if len(vals) > 1 {
		return Value[T]{}, ErrInvalidQueryParams
	}
	s := vals[0]
	if s == "" {
		// The actual value is sent in query, so nullable is set, but empty.
		return Value[T]{Valid: true}, nil
	}
	return parser(s)
}

func ParseString(s string) (Value[string], error) {
	return Value[string]{Valid: true, Value: s}, nil
}

func ParseInt(s string) (Value[int], error) {
	val, err := strconv.Atoi(s)
	if err != nil {
		return Value[int]{}, err
	}
	return Value[int]{Valid: true, Value: val}, nil
}

func ParseFloat(s string) (Value[float64], error) {
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return Value[float64]{}, err
	}
	return Value[float64]{Valid: true, Value: val}, nil
}

func ParseBool(s string) (Value[bool], error) {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return Value[bool]{}, err
	}
	return Value[bool]{Valid: true, Value: b}, nil
}

func ParseU16(s string) (Value[uint16], error) {
	val, err := strconv.ParseUint(s, 10, 16)
	if err != nil {
		return Value[uint16]{}, err
	}
	return Value[uint16]{Valid: true, Value: uint16(val)}, nil
}

func ParseU64(s string) (Value[uint64], error) {
	val, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return Value[uint64]{}, err
	}
	return Value[uint64]{Valid: true, Value: val}, nil
}
