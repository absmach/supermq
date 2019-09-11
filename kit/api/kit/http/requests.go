//
// Copyright (c) 2019
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package http

import "github.com/mainflux/mainflux/kit"

type apiReq interface {
	validate() error
}

type pingReq struct {
	Secret string
}

func (req pingReq) validate() error {
	if req.Secret == "" {
		return kit.ErrUnauthorizedAccess
	}

	return nil
}
