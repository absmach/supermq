// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package auth_test

import (
	"testing"

	"github.com/absmach/supermq/auth"
	"github.com/stretchr/testify/assert"
)

func TestOperationString(t *testing.T) {
	cases := []struct {
		desc     string
		op       auth.Operation
		expected string
	}{
		{
			desc:     "Create operation",
			op:       auth.CreateOp,
			expected: "create",
		},
		{
			desc:     "Read operation",
			op:       auth.ReadOp,
			expected: "read",
		},
		{
			desc:     "List operation",
			op:       auth.ListOp,
			expected: "list",
		},
		{
			desc:     "Update operation",
			op:       auth.UpdateOp,
			expected: "update",
		},
		{
			desc:     "Delete operation",
			op:       auth.DeleteOp,
			expected: "delete",
		},
		{
			desc:     "Share operation",
			op:       auth.ShareOp,
			expected: "share",
		},
		{
			desc:     "Unshare operation",
			op:       auth.UnshareOp,
			expected: "unshare",
		},
		{
			desc:     "Publish operation",
			op:       auth.PublishOp,
			expected: "publish",
		},
		{
			desc:     "Subscribe operation",
			op:       auth.SubscribeOp,
			expected: "subscribe",
		},
		{
			desc:     "Unknown operation",
			op:       auth.Operation(100),
			expected: "unknown operation type 100",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got := tc.op.String()
			assert.Equal(t, tc.expected, got, "String() = %v, expected %v", got, tc.expected)
		})
	}
}

func TestOperationValidString(t *testing.T) {
	cases := []struct {
		desc     string
		op       auth.Operation
		expected string
		err      bool
	}{
		{
			desc:     "Valid create operation",
			op:       auth.CreateOp,
			expected: "create",
			err:      false,
		},
		{
			desc:     "Valid read operation",
			op:       auth.ReadOp,
			expected: "read",
			err:      false,
		},
		{
			desc:     "Invalid operation",
			op:       auth.Operation(100),
			expected: "",
			err:      true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := tc.op.ValidString()
			if tc.err {
				assert.Error(t, err, "ValidString() should return error")
			} else {
				assert.NoError(t, err, "ValidString() should not return error")
				assert.Equal(t, tc.expected, got, "ValidString() = %v, expected %v", got, tc.expected)
			}
		})
	}
}

func TestParseOperation(t *testing.T) {
	cases := []struct {
		desc     string
		op       string
		expected auth.Operation
		err      bool
	}{
		{
			desc:     "Parse create",
			op:       "create",
			expected: auth.CreateOp,
			err:      false,
		},
		{
			desc:     "Parse read",
			op:       "read",
			expected: auth.ReadOp,
			err:      false,
		},
		{
			desc:     "Parse list",
			op:       "list",
			expected: auth.ListOp,
			err:      false,
		},
		{
			desc:     "Parse update",
			op:       "update",
			expected: auth.UpdateOp,
			err:      false,
		},
		{
			desc:     "Parse delete",
			op:       "delete",
			expected: auth.DeleteOp,
			err:      false,
		},
		{
			desc:     "Parse share",
			op:       "share",
			expected: auth.ShareOp,
			err:      false,
		},
		{
			desc:     "Parse unshare",
			op:       "unshare",
			expected: auth.UnshareOp,
			err:      false,
		},
		{
			desc:     "Parse publish",
			op:       "publish",
			expected: auth.PublishOp,
			err:      false,
		},
		{
			desc:     "Parse subscribe",
			op:       "subscribe",
			expected: auth.SubscribeOp,
			err:      false,
		},
		{
			desc:     "Parse unknown operation",
			op:       "unknown",
			expected: auth.Operation(0),
			err:      true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := auth.ParseOperation(tc.op)
			if tc.err {
				assert.Error(t, err, "ParseOperation() should return error")
			} else {
				assert.NoError(t, err, "ParseOperation() should not return error")
				assert.Equal(t, tc.expected, got, "ParseOperation() = %v, expected %v", got, tc.expected)
			}
		})
	}
}

func TestOperationMarshalJSON(t *testing.T) {
	cases := []struct {
		desc     string
		op       auth.Operation
		expected []byte
		err      error
	}{
		{
			desc:     "Marshal create",
			op:       auth.CreateOp,
			expected: []byte(`"create"`),
			err:      nil,
		},
		{
			desc:     "Marshal read",
			op:       auth.ReadOp,
			expected: []byte(`"read"`),
			err:      nil,
		},
		{
			desc:     "Marshal delete",
			op:       auth.DeleteOp,
			expected: []byte(`"delete"`),
			err:      nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := tc.op.MarshalJSON()
			assert.Equal(t, tc.err, err, "MarshalJSON() error = %v, expected %v", err, tc.err)
			assert.Equal(t, tc.expected, got, "MarshalJSON() = %v, expected %v", got, tc.expected)
		})
	}
}

