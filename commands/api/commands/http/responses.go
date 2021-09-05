// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"net/http"

	"github.com/mainflux/mainflux"
)

var _ mainflux.Response = (*pingRes)(nil)
var _ mainflux.Response = (*getRes)(nil)

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

type getRes struct {
	Greeting string `json:"greeting"`
}

func (res getRes) Code() int {
	return http.StatusOK
}

func (res getRes) Headers() map[string]string {
	return map[string]string{}
}

func (res getRes) Empty() bool {
	return false
}
