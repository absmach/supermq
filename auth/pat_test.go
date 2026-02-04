// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package auth_test

import (
	"testing"

	"github.com/absmach/supermq/auth"
	chOperations "github.com/absmach/supermq/channels/operations"
	cOperations "github.com/absmach/supermq/clients/operations"
	gOperations "github.com/absmach/supermq/groups/operations"
	"github.com/stretchr/testify/assert"
)

func TestEntityType(t *testing.T) {
	cases := []struct {
		desc     string
		et       auth.EntityType
		expected string
	}{
		{
			desc:     "Groups entity type",
			et:       auth.EntityType(gOperations.GroupsType),
			expected: "groups",
		},
		{
			desc:     "Channels entity type",
			et:       auth.EntityType(chOperations.ChannelsType),
			expected: "channels",
		},
		{
			desc:     "Clients entity type",
			et:       auth.EntityType(cOperations.ClientsType),
			expected: "clients",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			assert.Equal(t, tc.expected, string(tc.et), "EntityType = %v, expected %v", tc.et, tc.expected)
		})
	}
}
