// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package sdk

// UserPasswordReq contains old and new passwords
type UserPasswordReq struct {
	OldPassword string `json:"old_password,omitempty"`
	Password    string `json:"password,omitempty"`
}