func TestOperationUnmarshalJSON(t *testing.T) {
	cases := []struct {
		desc     string
		data     []byte
		expected auth.Operation
		err      bool
	}{
		{
			desc:     "Unmarshal create",
			data:     []byte(`"create"`),
			expected: auth.CreateOp,
			err:      false,
		},
		{
			desc:     "Unmarshal read",
			data:     []byte(`"read"`),
			expected: auth.ReadOp,
			err:      false,
		},
		{
			desc:     "Unmarshal unknown",
			data:     []byte(`"unknown"`),
			expected: auth.Operation(0),
			err:      true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			var op auth.Operation
			err := op.UnmarshalJSON(tc.data)
			if tc.err {
				assert.Error(t, err, "UnmarshalJSON() should return error")
			} else {
				assert.NoError(t, err, "UnmarshalJSON() should not return error")
				assert.Equal(t, tc.expected, op, "UnmarshalJSON() = %v, expected %v", op, tc.expected)
			}
		})
	}
}

func TestOperationMarshalText(t *testing.T) {
	cases := []struct {
		desc     string
		op       auth.Operation
		expected []byte
		err      error
	}{
		{
			desc:     "Marshal create as text",
			op:       auth.CreateOp,
			expected: []byte("create"),
			err:      nil,
		},
		{
			desc:     "Marshal read as text",
			op:       auth.ReadOp,
			expected: []byte("read"),
			err:      nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := tc.op.MarshalText()
			assert.Equal(t, tc.err, err, "MarshalText() error = %v, expected %v", err, tc.err)
			assert.Equal(t, tc.expected, got, "MarshalText() = %v, expected %v", got, tc.expected)
		})
	}
}

func TestOperationUnmarshalText(t *testing.T) {
	cases := []struct {
		desc     string
		data     []byte
		expected auth.Operation
		err      bool
	}{
		{
			desc:     "Unmarshal create from text",
			data:     []byte("create"),
			expected: auth.CreateOp,
			err:      false,
		},
		{
			desc:     "Unmarshal read from text",
			data:     []byte("read"),
			expected: auth.ReadOp,
			err:      false,
		},
		{
			desc:     "Unmarshal unknown from text",
			data:     []byte("unknown"),
			expected: auth.Operation(0),
			err:      true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			var op auth.Operation
			err := op.UnmarshalText(tc.data)
			if tc.err {
				assert.Error(t, err, "UnmarshalText() should return error")
			} else {
				assert.NoError(t, err, "UnmarshalText() should not return error")
				assert.Equal(t, tc.expected, op, "UnmarshalText() = %v, expected %v", op, tc.expected)
			}
		})
	}
}

func TestEntityTypeString(t *testing.T) {
	cases := []struct {
		desc     string
		et       auth.EntityType
		expected string
	}{
		{
			desc:     "Groups entity type",
			et:       auth.GroupsType,
			expected: "groups",
		},
		{
			desc:     "Channels entity type",
			et:       auth.ChannelsType,
			expected: "channels",
		},
		{
			desc:     "Clients entity type",
			et:       auth.ClientsType,
			expected: "clients",
		},
		{
			desc:     "Domains entity type",
			et:       auth.DomainsType,
			expected: "domains",
		},
		{
			desc:     "Users entity type",
			et:       auth.UsersType,
			expected: "users",
		},
		{
			desc:     "Dashboard entity type",
			et:       auth.DashboardType,
			expected: "dashboards",
		},
		{
			desc:     "Messages entity type",
			et:       auth.MessagesType,
			expected: "messages",
		},
		{
			desc:     "Unknown entity type",
			et:       auth.EntityType(100),
			expected: "unknown domain entity type 100",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got := tc.et.String()
			assert.Equal(t, tc.expected, got, "String() = %v, expected %v", got, tc.expected)
		})
	}
}

