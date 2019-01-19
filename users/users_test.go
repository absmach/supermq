// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package users_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mainflux/mainflux/errors"
	"github.com/mainflux/mainflux/users"
	"github.com/stretchr/testify/assert"
)

const (
	email           = "user@example.com"
	password        = "password"
	sampleUtfString = "\xe7\x94\xa8\xe6\x88\xb7" // 用户
)

func TestValidate(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		user users.User
		err  error
	}{
		"validate user with valid data": {
			user: users.User{
				Email:    email,
				Password: password,
			},
			err: nil,
		},
		"validate user with empty email": {
			user: users.User{
				Email:    "",
				Password: password,
			},
			err: users.ErrMalformedEntity,
		},
		"validate user with empty password": {
			user: users.User{
				Email:    email,
				Password: "",
			},
			err: users.ErrMalformedEntity,
		},
		"validate email without at": {
			user: users.User{
				Email:    "userexample.com",
				Password: password,
			},
			err: users.ErrMalformedEntity,
		},
		"validate email with unexpected characters in local": {
			user: users.User{
				Email:    "<user>@xample.com",
				Password: password,
			},
			err: users.ErrMalformedEntity,
		},
		"validate email without local": {
			user: users.User{
				Email:    " @example.com",
				Password: password,
			},
			err: users.ErrMalformedEntity,
		},
		"validate email without host": {
			user: users.User{
				Email:    "user@ ",
				Password: password,
			},
			err: users.ErrMalformedEntity,
		},
		"validate email with consecutive dots in local": {
			user: users.User{
				Email:    "user..email@example.com",
				Password: password,
			},
			err: users.ErrMalformedEntity,
		},
		"validate email with consecutive dots in host": {
			user: users.User{
				Email:    "useremail@example..com",
				Password: password,
			},
			err: users.ErrMalformedEntity,
		},
		"validate email starting with dot": {
			user: users.User{
				Email:    ".useremail@example.com",
				Password: password,
			},
			err: users.ErrMalformedEntity,
		},
		"validate email ending with dot": {
			user: users.User{
				Email:    "useremail.@example.com",
				Password: password,
			},
			err: users.ErrMalformedEntity,
		},
		"validate email with dots": {
			user: users.User{
				Email:    "user.ema.ill@example.com",
				Password: password,
			},
			err: nil,
		},
		"validate email with utf-8": {
			user: users.User{
				Email:    "用户+“偽名”@例子.广告",
				Password: password,
			},
			err: nil,
		},
		"validate email with single letter segments": {
			user: users.User{
				Email:    "u@e",
				Password: password,
			},
			err: nil,
		},
		"validate email with 64B local": {
			user: users.User{
				Email:    fmt.Sprintf("%s%s@example.com", sampleUtfString, strings.Repeat("a", 64-len(sampleUtfString))),
				Password: password,
			},
			err: nil,
		},
		"validate email with 255B host": {
			user: users.User{
				Email:    fmt.Sprintf("user@%s%s", sampleUtfString, strings.Repeat("a", 255-len(sampleUtfString))),
				Password: password,
			},
			err: nil,
		},
		"validate email with 65B local": {
			user: users.User{
				Email:    fmt.Sprintf("%s%s@example.com`", sampleUtfString, strings.Repeat(sampleUtfString, 65-len(sampleUtfString))),
				Password: password,
			},
			err: users.ErrMalformedEntity,
		},
		"validate email with 256B host": {
			user: users.User{
				Email:    fmt.Sprintf("user@%s%s", sampleUtfString, strings.Repeat(sampleUtfString, 256-len(sampleUtfString))),
				Password: password,
			},
			err: users.ErrMalformedEntity,
		},
	}

	for desc, tc := range cases {
		err := tc.user.Validate()
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
	}
}
