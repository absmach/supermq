// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"net/http"

	"github.com/absmach/supermq/auditlogs"
)

type retrieveAuditLogRes struct {
	auditlogs.AuditLog
}

func (res retrieveAuditLogRes) Code() int {
	return http.StatusOK
}

func (res retrieveAuditLogRes) Headers() map[string]string {
	return map[string]string{}
}

func (res retrieveAuditLogRes) Empty() bool {
	return false
}

type retrieveAllAuditLogsRes struct {
	auditlogs.AuditLogPage
}

func (res retrieveAllAuditLogsRes) Code() int {
	return http.StatusOK
}

func (res retrieveAllAuditLogsRes) Headers() map[string]string {
	return map[string]string{}
}

func (res retrieveAllAuditLogsRes) Empty() bool {
	return false
}
