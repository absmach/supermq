// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package nullable

import (
	"errors"
	"net/url"
	"strconv"
)

var ErrInvalidQueryParams = errors.New("invalid query parameters")

func ParseString(s string) (Nullable[string], error) {
	ret := Nullable[string]{Value: s}
	if s != "" {
		ret.Set = true
	}
	return ret, nil
}

func ParseFloatFromQuery(query url.Values, key string, def float64) (Nullable[float64], error) {
	vals, ok := query[key]
	if len(vals) > 1 {
		return Nullable[float64]{}, ErrInvalidQueryParams
	}

	if !ok {
		return Nullable[float64]{Set: false, Value: def}, nil
	}
	return ParseFloat(vals[0])
}

func ParseIntFromQuery(query url.Values, key string, def int) (Nullable[int], error) {
	vals, ok := query[key]
	if len(vals) > 1 {
		return Nullable[int]{}, ErrInvalidQueryParams
	}

	if !ok {
		return Nullable[int]{Set: false, Value: def}, nil
	}
	return ParseInt(vals[0])
}

func ParseInt(s string) (Nullable[int], error) {
	if s == "" {
		return Nullable[int]{}, nil
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return Nullable[int]{}, err
	}
	return Nullable[int]{Set: true, Value: val}, nil
}

func ParseFloat(s string) (Nullable[float64], error) {
	if s == "" {
		return Nullable[float64]{}, nil
	}
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return Nullable[float64]{}, err
	}
	return Nullable[float64]{Set: true, Value: val}, err
}

func ParseBool(s string) (Nullable[bool], error) {
	if s == "" {
		return Nullable[bool]{}, nil
	}
	b, err := strconv.ParseBool(s)
	if err != nil {
		return Nullable[bool]{}, err
	}
	return Nullable[bool]{Set: true, Value: b}, nil
}

func ParseU16(s string) (Nullable[uint16], error) {
	if s == "" {
		return Nullable[uint16]{}, nil
	}
	val, err := strconv.ParseUint(s, 10, 16)
	if err != nil {
		return Nullable[uint16]{}, err
	}
	return Nullable[uint16]{Set: true, Value: uint16(val)}, nil
}

func ParseU64(s string) (Nullable[uint64], error) {
	if s == "" {
		return Nullable[uint64]{}, nil
	}
	val, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return Nullable[uint64]{}, err
	}
	return Nullable[uint64]{Set: true, Value: val}, nil
}
