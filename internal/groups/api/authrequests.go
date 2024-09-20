// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"

	"github.com/absmach/magistrala"
	"github.com/absmach/magistrala/internal/api"
	"github.com/absmach/magistrala/pkg/apiutil"
	"github.com/absmach/magistrala/pkg/auth"
	"github.com/absmach/magistrala/pkg/errors"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	"github.com/absmach/magistrala/pkg/policies"
)

func CreateGroupAuthReq(ctx context.Context, request interface{}) ([]*magistrala.AuthorizeReq, error) {
	req := request.(createGroupReq)
	if err := req.validate(); err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, err)
	}

	session, ok := ctx.Value(api.SessionKey).(auth.Session)
	if !ok {
		return nil, svcerr.ErrAuthorization
	}

	prs := []*magistrala.AuthorizeReq{
		{
			SubjectType: policies.UserType,
			SubjectKind: policies.UsersKind,
			Subject:     session.DomainUserID,
			Permission:  policies.CreatePermission,
			ObjectType:  policies.DomainType,
			Object:      session.DomainID,
		},
	}

	if req.Group.Parent != "" {
		prs = append(prs, &magistrala.AuthorizeReq{
			SubjectType: policies.UserType,
			SubjectKind: policies.TokenKind,
			Subject:     req.token,
			Permission:  policies.EditPermission,
			ObjectType:  policies.GroupType,
			Object:      req.Group.Parent,
		})
	}

	return prs, nil
}

func ViewGroupAuthReq(_ context.Context, request interface{}) ([]*magistrala.AuthorizeReq, error) {
	req := request.(groupReq)
	if err := req.validate(); err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, err)
	}

	prs := []*magistrala.AuthorizeReq{
		{
			SubjectType: policies.UserType,
			SubjectKind: policies.TokenKind,
			Subject:     req.token,
			Permission:  policies.ViewPermission,
			ObjectType:  policies.GroupType,
			Object:      req.id,
		},
	}

	return prs, nil
}

func UpdateGroupAuthReq(_ context.Context, request interface{}) ([]*magistrala.AuthorizeReq, error) {
	req := request.(updateGroupReq)
	if err := req.validate(); err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, err)
	}

	prs := []*magistrala.AuthorizeReq{
		{
			SubjectType: policies.UserType,
			SubjectKind: policies.TokenKind,
			Subject:     req.token,
			Permission:  policies.EditPermission,
			ObjectType:  policies.GroupType,
			Object:      req.id,
		},
	}

	return prs, nil
}

func ChangeGroupStatusAuthReq(_ context.Context, request interface{}) ([]*magistrala.AuthorizeReq, error) {
	req := request.(changeGroupStatusReq)
	if err := req.validate(); err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, err)
	}

	prs := []*magistrala.AuthorizeReq{
		{
			SubjectType: policies.UserType,
			SubjectKind: policies.TokenKind,
			Subject:     req.token,
			Permission:  policies.EditPermission,
			ObjectType:  policies.GroupType,
			Object:      req.id,
		},
	}

	return prs, nil
}

func ListGroupsByUserAuthReq(ctx context.Context, request interface{}) ([]*magistrala.AuthorizeReq, error) {
	req := request.(listGroupsReq)
	req.memberKind = policies.UsersKind
	if err := req.validate(); err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, err)
	}

	session, ok := ctx.Value(api.SessionKey).(auth.Session)
	if !ok {
		return nil, svcerr.ErrAuthorization
	}
	prs := []*magistrala.AuthorizeReq{}
	if req.memberID != "" && session.UserID != req.memberID {
		prs = append(prs, &magistrala.AuthorizeReq{
			Domain:      session.DomainID,
			SubjectType: policies.UserType,
			SubjectKind: policies.UsersKind,
			Subject:     session.DomainUserID,
			Permission:  policies.AdminPermission,
			ObjectType:  policies.DomainType,
			Object:      session.DomainID,
		})
		return prs, nil
	}
	if !session.SuperAdmin {
		prs = append(prs, &magistrala.AuthorizeReq{
			SubjectType: policies.UserType,
			SubjectKind: policies.UsersKind,
			Subject:     session.DomainUserID,
			Permission:  policies.MembershipPermission,
			ObjectType:  policies.DomainType,
			Object:      session.DomainID,
		})
	}

	return prs, nil
}

