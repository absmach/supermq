// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package sdk_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/logger"
	sdk "github.com/mainflux/mainflux/pkg/sdk/go"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/mainflux/mainflux/users"
	"github.com/mainflux/mainflux/users/api"
	"github.com/mainflux/mainflux/users/mocks"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	invalidEmail = "userexample.com"
)

var (
	passRegex = regexp.MustCompile("^.{8,}$")
	// Limit query parameter
	limit uint64 = 5
	// Offset query parameter
	offset uint64 = 0
)

func newUserService() users.Service {
	usersRepo := mocks.NewUserRepository()
	hasher := mocks.NewHasher()
	userEmail := "user@example.com"

	mockAuthzDB := map[string][]mocks.SubjectSet{}
	mockAuthzDB[userEmail] = append(mockAuthzDB[userEmail], mocks.SubjectSet{Object: "authorities", Relation: "member"})
	auth := mocks.NewAuthService(map[string]string{userEmail: userEmail}, mockAuthzDB)

	emailer := mocks.NewEmailer()
	idProvider := uuid.New()

	return users.New(usersRepo, hasher, auth, emailer, idProvider, passRegex)
}

func newUserServer(svc users.Service) *httptest.Server {
	logger := logger.NewMock()
	mux := api.MakeHandler(svc, mocktracer.New(), logger)
	return httptest.NewServer(mux)
}

func TestCreateUser(t *testing.T) {
	svc := newUserService()
	ts := newUserServer(svc)
	defer ts.Close()
	sdkConf := sdk.Config{
		UsersURL:        ts.URL,
		MsgContentType:  contentType,
		TLSVerification: false,
	}

	user := sdk.User{Email: "user@example.com", Password: "password"}

	mockAuthzDB := map[string][]mocks.SubjectSet{}
	mockAuthzDB[user.Email] = append(mockAuthzDB[user.Email], mocks.SubjectSet{Object: "authorities", Relation: "member"})
	auth := mocks.NewAuthService(map[string]string{user.Email: user.Email}, mockAuthzDB)

	tkn, _ := auth.Issue(context.Background(), &mainflux.IssueReq{Id: user.ID, Email: user.Email, Type: 0})
	token := tkn.GetValue()

	mainfluxSDK := sdk.NewSDK(sdkConf)
	cases := []struct {
		desc  string
		user  sdk.User
		token string
		err   error
	}{
		{
			desc:  "register new user",
			user:  user,
			token: token,
			err:   nil,
		},
		{
			desc:  "register existing user",
			user:  user,
			token: token,
			err:   createError(sdk.ErrFailedCreation, http.StatusConflict),
		},
		{
			desc:  "register user with invalid email address",
			user:  sdk.User{Email: invalidEmail, Password: "password"},
			token: token,
			err:   createError(sdk.ErrFailedCreation, http.StatusBadRequest),
		},
		{
			desc:  "register user with empty password",
			user:  sdk.User{Email: "user2@example.com", Password: ""},
			token: token,
			err:   createError(sdk.ErrFailedCreation, http.StatusBadRequest),
		},
		{
			desc:  "register user without password",
			user:  sdk.User{Email: "user2@example.com"},
			token: token,
			err:   createError(sdk.ErrFailedCreation, http.StatusBadRequest),
		},
		{
			desc:  "register user without email",
			user:  sdk.User{Password: "password"},
			token: token,
			err:   createError(sdk.ErrFailedCreation, http.StatusBadRequest),
		},
		{
			desc:  "register empty user",
			user:  sdk.User{},
			token: token,
			err:   createError(sdk.ErrFailedCreation, http.StatusBadRequest),
		},
	}

	for _, tc := range cases {
		_, err := mainfluxSDK.CreateUser(tc.token, tc.user)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
	}
}

