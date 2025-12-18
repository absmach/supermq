// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package oauth2_test

import (
	"testing"

	"github.com/absmach/supermq/pkg/oauth2"
	"github.com/absmach/supermq/users"
	"github.com/stretchr/testify/assert"
)

func TestNormalizeUser(t *testing.T) {
	cases := []struct {
		desc       string
		inputJSON  string
		provider   string
		wantUser   users.User
		wantErrStr string
	}{
		{
			desc: "valid user with Google keys (given_name, family_name)",
			inputJSON: `{
				"id": "123",
				"given_name": "Jane",
				"family_name": "Doe",
				"email": "jane@example.com",
				"picture": "pic.jpg"
			}`,
			provider: "google",
			wantUser: users.User{
				ID:             "123",
				FirstName:      "Jane",
				LastName:       "Doe",
				Email:          "jane@example.com",
				ProfilePicture: "pic.jpg",
				Metadata:       users.Metadata{"oauth_provider": "google"},
			},
			wantErrStr: "",
		},
		{
			desc: "valid user with alternative key variants (givenName, familyName, emailAddress, profilePicture)",
			inputJSON: `{
				"id": "456",
				"givenName": "John",
				"familyName": "Smith",
				"user_name": "jsmith",
				"emailAddress": "john@smith.com",
				"profilePicture": "avatar.png"
			}`,
			provider: "github",
			wantUser: users.User{
				ID:             "456",
				FirstName:      "John",
				LastName:       "Smith",
				Email:          "john@smith.com",
				ProfilePicture: "avatar.png",
				Metadata:       users.Metadata{"oauth_provider": "github"},
			},
			wantErrStr: "",
		},
		{
			desc: "valid user with snake_case variants (first_name, last_name, email_address, profile_picture)",
			inputJSON: `{
				"id": "789",
				"first_name": "Alice",
				"last_name": "Brown",
				"username": "abrown",
				"email_address": "alice@brown.com",
				"profile_picture": "photo.jpg"
			}`,
			provider: "custom",
			wantUser: users.User{
				ID:             "789",
				FirstName:      "Alice",
				LastName:       "Brown",
				Email:          "alice@brown.com",
				ProfilePicture: "photo.jpg",
				Metadata:       users.Metadata{"oauth_provider": "custom"},
			},
			wantErrStr: "",
		},
		{
			desc: "valid user with lowercase variants (firstname, lastname, avatar)",
			inputJSON: `{
				"id": "101112",
				"firstname": "Bob",
				"lastname": "Wilson",
				"userName": "bwilson",
				"email": "bob@wilson.com",
				"avatar": "img.jpg"
			}`,
			provider: "oauth",
			wantUser: users.User{
				ID:             "101112",
				FirstName:      "Bob",
				LastName:       "Wilson",
				Email:          "bob@wilson.com",
				ProfilePicture: "img.jpg",
				Metadata:       users.Metadata{"oauth_provider": "oauth"},
			},
			wantErrStr: "",
		},
		{
			desc: "valid user with minimal required fields only",
			inputJSON: `{
				"id": "999",
				"given_name": "Min",
				"family_name": "Max",
				"email": "min@max.com"
			}`,
			provider: "minimal",
			wantUser: users.User{
				ID:             "999",
				FirstName:      "Min",
				LastName:       "Max",
				Email:          "min@max.com",
				ProfilePicture: "",
				Metadata:       users.Metadata{"oauth_provider": "minimal"},
			},
			wantErrStr: "",
		},
		{
			desc: "missing ID field",
			inputJSON: `{
				"given_name": "Jane",
				"family_name": "Doe",
				"email": "jane@example.com"
			}`,
			provider:   "google",
			wantUser:   users.User{},
			wantErrStr: "missing required fields: id",
		},
		{
			desc: "missing first_name field",
			inputJSON: `{
				"id": "123",
				"family_name": "Doe",
				"email": "jane@example.com"
			}`,
			provider:   "google",
			wantUser:   users.User{},
			wantErrStr: "missing required fields: first_name",
		},
		{
			desc: "missing last_name field",
			inputJSON: `{
				"id": "123",
				"given_name": "Jane",
				"email": "jane@example.com"
			}`,
			provider:   "google",
			wantUser:   users.User{},
			wantErrStr: "missing required fields: last_name",
		},
		{
			desc: "missing email field",
			inputJSON: `{
				"id": "123",
				"given_name": "Jane",
				"family_name": "Doe"
			}`,
			provider:   "google",
			wantUser:   users.User{},
			wantErrStr: "missing required fields: email",
		},
		{
			desc: "missing multiple required fields",
			inputJSON: `{
				"given_name": "Jane"
			}`,
			provider:   "google",
			wantUser:   users.User{},
			wantErrStr: "missing required fields: id, last_name, email",
		},
		{
			desc:       "missing all required fields",
			inputJSON:  `{}`,
			provider:   "google",
			wantUser:   users.User{},
			wantErrStr: "missing required fields: id, first_name, last_name, email",
		},
		{
			desc:       "invalid JSON syntax",
			inputJSON:  `{invalid json`,
			provider:   "google",
			wantUser:   users.User{},
			wantErrStr: "invalid character",
		},
		{
			desc:       "empty JSON",
			inputJSON:  ``,
			provider:   "google",
			wantUser:   users.User{},
			wantErrStr: "unexpected end of JSON input",
		},
		{
			desc: "unrecognized keys are ignored",
			inputJSON: `{
				"id": "567",
				"given_name": "Test",
				"family_name": "User",
				"email": "test@user.com",
				"unrecognized_field": "ignored",
				"another_field": 12345
			}`,
			provider: "test",
			wantUser: users.User{
				ID:             "567",
				FirstName:      "Test",
				LastName:       "User",
				Email:          "test@user.com",
				ProfilePicture: "",
				Metadata:       users.Metadata{"oauth_provider": "test"},
			},
			wantErrStr: "",
		},
		{
			desc: "key priority - first matching variant is used",
			inputJSON: `{
				"id": "priority",
				"given_name": "First",
				"first_name": "Second",
				"family_name": "Family1",
				"last_name": "Family2",
				"email": "email1@test.com",
				"email_address": "email2@test.com"
			}`,
			provider: "priority",
			wantUser: users.User{
				ID:             "priority",
				FirstName:      "First",
				LastName:       "Family1",
				Email:          "email1@test.com",
				ProfilePicture: "",
				Metadata:       users.Metadata{"oauth_provider": "priority"},
			},
			wantErrStr: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			user, err := oauth2.NormalizeUser([]byte(tc.inputJSON), tc.provider)
			if tc.wantErrStr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErrStr)
				assert.Equal(t, tc.wantUser, user)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantUser, user)
			}
		})
	}
}
