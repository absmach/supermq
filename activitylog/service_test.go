// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package activitylog_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/absmach/magistrala"
	"github.com/absmach/magistrala/activitylog"
	"github.com/absmach/magistrala/activitylog/mocks"
	"github.com/absmach/magistrala/auth"
	authmocks "github.com/absmach/magistrala/auth/mocks"
	"github.com/absmach/magistrala/internal/testsutil"
	"github.com/absmach/magistrala/pkg/errors"
	repoerr "github.com/absmach/magistrala/pkg/errors/repository"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	"github.com/stretchr/testify/assert"
)

func TestSave(t *testing.T) {
	repo := new(mocks.Repository)
	authsvc := new(authmocks.AuthClient)
	svc := activitylog.NewService(repo, authsvc)

	cases := []struct {
		desc     string
		activity activitylog.Activity
		repoErr  error
		err      error
	}{
		{
			desc: "successful with ID and EntityType",
			activity: activitylog.Activity{
				ID:         testsutil.GenerateUUID(t),
				Operation:  "user.create",
				OccurredAt: time.Now().Add(-time.Hour),
				Payload: map[string]interface{}{
					"temperature": rand.Float64(),
					"humidity":    rand.Float64(),
				},
			},
			repoErr: nil,
			err:     nil,
		},
		{
			desc:    "with repo error",
			repoErr: repoerr.ErrCreateEntity,
			err:     repoerr.ErrCreateEntity,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			repoCall := repo.On("Save", context.Background(), tc.activity).Return(tc.repoErr)
			err := svc.Save(context.Background(), tc.activity)
			assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
			repoCall.Unset()
		})
	}
}

func TestReadAll(t *testing.T) {
	repo := new(mocks.Repository)
	authsvc := new(authmocks.AuthClient)
	svc := activitylog.NewService(repo, authsvc)

	validToken := "token"
	validPage := activitylog.Page{
		Offset:     0,
		Limit:      10,
		ID:         testsutil.GenerateUUID(t),
		EntityType: auth.UserType,
	}

	cases := []struct {
		desc    string
		token   string
		page    activitylog.Page
		resp    activitylog.ActivitiesPage
		authRes *magistrala.AuthorizeRes
		authErr error
		idRes   *magistrala.IdentityRes
		idErr   error
		repoErr error
		err     error
	}{
		{
			desc:  "successful",
			token: validToken,
			page: activitylog.Page{
				Offset: 0,
				Limit:  10,
			},
			resp: activitylog.ActivitiesPage{
				Total:  1,
				Offset: 0,
				Limit:  10,
				Activities: []activitylog.Activity{
					{
						ID:         testsutil.GenerateUUID(t),
						Operation:  "user.create",
						OccurredAt: time.Now().Add(-time.Hour),
						Payload: map[string]interface{}{
							"temperature": rand.Float64(),
							"humidity":    rand.Float64(),
							"sensor_id":   rand.Intn(1000),
						},
					},
				},
			},
			idRes:   &magistrala.IdentityRes{UserId: testsutil.GenerateUUID(t)},
			idErr:   nil,
			repoErr: nil,
			err:     nil,
		},
		{
			desc:  "successful with ID and EntityType",
			token: validToken,
			page:  validPage,
			resp: activitylog.ActivitiesPage{
				Total:  1,
				Offset: 0,
				Limit:  10,
				Activities: []activitylog.Activity{
					{
						ID:         testsutil.GenerateUUID(t),
						Operation:  "user.create",
						OccurredAt: time.Now().Add(-time.Hour),
						Payload: map[string]interface{}{
							"temperature": rand.Float64(),
							"humidity":    rand.Float64(),
							"sensor_id":   rand.Intn(1000),
						},
					},
				},
			},
			authRes: &magistrala.AuthorizeRes{Authorized: true},
			authErr: nil,
			repoErr: nil,
			err:     nil,
		},
		{
			desc:  "invalid token",
			token: "invalid",
			page: activitylog.Page{
				Offset: 0,
				Limit:  10,
			},
			idRes: &magistrala.IdentityRes{},
			idErr: svcerr.ErrAuthentication,
			err:   svcerr.ErrAuthentication,
		},
		{
			desc:  "invalid token with no identity error",
			token: validToken,
			page: activitylog.Page{
				Offset: 0,
				Limit:  10,
			},
			idRes: &magistrala.IdentityRes{},
			idErr: nil,
			err:   svcerr.ErrAuthentication,
		},
		{
			desc:    "invalid token",
			token:   "invalid",
			page:    validPage,
			authRes: &magistrala.AuthorizeRes{Authorized: false},
			authErr: svcerr.ErrAuthorization,
			err:     svcerr.ErrAuthorization,
		},
		{
			desc:    "with repo error",
			token:   validToken,
			page:    validPage,
			resp:    activitylog.ActivitiesPage{},
			authRes: &magistrala.AuthorizeRes{Authorized: true},
			authErr: nil,
			repoErr: repoerr.ErrViewEntity,
			err:     repoerr.ErrViewEntity,
		},
		{
			desc:    "with failed to authorize",
			token:   validToken,
			page:    validPage,
			resp:    activitylog.ActivitiesPage{},
			authRes: &magistrala.AuthorizeRes{Authorized: false},
			authErr: nil,
			repoErr: nil,
			err:     svcerr.ErrAuthorization,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			authReq := &magistrala.AuthorizeReq{
				SubjectType: auth.UserType,
				SubjectKind: auth.TokenKind,
				Subject:     tc.token,
				Permission:  auth.ViewPermission,
				ObjectType:  tc.page.EntityType,
				Object:      tc.page.ID,
			}
			idReq := &magistrala.IdentityReq{Token: tc.token}
			authCall := authsvc.On("Identify", context.Background(), idReq).Return(tc.idRes, tc.idErr)
			authCall1 := authsvc.On("Authorize", context.Background(), authReq).Return(tc.authRes, tc.authErr)
			repoCall := repo.On("RetrieveAll", context.Background(), tc.page).Return(tc.resp, tc.repoErr)
			resp, err := svc.ReadAll(context.Background(), tc.token, tc.page)
			if tc.err == nil {
				assert.Equal(t, tc.resp, resp, tc.desc)
			}
			assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
			repoCall.Unset()
			authCall.Unset()
			authCall1.Unset()
		})
	}
}