func TestParseEntityType(t *testing.T) {
	cases := []struct {
		desc     string
		et       string
		expected auth.EntityType
		err      bool
	}{
		{
			desc:     "Parse groups",
			et:       "groups",
			expected: auth.GroupsType,
			err:      false,
		},
		{
			desc:     "Parse channels",
			et:       "channels",
			expected: auth.ChannelsType,
			err:      false,
		},
		{
			desc:     "Parse clients",
			et:       "clients",
			expected: auth.ClientsType,
			err:      false,
		},
		{
			desc:     "Parse domains",
			et:       "domains",
			expected: auth.DomainsType,
			err:      false,
		},
		{
			desc:     "Parse users",
			et:       "users",
			expected: auth.UsersType,
			err:      false,
		},
		{
			desc:     "Parse dashboards",
			et:       "dashboards",
			expected: auth.DashboardType,
			err:      false,
		},
		{
			desc:     "Parse unknown entity type",
			et:       "unknown",
			expected: auth.EntityType(0),
			err:      true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := auth.ParseEntityType(tc.et)
			if tc.err {
				assert.Error(t, err, "ParseEntityType() should return error")
			} else {
				assert.NoError(t, err, "ParseEntityType() should not return error")
				assert.Equal(t, tc.expected, got, "ParseEntityType() = %v, expected %v", got, tc.expected)
			}
		})
	}
}

func TestEntityTypeMarshalJSON(t *testing.T) {
	cases := []struct {
		desc     string
		et       auth.EntityType
		expected []byte
		err      error
	}{
		{
			desc:     "Marshal groups",
			et:       auth.GroupsType,
			expected: []byte(`"groups"`),
			err:      nil,
		},
		{
			desc:     "Marshal channels",
			et:       auth.ChannelsType,
			expected: []byte(`"channels"`),
			err:      nil,
		},
		{
			desc:     "Marshal clients",
			et:       auth.ClientsType,
			expected: []byte(`"clients"`),
			err:      nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := tc.et.MarshalJSON()
			assert.Equal(t, tc.err, err, "MarshalJSON() error = %v, expected %v", err, tc.err)
			assert.Equal(t, tc.expected, got, "MarshalJSON() = %v, expected %v", got, tc.expected)
		})
	}
}

func TestEntityTypeUnmarshalJSON(t *testing.T) {
	cases := []struct {
		desc     string
		data     []byte
		expected auth.EntityType
		err      bool
	}{
		{
			desc:     "Unmarshal groups",
			data:     []byte(`"groups"`),
			expected: auth.GroupsType,
			err:      false,
		},
		{
			desc:     "Unmarshal channels",
			data:     []byte(`"channels"`),
			expected: auth.ChannelsType,
			err:      false,
		},
		{
			desc:     "Unmarshal unknown",
			data:     []byte(`"unknown"`),
			expected: auth.EntityType(0),
			err:      true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			var et auth.EntityType
			err := et.UnmarshalJSON(tc.data)
			if tc.err {
				assert.Error(t, err, "UnmarshalJSON() should return error")
			} else {
				assert.NoError(t, err, "UnmarshalJSON() should not return error")
				assert.Equal(t, tc.expected, et, "UnmarshalJSON() = %v, expected %v", et, tc.expected)
			}
		})
	}
}

func TestEntityTypeMarshalText(t *testing.T) {
	cases := []struct {
		desc     string
		et       auth.EntityType
		expected []byte
		err      error
	}{
		{
			desc:     "Marshal groups as text",
			et:       auth.GroupsType,
			expected: []byte("groups"),
			err:      nil,
		},
		{
			desc:     "Marshal channels as text",
			et:       auth.ChannelsType,
			expected: []byte("channels"),
			err:      nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got, err := tc.et.MarshalText()
			assert.Equal(t, tc.err, err, "MarshalText() error = %v, expected %v", err, tc.err)
			assert.Equal(t, tc.expected, got, "MarshalText() = %v, expected %v", got, tc.expected)
		})
	}
}

func TestEntityTypeUnmarshalText(t *testing.T) {
	cases := []struct {
		desc     string
		data     []byte
		expected auth.EntityType
		err      bool
	}{
		{
			desc:     "Unmarshal groups from text",
			data:     []byte("groups"),
			expected: auth.GroupsType,
			err:      false,
		},
		{
			desc:     "Unmarshal channels from text",
			data:     []byte("channels"),
			expected: auth.ChannelsType,
			err:      false,
		},
		{
			desc:     "Unmarshal unknown from text",
			data:     []byte("unknown"),
			expected: auth.EntityType(0),
			err:      true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			var et auth.EntityType
			err := et.UnmarshalText(tc.data)
			if tc.err {
				assert.Error(t, err, "UnmarshalText() should return error")
			} else {
				assert.NoError(t, err, "UnmarshalText() should not return error")
				assert.Equal(t, tc.expected, et, "UnmarshalText() = %v, expected %v", et, tc.expected)
			}
		})
	}
}
