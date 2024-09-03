// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"net/http"

	"github.com/absmach/magistrala"
	"github.com/absmach/magistrala/pkg/roles"
)

var (
	_ magistrala.Response = (*createRoleRes)(nil)
	_ magistrala.Response = (*listRolesRes)(nil)
	_ magistrala.Response = (*viewRoleRes)(nil)
	_ magistrala.Response = (*updateRoleRes)(nil)
	_ magistrala.Response = (*deleteRoleRes)(nil)
	_ magistrala.Response = (*addRoleOperationsRes)(nil)
	_ magistrala.Response = (*listRoleOperationsRes)(nil)
	_ magistrala.Response = (*deleteRoleOperationsRes)(nil)
	_ magistrala.Response = (*deleteAllRoleOperationsRes)(nil)
	_ magistrala.Response = (*addRoleMembersRes)(nil)
	_ magistrala.Response = (*listRoleMembersRes)(nil)
	_ magistrala.Response = (*deleteRoleMembersRes)(nil)
	_ magistrala.Response = (*deleteAllRoleMemberRes)(nil)
)

type createRoleRes struct {
	roles.Role
}

func (res createRoleRes) Code() int {
	return http.StatusCreated
}

func (res createRoleRes) Headers() map[string]string {
	return map[string]string{}
}

func (res createRoleRes) Empty() bool {
	return false
}

type listRolesRes struct {
	roles.RolePage
}

func (res listRolesRes) Code() int {
	return http.StatusOK
}

func (res listRolesRes) Headers() map[string]string {
	return map[string]string{}
}

func (res listRolesRes) Empty() bool {
	return false
}

type viewRoleRes struct {
	roles.Role
}

func (res viewRoleRes) Code() int {
	return http.StatusOK
}

func (res viewRoleRes) Headers() map[string]string {
	return map[string]string{}
}

func (res viewRoleRes) Empty() bool {
	return false
}

type updateRoleRes struct {
	roles.Role
}

func (res updateRoleRes) Code() int {
	return http.StatusOK
}

func (res updateRoleRes) Headers() map[string]string {
	return map[string]string{}
}

func (res updateRoleRes) Empty() bool {
	return false
}

type deleteRoleRes struct {
}

func (res deleteRoleRes) Code() int {
	return http.StatusNoContent
}

func (res deleteRoleRes) Headers() map[string]string {
	return map[string]string{}
}

func (res deleteRoleRes) Empty() bool {
	return true
}

type addRoleOperationsRes struct {
	Operations []roles.Operation `json:"operations"`
}

func (res addRoleOperationsRes) Code() int {
	return http.StatusOK
}

func (res addRoleOperationsRes) Headers() map[string]string {
	return map[string]string{}
}

func (res addRoleOperationsRes) Empty() bool {
	return false
}

type listRoleOperationsRes struct {
	Operations []roles.Operation `json:"operations"`
}

func (res listRoleOperationsRes) Code() int {
	return http.StatusOK
}

func (res listRoleOperationsRes) Headers() map[string]string {
	return map[string]string{}
}

func (res listRoleOperationsRes) Empty() bool {
	return false
}

type deleteRoleOperationsRes struct{}

func (res deleteRoleOperationsRes) Code() int {
	return http.StatusNoContent
}

func (res deleteRoleOperationsRes) Headers() map[string]string {
	return map[string]string{}
}

func (res deleteRoleOperationsRes) Empty() bool {
	return true
}

type deleteAllRoleOperationsRes struct{}

func (res deleteAllRoleOperationsRes) Code() int {
	return http.StatusNoContent
}

func (res deleteAllRoleOperationsRes) Headers() map[string]string {
	return map[string]string{}
}

func (res deleteAllRoleOperationsRes) Empty() bool {
	return true
}

type addRoleMembersRes struct {
	Members []string `json:"members"`
}

func (res addRoleMembersRes) Code() int {
	return http.StatusOK
}

func (res addRoleMembersRes) Headers() map[string]string {
	return map[string]string{}
}

func (res addRoleMembersRes) Empty() bool {
	return false
}

type listRoleMembersRes struct {
	roles.MembersPage
}

func (res listRoleMembersRes) Code() int {
	return http.StatusOK
}

func (res listRoleMembersRes) Headers() map[string]string {
	return map[string]string{}
}

func (res listRoleMembersRes) Empty() bool {
	return false
}

type deleteRoleMembersRes struct{}

func (res deleteRoleMembersRes) Code() int {
	return http.StatusNoContent
}

func (res deleteRoleMembersRes) Headers() map[string]string {
	return map[string]string{}
}

func (res deleteRoleMembersRes) Empty() bool {
	return true
}

type deleteAllRoleMemberRes struct{}

func (res deleteAllRoleMemberRes) Code() int {
	return http.StatusNoContent
}

func (res deleteAllRoleMemberRes) Headers() map[string]string {
	return map[string]string{}
}

func (res deleteAllRoleMemberRes) Empty() bool {
	return true
}
