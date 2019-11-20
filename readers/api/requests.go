// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

type apiReq interface {
	validate() error
}

type listMessagesReq struct {
	token  string
	chanID string
	offset uint64
	limit  uint64
	query  map[string]string
}

func (req listMessagesReq) validate() error {
	if req.limit < 1 {
		return errInvalidRequest
	}

	return nil
}
