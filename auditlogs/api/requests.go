// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import "github.com/absmach/supermq/auditlogs"

type retriveAuditLogReq struct {
	id string
}

type retrieveAllAuditLogsReq struct {
	auditlogs.Page
}