func TestUser(t *testing.T) {
	svc := newUserService()
	ts := newUserServer(svc)
	defer ts.Close()
	sdkConf := sdk.Config{
		UsersURL:        ts.URL,
		MsgContentType:  contentType,
		TLSVerification: false,
	}

	mainfluxSDK := sdk.NewSDK(sdkConf)
	user := sdk.User{Email: "user@example.com", Password: "password"}

	mockAuthzDB := map[string][]mocks.SubjectSet{}
	mockAuthzDB[user.Email] = append(mockAuthzDB[user.Email], mocks.SubjectSet{Object: "authorities", Relation: "member"})
	auth := mocks.NewAuthService(map[string]string{user.Email: user.Email}, mockAuthzDB)

	tkn, _ := auth.Issue(context.Background(), &mainflux.IssueReq{Id: user.ID, Email: user.Email, Type: 0})
	token := tkn.GetValue()
	userID, err := mainfluxSDK.CreateUser(token, user)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
	usertoken, err := mainfluxSDK.CreateToken(user)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
	user.ID = userID
	user.Password = ""

	cases := []struct {
		desc     string
		userID   string
		token    string
		err      error
		response sdk.User
	}{
		{
			desc:     "get existing user",
			userID:   userID,
			token:    usertoken,
			err:      nil,
			response: user,
		},
		{
			desc:     "get non-existent user",
			userID:   "43",
			token:    usertoken,
			err:      createError(sdk.ErrFailedFetch, http.StatusUnauthorized),
			response: sdk.User{},
		},

		{
			desc:     "get user with invalid token",
			userID:   userID,
			token:    wrongValue,
			err:      createError(sdk.ErrFailedFetch, http.StatusUnauthorized),
			response: sdk.User{},
		},
	}
	for _, tc := range cases {
		respUs, err := mainfluxSDK.User(tc.userID, tc.token)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
		assert.Equal(t, tc.response, respUs, fmt.Sprintf("%s: expected response user %s, got %s", tc.desc, tc.response, respUs))
	}
}

func TestUsers(t *testing.T) {
	svc := newUserService()
	ts := newUserServer(svc)
	defer ts.Close()
	sdkConf := sdk.Config{
		UsersURL:        ts.URL,
		MsgContentType:  contentType,
		TLSVerification: false,
	}

	mainfluxSDK := sdk.NewSDK(sdkConf)
	user := sdk.User{Email: "user@example.com", Password: "password"}

	mockAuthzDB := map[string][]mocks.SubjectSet{}
	mockAuthzDB[user.Email] = append(mockAuthzDB[user.Email], mocks.SubjectSet{Object: "authorities", Relation: "member"})
	auth := mocks.NewAuthService(map[string]string{user.Email: user.Email}, mockAuthzDB)

	tkn, _ := auth.Issue(context.Background(), &mainflux.IssueReq{Id: user.ID, Email: user.Email, Type: 0})
	token := tkn.GetValue()

	var users []sdk.User

	for i := 1; i < 101; i++ {
		email := fmt.Sprintf("test-%d@example.com", i)
		password := fmt.Sprintf("password%d", i)
		metadata := map[string]interface{}{"name": fmt.Sprintf("test-%d", i)}
		us := sdk.User{Email: email, Password: password, Metadata: metadata}
		userID, err := mainfluxSDK.CreateUser(token, us)
		require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
		us.ID = userID
		us.Password = ""
		users = append(users, us)
	}

	var users2 []sdk.User
	for i := 1; i < 5; i++ {
		email := fmt.Sprintf("test3-%d@mainflux.com", i)
		password := fmt.Sprintf("password%d", i)
		metadata := map[string]interface{}{"name": "mainflux"}
		us := sdk.User{Email: email, Password: password, Metadata: metadata}
		userID, err := mainfluxSDK.CreateUser(token, us)
		require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
		us.ID = userID
		us.Password = ""
		users2 = append(users2, us)
	}

	cases := []struct {
		desc     string
		token    string
		offset   uint64
		limit    uint64
		err      error
		response []sdk.User
		email    string
		metadata map[string]interface{}
	}{
		{
			desc:   "get a list users",
			token:  token,
			offset: offset,
			limit:  limit,
			err:    nil,
			email:  email,
		},
		{
			desc:   "get a list of users with invalid token",
			token:  wrongValue,
			offset: offset,
			limit:  limit,
			err:    createError(sdk.ErrFailedFetch, http.StatusUnauthorized),
			email:  email,
		},
		{
			desc:   "get a list of users with empty token",
			token:  "",
			offset: offset,
			limit:  limit,
			err:    createError(sdk.ErrFailedFetch, http.StatusUnauthorized),
			email:  email,
		},
		{
			desc:   "get a list of users with zero limit",
			token:  token,
			offset: offset,
			limit:  0,
			err:    createError(sdk.ErrFailedFetch, http.StatusBadRequest),
			email:  email,
		},
		{
			desc:   "get a list of users with limit greater than max",
			token:  token,
			offset: offset,
			limit:  110,
			err:    createError(sdk.ErrFailedFetch, http.StatusBadRequest),
			email:  email,
		},
		{
			desc:     "get a list of users with same email address and no metadata",
			token:    token,
			offset:   0,
			limit:    100,
			err:      nil,
			email:    "mainflux.com",
			metadata: make(map[string]interface{}),
		},
		{
			desc:   "get a list of users with same email address and metadata",
			token:  token,
			offset: 0,
			limit:  100,
			err:    nil,
			email:  "mainflux.com",
			metadata: map[string]interface{}{
				"name": "mainflux",
			},
		},
		{
			desc:   "get a list of users with same metadata and no email address",
			token:  token,
			offset: 0,
			limit:  100,
			err:    nil,
			email:  "",
			metadata: map[string]interface{}{
				"name": "demo5",
			},
		},
		{
			desc:     "get a list of users with same no metadata and email address",
			token:    token,
			offset:   0,
			limit:    100,
			err:      nil,
			email:    "mainflux.com",
			metadata: make(map[string]interface{}),
		},
	}
	for _, tc := range cases {
		filter := sdk.PageMetadata{
			Email:    tc.email,
			Total:    uint64(200),
			Offset:   uint64(tc.offset),
			Limit:    uint64(tc.limit),
			Metadata: tc.metadata,
		}
		_, err := mainfluxSDK.Users(tc.token, filter)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
	}
}