func ListGroupsByThingAuthReq(ctx context.Context, request interface{}) ([]*magistrala.AuthorizeReq, error) {
	req := request.(listGroupsReq)
	req.memberKind = policies.ThingsKind
	if err := req.validate(); err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, err)
	}

	session, ok := ctx.Value(api.SessionKey).(auth.Session)
	if !ok {
		return nil, svcerr.ErrAuthorization
	}

	prs := []*magistrala.AuthorizeReq{
		{
			Domain:      session.DomainID,
			SubjectType: policies.UserType,
			SubjectKind: policies.UsersKind,
			Subject:     session.DomainUserID,
			Permission:  policies.ViewPermission,
			ObjectType:  policies.ThingType,
			Object:      req.memberID,
		},
	}

	return prs, nil
}

func ListGroupsByGroupAuthReq(ctx context.Context, request interface{}) ([]*magistrala.AuthorizeReq, error) {
	req := request.(listGroupsReq)
	req.memberKind = policies.GroupsKind
	if err := req.validate(); err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, err)
	}

	session, ok := ctx.Value(api.SessionKey).(auth.Session)
	if !ok {
		return nil, svcerr.ErrAuthorization
	}

	prs := []*magistrala.AuthorizeReq{
		{
			Domain:      session.DomainID,
			SubjectType: policies.UserType,
			SubjectKind: policies.UsersKind,
			Subject:     session.DomainUserID,
			Permission:  req.Page.Permission,
			ObjectType:  policies.GroupType,
			Object:      req.memberID,
		},
	}

	return prs, nil
}

func ListGroupsByChannelAuthReq(ctx context.Context, request interface{}) ([]*magistrala.AuthorizeReq, error) {
	req := request.(listGroupsReq)
	req.memberKind = policies.ChannelsKind
	if err := req.validate(); err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, err)
	}

	session, ok := ctx.Value(api.SessionKey).(auth.Session)
	if !ok {
		return nil, svcerr.ErrAuthorization
	}

	prs := []*magistrala.AuthorizeReq{
		{
			Domain:      session.DomainID,
			SubjectType: policies.UserType,
			SubjectKind: policies.UsersKind,
			Subject:     session.DomainUserID,
			Permission:  policies.ViewPermission,
			ObjectType:  policies.GroupType,
			Object:      req.memberID,
		},
	}

	return prs, nil
}

func ListMembersAuthReq(_ context.Context, request interface{}) ([]*magistrala.AuthorizeReq, error) {
	req := request.(listMembersReq)

	prs := []*magistrala.AuthorizeReq{
		{
			SubjectType: policies.UserType,
			SubjectKind: policies.TokenKind,
			Subject:     req.token,
			Permission:  policies.ViewPermission,
			ObjectType:  policies.GroupType,
			Object:      req.groupID,
		},
	}

	return prs, nil
}

func DeleteGroupAuthReq(ctx context.Context, request interface{}) ([]*magistrala.AuthorizeReq, error) {
	req := request.(groupReq)
	if err := req.validate(); err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, err)
	}

	session, ok := ctx.Value(api.SessionKey).(auth.Session)
	if !ok {
		return nil, svcerr.ErrAuthorization
	}

	prs := []*magistrala.AuthorizeReq{
		{
			Domain:      session.DomainID,
			SubjectType: policies.UserType,
			SubjectKind: policies.UsersKind,
			Subject:     session.DomainUserID,
			Permission:  policies.DeletePermission,
			ObjectType:  policies.GroupType,
			Object:      req.id,
		},
	}

	return prs, nil
}
