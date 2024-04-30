// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package activitylog

import (
	"context"

	"github.com/absmach/magistrala"
	"github.com/absmach/magistrala/pkg/errors"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
)

type service struct {
	auth       magistrala.AuthServiceClient
	repository Repository
}

func NewService(repository Repository, authClient magistrala.AuthServiceClient) Service {
	return &service{
		auth:       authClient,
		repository: repository,
	}
}

func (svc *service) Save(ctx context.Context, activity Activity) error {
	return svc.repository.Save(ctx, activity)
}

func (svc *service) ReadAll(ctx context.Context, token string, page Page) (ActivitiesPage, error) {
	if err := svc.identify(ctx, token); err != nil {
		return ActivitiesPage{}, err
	}

	return svc.repository.RetrieveAll(ctx, page)
}

func (svc *service) identify(ctx context.Context, token string) error {
	user, err := svc.auth.Identify(ctx, &magistrala.IdentityReq{Token: token})
	if err != nil {
		return errors.Wrap(svcerr.ErrAuthentication, err)
	}
	if user.GetUserId() == "" {
		return svcerr.ErrAuthentication
	}

	return nil
}
