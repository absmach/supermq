// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0
package nullable

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseHelpers(t *testing.T) {
	t.Run("ParseString", func(t *testing.T) {
		val, err := ParseString("hello")
		assert.NoError(t, err)
		assert.True(t, val.Set)
		assert.Equal(t, "hello", val.Value)
	})

	t.Run("ParseInt", func(t *testing.T) {
		val, err := ParseInt("42")
		assert.NoError(t, err)
		assert.True(t, val.Set)
		assert.Equal(t, 42, val.Value)

		val, err = ParseInt("notanint")
		assert.Error(t, err)
		assert.False(t, val.Set)
	})

	t.Run("ParseFloat", func(t *testing.T) {
		val, err := ParseFloat("3.14")
		assert.NoError(t, err)
		assert.True(t, val.Set)
		assert.Equal(t, 3.14, val.Value)
	})

	t.Run("ParseBool", func(t *testing.T) {
		val, err := ParseBool("true")
		assert.NoError(t, err)
		assert.True(t, val.Set)
		assert.True(t, val.Value)

		val, err = ParseBool("false")
		assert.NoError(t, err)
		assert.True(t, val.Set)
		assert.False(t, val.Value)

		val, err = ParseBool("maybe")
		assert.Error(t, err)
		assert.False(t, val.Set)
	})

	t.Run("ParseU16", func(t *testing.T) {
		val, err := ParseU16("65535")
		assert.NoError(t, err)
		assert.True(t, val.Set)
		assert.Equal(t, uint16(65535), val.Value)

		val, err = ParseU16("70000")
		assert.Error(t, err)
		assert.False(t, val.Set)
	})

	t.Run("ParseU64", func(t *testing.T) {
		val, err := ParseU64("1234567890")
		assert.NoError(t, err)
		assert.True(t, val.Set)
		assert.Equal(t, uint64(1234567890), val.Value)
	})
}

func TestParseQueryParam(t *testing.T) {
	type useCase struct {
		name      string
		query     url.Values
		key       string
		parser    func(string) (Value[int], error)
		expect    Value[int]
		expectErr bool
	}

	cases := []useCase{
		{
			name:   "missing key",
			query:  url.Values{},
			key:    "limit",
			parser: ParseInt,
			expect: Value[int]{Set: false},
		},
		{
			name:   "empty value",
			query:  url.Values{"limit": {""}},
			key:    "limit",
			parser: ParseInt,
			expect: Value[int]{Set: true},
		},
		{
			name:   "valid int",
			query:  url.Values{"limit": {"10"}},
			key:    "limit",
			parser: ParseInt,
			expect: Value[int]{Set: true, Value: 10},
		},
		{
			name:      "invalid int",
			query:     url.Values{"limit": {"bad"}},
			key:       "limit",
			parser:    ParseInt,
			expectErr: true,
		},
		{
			name:      "multiple values",
			query:     url.Values{"limit": {"1", "2"}},
			key:       "limit",
			parser:    ParseInt,
			expectErr: true,
		},
	}

	for _, uc := range cases {
		t.Run(uc.name, func(t *testing.T) {
			val, err := Parse(uc.query, uc.key, uc.parser)
			if uc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, uc.expect, val)
			}
		})
	}
}
