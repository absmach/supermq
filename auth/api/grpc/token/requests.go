// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package token

import (
	apiutil "github.com/absmach/supermq/api/http/util"
	"github.com/absmach/supermq/auth"
)

type issueReq struct {
	userID   string
	userRole auth.Role
	keyType  auth.KeyType
}

func (req issueReq) validate() error {
	if req.keyType != auth.AccessKey &&
		req.keyType != auth.APIKey &&
		req.keyType != auth.RecoveryKey &&
		req.keyType != auth.InvitationKey {
		return apiutil.ErrInvalidAuthKey
	}

	return nil
}

type refreshReq struct {
	refreshToken string
}

func (req refreshReq) validate() error {
	if req.refreshToken == "" {
		return apiutil.ErrMissingSecret
	}

	return nil
}
