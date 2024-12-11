// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package pats

import (
	"net/http"

	"github.com/absmach/supermq"
	"github.com/absmach/supermq/auth"
)

var (
	_ supermq.Response = (*createPatRes)(nil)
	_ supermq.Response = (*retrievePatRes)(nil)
	_ supermq.Response = (*updatePatNameRes)(nil)
	_ supermq.Response = (*updatePatDescriptionRes)(nil)
	_ supermq.Response = (*deletePatRes)(nil)
	_ supermq.Response = (*resetPatSecretRes)(nil)
	_ supermq.Response = (*revokePatSecretRes)(nil)
	_ supermq.Response = (*addPatScopeEntryRes)(nil)
	_ supermq.Response = (*removePatScopeEntryRes)(nil)
	_ supermq.Response = (*clearAllScopeEntryRes)(nil)
)

type createPatRes struct {
	auth.PAT
}

func (res createPatRes) Code() int {
	return http.StatusCreated
}

func (res createPatRes) Headers() map[string]string {
	return map[string]string{}
}

func (res createPatRes) Empty() bool {
	return false
}

type retrievePatRes struct {
	auth.PAT
}

func (res retrievePatRes) Code() int {
	return http.StatusOK
}

func (res retrievePatRes) Headers() map[string]string {
	return map[string]string{}
}

func (res retrievePatRes) Empty() bool {
	return false
}

type updatePatNameRes struct {
	auth.PAT
}

func (res updatePatNameRes) Code() int {
	return http.StatusAccepted
}

func (res updatePatNameRes) Headers() map[string]string {
	return map[string]string{}
}

func (res updatePatNameRes) Empty() bool {
	return false
}

type updatePatDescriptionRes struct {
	auth.PAT
}

func (res updatePatDescriptionRes) Code() int {
	return http.StatusAccepted
}

func (res updatePatDescriptionRes) Headers() map[string]string {
	return map[string]string{}
}

func (res updatePatDescriptionRes) Empty() bool {
	return false
}

type listPatsRes struct {
	auth.PATSPage
}

func (res listPatsRes) Code() int {
	return http.StatusOK
}

func (res listPatsRes) Headers() map[string]string {
	return map[string]string{}
}

func (res listPatsRes) Empty() bool {
	return false
}

type deletePatRes struct{}

func (res deletePatRes) Code() int {
	return http.StatusNoContent
}

func (res deletePatRes) Headers() map[string]string {
	return map[string]string{}
}

func (res deletePatRes) Empty() bool {
	return true
}

type resetPatSecretRes struct {
	auth.PAT
}

func (res resetPatSecretRes) Code() int {
	return http.StatusOK
}

func (res resetPatSecretRes) Headers() map[string]string {
	return map[string]string{}
}

func (res resetPatSecretRes) Empty() bool {
	return false
}

type revokePatSecretRes struct{}

func (res revokePatSecretRes) Code() int {
	return http.StatusNoContent
}

func (res revokePatSecretRes) Headers() map[string]string {
	return map[string]string{}
}

func (res revokePatSecretRes) Empty() bool {
	return true
}

type addPatScopeEntryRes struct {
	auth.Scope
}

func (res addPatScopeEntryRes) Code() int {
	return http.StatusOK
}

func (res addPatScopeEntryRes) Headers() map[string]string {
	return map[string]string{}
}

func (res addPatScopeEntryRes) Empty() bool {
	return false
}

type removePatScopeEntryRes struct {
	auth.Scope
}

func (res removePatScopeEntryRes) Code() int {
	return http.StatusOK
}

func (res removePatScopeEntryRes) Headers() map[string]string {
	return map[string]string{}
}

func (res removePatScopeEntryRes) Empty() bool {
	return false
}

type clearAllScopeEntryRes struct{}

func (res clearAllScopeEntryRes) Code() int {
	return http.StatusOK
}

func (res clearAllScopeEntryRes) Headers() map[string]string {
	return map[string]string{}
}

func (res clearAllScopeEntryRes) Empty() bool {
	return true
}
