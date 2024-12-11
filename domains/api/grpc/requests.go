// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"github.com/absmach/supermq/pkg/apiutil"
)

type deleteUserPoliciesReq struct {
	ID string
}

func (req deleteUserPoliciesReq) validate() error {
	if req.ID == "" {
		return apiutil.ErrMissingID
	}

	return nil
}

type retrieveDomainStatusReq struct {
	ID string
}

func (req retrieveDomainStatusReq) validate() error {
	if req.ID == "" {
		return apiutil.ErrMissingID
	}

	return nil
}