func TestCreateToken(t *testing.T) {
	svc := newUserService()
	ts := newUserServer(svc)
	defer ts.Close()
	sdkConf := sdk.Config{
		UsersURL:        ts.URL,
		MsgContentType:  contentType,
		TLSVerification: false,
	}

	mainfluxSDK := sdk.NewSDK(sdkConf)
	user := sdk.User{Email: "user@example.com", Password: "password"}

	mockAuthzDB := map[string][]mocks.SubjectSet{}
	mockAuthzDB[user.Email] = append(mockAuthzDB[user.Email], mocks.SubjectSet{Object: "authorities", Relation: "member"})
	auth := mocks.NewAuthService(map[string]string{user.Email: user.Email}, mockAuthzDB)

	tkn, _ := auth.Issue(context.Background(), &mainflux.IssueReq{Id: user.ID, Email: user.Email, Type: 0})
	token := tkn.GetValue()
	_, err := mainfluxSDK.CreateUser(token, user)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	cases := []struct {
		desc  string
		user  sdk.User
		token string
		err   error
	}{
		{
			desc:  "create token for user",
			user:  user,
			token: token,
			err:   nil,
		},
		{
			desc:  "create token for non existing user",
			user:  sdk.User{Email: "user2@example.com", Password: "password"},
			token: "",
			err:   createError(sdk.ErrFailedCreation, http.StatusUnauthorized),
		},
		{
			desc:  "create user with empty email",
			user:  sdk.User{Email: "", Password: "password"},
			token: "",
			err:   createError(sdk.ErrFailedCreation, http.StatusBadRequest),
		},
	}
	for _, tc := range cases {
		token, err := mainfluxSDK.CreateToken(tc.user)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
		assert.Equal(t, tc.token, token, fmt.Sprintf("%s: expected response: %s, got:  %s", tc.desc, token, tc.token))
	}
}

