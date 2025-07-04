// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package nullable

import (
	"net/url"
	"strconv"
	"testing"
)

func BenchmarkPointerString(b *testing.B) {
	p := func() *string {
		x := "test"
		return &x
	}()
	for b.Loop() {
		if p != nil {
			_ = *p + "test"
		}
	}
}

func BenchmarkNullableString(b *testing.B) {
	n := func() Value[string] {
		return Value[string]{Valid: true, Value: "test"}
	}()

	for b.Loop() {
		if n.Valid {
			_ = n.Value + "test"
		}
	}
}

func BenchmarkNullableStringParse(b *testing.B) {
	for b.Loop() {
		n, _ := ParseString("123")
		_ = n.Value
	}
}

func BenchmarkPointerStringParse(b *testing.B) {
	parser := func(s string) (*string, error) {
		return &s, nil
	}

	for b.Loop() {
		n, _ := parser("123")
		_ = *n
	}
}

func BenchmarkPointerInt(b *testing.B) {
	p := func() *int {
		x := 42
		return &x
	}()

	for b.Loop() {
		if p != nil {
			_ = *p + 1
		}
	}
}

func BenchmarkNullableInt(b *testing.B) {
	n := func() Value[int] {
		return Value[int]{Valid: true, Value: 42}
	}()

	for b.Loop() {
		if n.Valid {
			_ = n.Value + 1
		}
	}
}

func BenchmarkNullableIntParse(b *testing.B) {
	for b.Loop() {
		n, _ := ParseInt("123")
		_ = n.Value
	}
}

func BenchmarkPointerIntParse(b *testing.B) {
	parser := func(s string) (*int, error) {
		v, err := strconv.Atoi(s)
		if err != nil {
			return nil, err
		}
		return &v, nil
	}

	for b.Loop() {
		n, _ := parser("123")
		_ = *n
	}
}

func BenchmarkPointerFloat(b *testing.B) {
	p := func() *float64 {
		x := float64(42)
		return &x
	}()

	for b.Loop() {
		if p != nil {
			_ = *p + 1
		}
	}
}

func BenchmarkNullableFloat(b *testing.B) {
	n := func() Value[float64] {
		return Value[float64]{Valid: true, Value: 42}
	}()

	for b.Loop() {
		if n.Valid {
			_ = n.Value + 1
		}
	}
}

func BenchmarkNullableFloatParse(b *testing.B) {
	for b.Loop() {
		n, _ := ParseFloat("123.45")
		_ = n.Value
	}
}

func BenchmarkPointerFloatParse(b *testing.B) {
	parser := func(s string) (*float64, error) {
		v, err := strconv.ParseFloat(s, 10)
		if err != nil {
			return nil, err
		}
		return &v, nil
	}

	for b.Loop() {
		n, _ := parser("123.45")
		_ = *n
	}
}

func BenchmarkParseNullable(b *testing.B) {
	for b.Loop() {
		val, _ := Parse(url.Values{"key": []string{"123.456"}}, "key", ParseFloat)
		_ = val
	}
}

func BenchmarkParsePointer(b *testing.B) {
	parser := func(q url.Values, key string) (*float64, error) {
		vals, ok := q[key]
		if !ok {
			return nil, nil
		}
		if len(vals) > 1 {
			return nil, ErrInvalidQueryParams
		}
		s := vals[0]
		if s == "" {
			return nil, nil // not nil, but empty
		}

		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, err
		}
		return &v, nil
	}

	for b.Loop() {
		val, _ := parser(url.Values{"key": []string{"123.456"}}, "key")
		_ = val
	}
}
