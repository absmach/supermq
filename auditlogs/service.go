// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package auditlogs

import (
	"context"

	"github.com/absmach/supermq"
	"github.com/absmach/supermq/pkg/authn"
	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
)

type service struct {
	idProvider supermq.IDProvider
	repository Repository
}

func NewService(idp supermq.IDProvider, repository Repository) Service {
	return &service{
		idProvider: idp,
		repository: repository,
	}
}

func (svc *service) Save(ctx context.Context, log AuditLog) error {
	id, err := svc.idProvider.ID()
	if err != nil {
		return err
	}
	log.ID = id

	if err := svc.repository.Save(ctx, log); err != nil {
		return errors.Wrap(svcerr.ErrCreateEntity, err)
	}

	return nil
}

func (svc *service) RetrieveByID(ctx context.Context, session authn.Session, id string) (AuditLog, error) {
	log, err := svc.repository.RetrieveByID(ctx, id)
	if err != nil {
		return AuditLog{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}

	return log, nil
}

func (svc *service) RetrieveAll(ctx context.Context, session authn.Session, pm Page) (AuditLogPage, error) {
	page, err := svc.repository.RetrieveAll(ctx, pm)
	if err != nil {
		return AuditLogPage{}, errors.Wrap(svcerr.ErrViewEntity, err)
	}

	return page, nil
}
