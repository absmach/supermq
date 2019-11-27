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
	"time"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/twins"
)

var (
	_ mainflux.Response = (*twinRes)(nil)
	_ mainflux.Response = (*viewTwinRes)(nil)
	_ mainflux.Response = (*twinsSetRes)(nil)
	_ mainflux.Response = (*removeRes)(nil)
)

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

type viewTwinRes struct {
	Owner       string                 `json:"owner,omitempty"`
	ID          string                 `json:"id"`
	Key         string                 `json:"key"`
	ThingID     string                 `json:"thingID"`
	Name        string                 `json:"name,omitempty"`
	Revision    int                    `json:"revision"`
	Created     time.Time              `json:"created"`
	Updated     time.Time              `json:"updated"`
	Definitions []twins.Definition     `json:"definitions,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

func (res viewTwinRes) Code() int {
	return http.StatusOK
}

func (res viewTwinRes) Headers() map[string]string {
	return map[string]string{}
}

func (res viewTwinRes) Empty() bool {
	return false
}

type setRes struct {
	Total uint64 `json:"total"`
	Limit uint64 `json:"limit"`
}

type twinsSetRes struct {
	setRes
	Twins []viewTwinRes `json:"twins"`
}

func (res twinsSetRes) Code() int {
	return http.StatusOK
}

func (res twinsSetRes) Headers() map[string]string {
	return map[string]string{}
}

func (res twinsSetRes) Empty() bool {
	return false
}

type removeRes struct{}

func (res removeRes) Code() int {
	return http.StatusNoContent
}

func (res removeRes) Headers() map[string]string {
	return map[string]string{}
}

func (res removeRes) Empty() bool {
	return true
}
