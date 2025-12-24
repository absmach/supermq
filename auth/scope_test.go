// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package auth_test

import (
	"testing"
	"time"

	apiutil "github.com/absmach/supermq/api/http/util"
	"github.com/absmach/supermq/auth"
	"github.com/stretchr/testify/assert"
)

func TestScopeAuthorized(t *testing.T) {
	cases := []struct {
		desc             string
		scope            *auth.Scope
		entityType       auth.EntityType
		optionalDomainID string
		operation        auth.Operation
		entityID         string
		expected         bool
	}{
		{
			desc: "Authorized with matching entity type, domain, operation and entity ID",
			scope: &auth.Scope{
				EntityType:       auth.GroupsType,
				OptionalDomainID: "domain1",
				Operation:        auth.CreateOp,
				EntityID:         "entity1",
			},
			entityType:       auth.GroupsType,
			optionalDomainID: "domain1",
			operation:        auth.CreateOp,
			entityID:         "entity1",
			expected:         true,
		},
		{
			desc: "Authorized with wildcard entity ID",
			scope: &auth.Scope{
				EntityType:       auth.GroupsType,
				OptionalDomainID: "domain1",
				Operation:        auth.CreateOp,
				EntityID:         "*",
			},
			entityType:       auth.GroupsType,
			optionalDomainID: "domain1",
			operation:        auth.CreateOp,
			entityID:         "any-entity",
			expected:         true,
		},
		{
			desc: "Authorized without domain ID",
			scope: &auth.Scope{
				EntityType:       auth.UsersType,
				OptionalDomainID: "",
				Operation:        auth.ReadOp,
				EntityID:         "user1",
			},
			entityType:       auth.UsersType,
			optionalDomainID: "",
			operation:        auth.ReadOp,
			entityID:         "user1",
			expected:         true,
		},
		{
			desc: "Not authorized with different entity type",
			scope: &auth.Scope{
				EntityType:       auth.GroupsType,
				OptionalDomainID: "domain1",
				Operation:        auth.CreateOp,
				EntityID:         "entity1",
			},
			entityType:       auth.ChannelsType,
			optionalDomainID: "domain1",
			operation:        auth.CreateOp,
			entityID:         "entity1",
			expected:         false,
		},
		{
			desc: "Not authorized with different domain ID",
			scope: &auth.Scope{
				EntityType:       auth.GroupsType,
				OptionalDomainID: "domain1",
				Operation:        auth.CreateOp,
				EntityID:         "entity1",
			},
			entityType:       auth.GroupsType,
			optionalDomainID: "domain2",
			operation:        auth.CreateOp,
			entityID:         "entity1",
			expected:         false,
		},
		{
			desc: "Not authorized with different operation",
			scope: &auth.Scope{
				EntityType:       auth.GroupsType,
				OptionalDomainID: "domain1",
				Operation:        auth.CreateOp,
				EntityID:         "entity1",
			},
			entityType:       auth.GroupsType,
			optionalDomainID: "domain1",
			operation:        auth.DeleteOp,
			entityID:         "entity1",
			expected:         false,
		},
		{
			desc: "Not authorized with different entity ID",
			scope: &auth.Scope{
				EntityType:       auth.GroupsType,
				OptionalDomainID: "domain1",
				Operation:        auth.CreateOp,
				EntityID:         "entity1",
			},
			entityType:       auth.GroupsType,
			optionalDomainID: "domain1",
			operation:        auth.CreateOp,
			entityID:         "entity2",
			expected:         false,
		},
		{
			desc:             "Not authorized with nil scope",
			scope:            nil,
			entityType:       auth.GroupsType,
			optionalDomainID: "domain1",
			operation:        auth.CreateOp,
			entityID:         "entity1",
			expected:         false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			result := tc.scope.Authorized(tc.entityType, tc.optionalDomainID, tc.operation, tc.entityID)
			assert.Equal(t, tc.expected, result, "Authorized() = %v, expected %v", result, tc.expected)
		})
	}
}

