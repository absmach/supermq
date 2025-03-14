// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"

	"github.com/absmach/supermq/auditlogs"
	"github.com/absmach/supermq/pkg/authn"
	smqauthz "github.com/absmach/supermq/pkg/authz"
	"github.com/absmach/supermq/pkg/policies"
)

type authorizationMiddleware struct {
	svc   auditlogs.Service
	authz smqauthz.Authorization
}

func AuthorizationMiddleware(svc auditlogs.Service, authz smqauthz.Authorization) auditlogs.Service {
	return &authorizationMiddleware{
		svc:   svc,
		authz: authz,
	}
}

func (m *authorizationMiddleware) Save(ctx context.Context, log auditlogs.AuditLog) error {
	return m.svc.Save(ctx, log)
}

func (m *authorizationMiddleware) RetrieveByID(ctx context.Context, session authn.Session,id string) (auditlogs.AuditLog, error) {
	if err := m.checkSuperAdmin(ctx, session.UserID); err != nil {
		return auditlogs.AuditLog{}, err
	}

	return m.svc.RetrieveByID(ctx,session, id)
}

func (m *authorizationMiddleware) RetrieveAll(ctx context.Context, session authn.Session, pm auditlogs.Page) (auditlogs.AuditLogPage, error) {
	if err := m.checkSuperAdmin(ctx, session.UserID); err != nil {
		return auditlogs.AuditLogPage{}, err
	}

	return m.svc.RetrieveAll(ctx, session, pm)
}


func (am *authorizationMiddleware) checkSuperAdmin(ctx context.Context, userID string) error {
	if err := am.authz.Authorize(ctx, smqauthz.PolicyReq{
		SubjectType: policies.UserType,
		Subject:     userID,
		Permission:  policies.AdminPermission,
		ObjectType:  policies.PlatformType,
		Object:      policies.SuperMQObject,
	}); err != nil {
		return err
	}
	return nil
}