func TestUpdateUser(t *testing.T) {
	svc := newUserService()
	ts := newUserServer(svc)
	defer ts.Close()
	sdkConf := sdk.Config{
		UsersURL:        ts.URL,
		MsgContentType:  contentType,
		TLSVerification: false,
	}

	mainfluxSDK := sdk.NewSDK(sdkConf)
	user := sdk.User{Email: "user@example.com", Password: "password"}

	mockAuthzDB := map[string][]mocks.SubjectSet{}
	mockAuthzDB[user.Email] = append(mockAuthzDB[user.Email], mocks.SubjectSet{Object: "authorities", Relation: "member"})
	auth := mocks.NewAuthService(map[string]string{user.Email: user.Email}, mockAuthzDB)

	tkn, _ := auth.Issue(context.Background(), &mainflux.IssueReq{Id: user.ID, Email: user.Email, Type: 0})
	token := tkn.GetValue()
	userID, err := mainfluxSDK.CreateUser(token, user)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
	usertoken, err := mainfluxSDK.CreateToken(user)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	cases := []struct {
		desc  string
		user  sdk.User
		token string
		err   error
	}{
		{
			desc:  "update email for user",
			user:  sdk.User{ID: userID, Email: "user2@example.com", Password: "password"},
			token: usertoken,
			err:   nil,
		},
		{
			desc:  "update email for non existing user",
			user:  sdk.User{ID: "0", Email: "user2@example.com", Password: "password"},
			token: wrongValue,
			err:   createError(sdk.ErrFailedUpdate, http.StatusUnauthorized),
		},
		{
			desc:  "update email for user with invalid token",
			user:  sdk.User{ID: userID, Email: "user2@example.com", Password: "password"},
			token: wrongValue,
			err:   createError(sdk.ErrFailedUpdate, http.StatusUnauthorized),
		},
		{
			desc:  "update email for user with empty token",
			user:  sdk.User{ID: userID, Email: "user2@example.com", Password: "password"},
			token: "",
			err:   createError(sdk.ErrFailedUpdate, http.StatusUnauthorized),
		},
		{
			desc:  "update metadata for user",
			user:  sdk.User{ID: userID, Metadata: metadata, Password: "password"},
			token: usertoken,
			err:   nil,
		},
	}
	for _, tc := range cases {
		err := mainfluxSDK.UpdateUser(tc.user, tc.token)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
	}
}

func TestUpdatePassword(t *testing.T) {
	svc := newUserService()
	ts := newUserServer(svc)
	defer ts.Close()
	sdkConf := sdk.Config{
		UsersURL:        ts.URL,
		MsgContentType:  contentType,
		TLSVerification: false,
	}

	mainfluxSDK := sdk.NewSDK(sdkConf)
	user := sdk.User{Email: "user@example.com", Password: "password"}

	mockAuthzDB := map[string][]mocks.SubjectSet{}
	mockAuthzDB[user.Email] = append(mockAuthzDB[user.Email], mocks.SubjectSet{Object: "authorities", Relation: "member"})
	auth := mocks.NewAuthService(map[string]string{user.Email: user.Email}, mockAuthzDB)

	tkn, _ := auth.Issue(context.Background(), &mainflux.IssueReq{Id: user.ID, Email: user.Email, Type: 0})
	token := tkn.GetValue()
	_, err := mainfluxSDK.CreateUser(token, user)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
	usertoken, err := mainfluxSDK.CreateToken(user)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	cases := []struct {
		desc    string
		oldPass string
		newPass string
		token   string
		err     error
	}{
		{
			desc:    "update password for user",
			oldPass: "password",
			newPass: "password123",
			token:   usertoken,
			err:     nil,
		},
		{
			desc:    "update password for user with invalid token",
			oldPass: "password",
			newPass: "password123",
			token:   wrongValue,
			err:     createError(sdk.ErrFailedUpdate, http.StatusUnauthorized),
		},
		{
			desc:    "update password for user with empty token",
			oldPass: "password",
			newPass: "password123",
			token:   "",
			err:     createError(sdk.ErrFailedUpdate, http.StatusUnauthorized),
		},
	}
	for _, tc := range cases {
		err := mainfluxSDK.UpdatePassword(tc.oldPass, tc.newPass, tc.token)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
	}
}
