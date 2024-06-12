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
	if err := svc.authorize(ctx, token, page.EntityID, page.EntityType.AuthString()); err != nil {
		return ActivitiesPage{}, err
	}

	return svc.repository.RetrieveAll(ctx, page)
}

func (svc *service) authorize(ctx context.Context, token, entityID, entityType string) error {
	permission := auth.ViewPermission
	objectType := entityType
	object := entityID
	// If the entity is a user, we need to check if the user is an admin
	if entityType == auth.UserType {
		permission = auth.AdminPermission
		objectType = auth.PlatformType
		object = auth.MagistralaObject
	}

	req := &magistrala.AuthorizeReq{
		SubjectType: auth.UserType,
		SubjectKind: auth.TokenKind,
		Subject:     token,
		Permission:  permission,
		ObjectType:  objectType,
		Object:      object,
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
