// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package auth

import smqauth "github.com/absmach/supermq/auth"

type authenticateRes struct {
	id       string
	userID   string
	userRole smqauth.Role
}

type authorizeRes struct {
	id         string
	authorized bool
}
