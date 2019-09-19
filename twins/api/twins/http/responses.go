//
// Copyright (c) 2019
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package http

import (
	"fmt"
	"net/http"

	"github.com/mainflux/mainflux"
)

var (
	_ mainflux.Response = (*pingRes)(nil)
	_ mainflux.Response = (*twinRes)(nil)
)

type pingRes struct {
	Greeting string `json:"greeting"`
}

func (res pingRes) Code() int {
	return http.StatusOK
}

func (res pingRes) Headers() map[string]string {
	return map[string]string{}
}

func (res pingRes) Empty() bool {
	return false
}

type twinRes struct {
	id      string
	created bool
}

func (res twinRes) Code() int {
	if res.created {
		return http.StatusCreated
	}

	return http.StatusOK
}

func (res twinRes) Headers() map[string]string {
	if res.created {
		return map[string]string{
			"Location": fmt.Sprintf("/twins/%s", res.id),
		}
	}

	return map[string]string{}
}

func (res twinRes) Empty() bool {
	return true
}
