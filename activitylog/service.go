// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package activitylog

import (
	"context"

	"github.com/absmach/magistrala"
	"github.com/absmach/magistrala/auth"
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

func (svc *service) RetrieveAll(ctx context.Context, token string, page Page) (ActivitiesPage, error) {
	userID, domainID, err := svc.identify(ctx, token)
	if err != nil {
		return ActivitiesPage{}, err
	}
	if page.EntityID != "" {
		if err := svc.authorize(ctx, userID, domainID, auth.ViewPermission, page.EntityID, page.EntityType.AuthString()); err != nil {
			return ActivitiesPage{}, err
		}

		return svc.repository.RetrieveAll(ctx, page)
	}

	if err := svc.authorize(ctx, userID, domainID, auth.AdminPermission, auth.PlatformType, auth.MagistralaObject); err != nil {
		page.WithAttributes = false
		page.WithMetadata = false
	}

	return svc.repository.RetrieveAll(ctx, page)
}

func (svc *service) identify(ctx context.Context, token string) (string, string, error) {
	user, err := svc.auth.Identify(ctx, &magistrala.IdentityReq{Token: token})
	if err != nil {
		return "", "", errors.Wrap(svcerr.ErrAuthentication, err)
	}
	if user.GetUserId() == "" {
		return "", "", svcerr.ErrAuthentication
	}

	return user.GetUserId(), user.GetDomainId(), nil
}

func (svc *service) authorize(ctx context.Context, userID, domainID, permission, entityID, entityType string) error {
	req := &magistrala.AuthorizeReq{
		Domain:      domainID,
		SubjectType: auth.UserType,
		SubjectKind: auth.UsersKind,
		Subject:     userID,
		Permission:  permission,
		ObjectType:  entityType,
		Object:      entityID,
	}

	res, err := svc.auth.Authorize(ctx, req)
	if err != nil {
		return errors.Wrap(svcerr.ErrAuthorization, err)
	}
	if !res.GetAuthorized() {
		return svcerr.ErrAuthorization
	}

	return nil
}