func TestScopeValidate(t *testing.T) {
	cases := []struct {
		desc  string
		scope *auth.Scope
		err   error
	}{
		{
			desc: "Valid scope for groups with domain ID",
			scope: &auth.Scope{
				EntityType:       auth.GroupsType,
				OptionalDomainID: "domain1",
				Operation:        auth.CreateOp,
				EntityID:         "entity1",
			},
			err: nil,
		},
		{
			desc: "Valid scope for channels with domain ID",
			scope: &auth.Scope{
				EntityType:       auth.ChannelsType,
				OptionalDomainID: "domain1",
				Operation:        auth.ReadOp,
				EntityID:         "channel1",
			},
			err: nil,
		},
		{
			desc: "Valid scope for clients with domain ID",
			scope: &auth.Scope{
				EntityType:       auth.ClientsType,
				OptionalDomainID: "domain1",
				Operation:        auth.UpdateOp,
				EntityID:         "client1",
			},
			err: nil,
		},
		{
			desc: "Valid scope for users without domain ID",
			scope: &auth.Scope{
				EntityType:       auth.UsersType,
				OptionalDomainID: "",
				Operation:        auth.DeleteOp,
				EntityID:         "user1",
			},
			err: nil,
		},
		{
			desc: "Valid scope for domains without domain ID",
			scope: &auth.Scope{
				EntityType:       auth.DomainsType,
				OptionalDomainID: "",
				Operation:        auth.ListOp,
				EntityID:         "domain1",
			},
			err: nil,
		},
		{
			desc: "Valid scope with wildcard entity ID",
			scope: &auth.Scope{
				EntityType:       auth.GroupsType,
				OptionalDomainID: "domain1",
				Operation:        auth.CreateOp,
				EntityID:         "*",
			},
			err: nil,
		},
		{
			desc:  "Invalid nil scope",
			scope: nil,
			err:   assert.AnError, // Will be checked with Contains
		},
		{
			desc: "Invalid scope without entity ID",
			scope: &auth.Scope{
				EntityType:       auth.GroupsType,
				OptionalDomainID: "domain1",
				Operation:        auth.CreateOp,
				EntityID:         "",
			},
			err: apiutil.ErrMissingEntityID,
		},
		{
			desc: "Invalid scope for groups without domain ID",
			scope: &auth.Scope{
				EntityType:       auth.GroupsType,
				OptionalDomainID: "",
				Operation:        auth.CreateOp,
				EntityID:         "entity1",
			},
			err: apiutil.ErrMissingDomainID,
		},
		{
			desc: "Invalid scope for channels without domain ID",
			scope: &auth.Scope{
				EntityType:       auth.ChannelsType,
				OptionalDomainID: "",
				Operation:        auth.CreateOp,
				EntityID:         "channel1",
			},
			err: apiutil.ErrMissingDomainID,
		},
		{
			desc: "Invalid scope for clients without domain ID",
			scope: &auth.Scope{
				EntityType:       auth.ClientsType,
				OptionalDomainID: "",
				Operation:        auth.CreateOp,
				EntityID:         "client1",
			},
			err: apiutil.ErrMissingDomainID,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.scope.Validate()
			if tc.err != nil {
				assert.Error(t, err, "Validate() should return error")
				if tc.err != assert.AnError {
					assert.Equal(t, tc.err, err, "Validate() error = %v, expected %v", err, tc.err)
				}
			} else {
				assert.NoError(t, err, "Validate() should not return error")
			}
		})
	}
}

func TestPATValidate(t *testing.T) {
	cases := []struct {
		desc string
		pat  *auth.PAT
		err  bool
	}{
		{
			desc: "Valid PAT",
			pat: &auth.PAT{
				ID:          "pat-id",
				User:        "user-id",
				Name:        "test-pat",
				Description: "test description",
			},
			err: false,
		},
		{
			desc: "Invalid nil PAT",
			pat:  nil,
			err:  true,
		},
		{
			desc: "Invalid PAT without name",
			pat: &auth.PAT{
				ID:          "pat-id",
				User:        "user-id",
				Name:        "",
				Description: "test description",
			},
			err: true,
		},
		{
			desc: "Invalid PAT without user",
			pat: &auth.PAT{
				ID:          "pat-id",
				User:        "",
				Name:        "test-pat",
				Description: "test description",
			},
			err: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.pat.Validate()
			if tc.err {
				assert.Error(t, err, "Validate() should return error")
			} else {
				assert.NoError(t, err, "Validate() should not return error")
			}
		})
	}
}

func TestPATMarshalUnmarshalBinary(t *testing.T) {
	pat := auth.PAT{
		ID:          "pat-id",
		User:        "user-id",
		Name:        "test-pat",
		Description: "test description",
		Secret:      "secret",
		IssuedAt:    time.Now().UTC().Round(time.Second),
		ExpiresAt:   time.Now().UTC().Add(24 * time.Hour).Round(time.Second),
		Status:      auth.ActiveStatus,
	}

	// Marshal
	data, err := pat.MarshalBinary()
	assert.NoError(t, err, "MarshalBinary() should not return error")
	assert.NotNil(t, data, "MarshalBinary() should return data")

	// Unmarshal
	var newPAT auth.PAT
	err = newPAT.UnmarshalBinary(data)
	assert.NoError(t, err, "UnmarshalBinary() should not return error")

	assert.Equal(t, pat.ID, newPAT.ID, "ID mismatch")
	assert.Equal(t, pat.User, newPAT.User, "User mismatch")
	assert.Equal(t, pat.Name, newPAT.Name, "Name mismatch")
	assert.Equal(t, pat.Description, newPAT.Description, "Description mismatch")
	assert.Equal(t, pat.Secret, newPAT.Secret, "Secret mismatch")
	assert.Equal(t, pat.Status, newPAT.Status, "Status mismatch")
}

func TestPATString(t *testing.T) {
	pat := &auth.PAT{
		ID:          "pat-id",
		User:        "user-id",
		Name:        "test-pat",
		Description: "test description",
		Status:      auth.ActiveStatus,
	}

	str := pat.String()
	assert.NotEmpty(t, str, "String() should return non-empty string")
	assert.Contains(t, str, "pat-id", "String() should contain ID")
	assert.Contains(t, str, "user-id", "String() should contain User")
	assert.Contains(t, str, "test-pat", "String() should contain Name")
}
