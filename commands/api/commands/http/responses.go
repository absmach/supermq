// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"net/http"

	"github.com/mainflux/mainflux"
)

var _ mainflux.Response = (*viewCommandsRes)(nil)
var _ mainflux.Response = (*listCommandsRes)(nil)

type viewCommandsRes struct {
	Greeting string `json:"greeting"`
}

func (res viewCommandsRes) Code() int {
	return http.StatusOK
}

func (res viewCommandsRes) Headers() map[string]string {
	return map[string]string{}
}

func (res viewCommandsRes) Empty() bool {
	return false
}

type listCommandsRes struct {
	Greeting string `json:"greeting"`
}

func (res listCommandsRes) Code() int {
	return http.StatusOK
}

func (res listCommandsRes) Headers() map[string]string {
	return map[string]string{}
}

func (res listCommandsRes) Empty() bool {
	return false
}
