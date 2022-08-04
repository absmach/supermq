package sdk_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/mainflux/mainflux/auth"
	httpapi "github.com/mainflux/mainflux/auth/api/http"
	"github.com/mainflux/mainflux/auth/jwt"
	"github.com/mainflux/mainflux/auth/mocks"
	"github.com/mainflux/mainflux/logger"
	sdk "github.com/mainflux/mainflux/pkg/sdk/go"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"

	"net/http/httptest"
)

const (
	loginDuration    = 30 * time.Minute
	secret           = "secret"
	groupID          = "testID"
	invalidGroupID   = "invalidID"
	authoritiesObj   = "authorities"
	memberRelation   = "member"
	membersUserType  = "user"
	membersThingType = "thing"
	groupName        = "testGroupName"
	invalidToken     = "invalidToken"
)

var (
	group = sdk.Group{
		ID:   groupID,
		Name: groupName,
	}

	invalidGroup = sdk.Group{
		ID:       "testID",
		Name:     "group",
		ParentID: "parentId",
	}
	noNameGroup = sdk.Group{
		ID: groupID,
	}
	memberID  = "testID"
	memberIDs = []string{"testID", "testID1", "testID2"}
)

func newThingsAuthServer(svc auth.Service) *httptest.Server {
	logger := logger.NewMock()
	mux := httpapi.MakeHandler(svc, mocktracer.New(), logger)
	return httptest.NewServer(mux)
}
func newThingAuthService() auth.Service {
	repo := mocks.NewKeyRepository()
	groupRepo := mocks.NewGroupRepository()
	idProvider := uuid.NewMock()

	mockAuthzDB := map[string][]mocks.MockSubjectSet{}
	mockAuthzDB[groupID] = append(mockAuthzDB[groupID], mocks.MockSubjectSet{Object: authoritiesObj, Relation: memberRelation})
	ketoMock := mocks.NewKetoMock(mockAuthzDB)

	t := jwt.New(secret)
	return auth.New(repo, groupRepo, idProvider, t, ketoMock, loginDuration)
}

func TestCreateGroup(t *testing.T) {
	svc := newThingAuthService()
	ts := newThingsAuthServer(svc)
	defer ts.Close()
	sdkConf := sdk.Config{
		AuthURL:         ts.URL,
		MsgContentType:  contentType,
		TLSVerification: true,
	}
	mainfluxSDK := sdk.NewSDK(sdkConf)
	_, loginSecret, err := svc.Issue(context.Background(), "", auth.Key{Type: auth.LoginKey, IssuedAt: time.Now(), IssuerID: groupID, Subject: email})
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))

	cases := []struct {
		desc  string
		group sdk.Group
		token string
		err   error
	}{
		{
			desc:  "create new group",
			group: group,
			token: loginSecret,
			err:   nil,
		},
		{
			desc:  "create new group with empty token",
			group: group,
			token: "",
			err:   createError(sdk.ErrFailedCreation, http.StatusInternalServerError),
		},
		{
			desc:  "create new group with invalid token",
			group: group,
			token: invalidToken,
			err:   createError(sdk.ErrFailedCreation, http.StatusUnauthorized),
		},
		{
			desc:  "create new group with invalid parent",
			group: invalidGroup,
			token: loginSecret,
			err:   createError(sdk.ErrFailedCreation, http.StatusInternalServerError),
		},
		{
			desc:  "create new group without group name",
			group: noNameGroup,
			token: loginSecret,
			err:   createError(sdk.ErrFailedCreation, http.StatusBadRequest),
		},
		{
			desc:  "create new group with empty group",
			group: noNameGroup,
			token: loginSecret,
			err:   createError(sdk.ErrFailedCreation, http.StatusBadRequest),
		},
	}
	for _, tc := range cases {
		_, err := mainfluxSDK.CreateGroup(tc.group, tc.token)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
	}
}

func TestDeleteGroup(t *testing.T) {
	svc := newThingAuthService()
	ts := newThingsAuthServer(svc)
	defer ts.Close()
	sdkConf := sdk.Config{
		AuthURL:         ts.URL,
		MsgContentType:  contentType,
		TLSVerification: true,
	}
	mainfluxSDK := sdk.NewSDK(sdkConf)

	_, loginSecret, err := svc.Issue(context.Background(), "", auth.Key{Type: auth.LoginKey, IssuedAt: time.Now(), IssuerID: groupID, Subject: email})
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))

	id, err := mainfluxSDK.CreateGroup(group, loginSecret)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	cases := []struct {
		desc    string
		groupID string
		token   string
		err     error
	}{
		{
			desc:    "delete group with invalid token",
			groupID: id,
			token:   "wrongToken",
			err:     createError(sdk.ErrFailedRemoval, http.StatusUnauthorized),
		},
		{
			desc:    "delete non-existing group",
			groupID: invalidGroupID,
			token:   loginSecret,
			err:     createError(sdk.ErrFailedRemoval, http.StatusNotFound),
		},
		{
			desc:    "delete group with invalid id",
			groupID: "",
			token:   loginSecret,
			err:     createError(sdk.ErrFailedRemoval, http.StatusBadRequest),
		},
		{
			desc:    "delete group with empty token",
			groupID: id,
			token:   "",
			err:     createError(sdk.ErrFailedRemoval, http.StatusInternalServerError),
		},
		{
			desc:    "delete existing group",
			groupID: id,
			token:   loginSecret,
			err:     nil,
		},
		{
			desc:    "delete deleted group",
			groupID: id,
			token:   loginSecret,
			err:     createError(sdk.ErrFailedRemoval, http.StatusNotFound),
		},
	}
	for _, tc := range cases {
		err := mainfluxSDK.DeleteGroup(tc.groupID, tc.token)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
	}
}

func TestAssign(t *testing.T) {
	svc := newThingAuthService()
	ts := newThingsAuthServer(svc)
	defer ts.Close()
	sdkConf := sdk.Config{
		AuthURL:         ts.URL,
		MsgContentType:  contentType,
		TLSVerification: true,
	}
	mainfluxSDK := sdk.NewSDK(sdkConf)

	_, loginSecret, err := svc.Issue(context.Background(), "", auth.Key{Type: auth.LoginKey, IssuedAt: time.Now(), IssuerID: groupID, Subject: email})
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))

	id, err := mainfluxSDK.CreateGroup(group, loginSecret)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	cases := []struct {
		desc        string
		memberIDs   []string
		membersType string
		groupID     string
		token       string
		err         error
	}{
		{
			desc:        "assign members to a group without memberIDs",
			memberIDs:   nil,
			membersType: membersUserType,
			groupID:     id,
			token:       loginSecret,
			err:         createError(sdk.ErrMemberAdd, http.StatusBadRequest),
		},
		{
			desc:        "assign members to a group without members type",
			memberIDs:   memberIDs,
			membersType: "",
			groupID:     id,
			token:       loginSecret,
			err:         createError(sdk.ErrMemberAdd, http.StatusBadRequest),
		},
		{
			desc:        "assign members to a group without group ID",
			memberIDs:   memberIDs,
			membersType: membersUserType,
			groupID:     "",
			token:       loginSecret,
			err:         createError(sdk.ErrMemberAdd, http.StatusBadRequest),
		},
		{
			desc:        "assign members to a group without token",
			memberIDs:   memberIDs,
			membersType: membersUserType,
			groupID:     id,
			token:       "",
			err:         createError(sdk.ErrMemberAdd, http.StatusInternalServerError),
		},
		{
			desc:        "assign members to a user group",
			memberIDs:   memberIDs,
			membersType: membersUserType,
			groupID:     id,
			token:       loginSecret,
			err:         nil,
		},
		{
			desc:        "assign members to a thing group",
			memberIDs:   memberIDs,
			membersType: membersThingType,
			groupID:     id,
			token:       loginSecret,
			err:         nil,
		},
	}
	for _, tc := range cases {
		err := mainfluxSDK.Assign(tc.memberIDs, tc.membersType, tc.groupID, tc.token)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
	}
}

func TestUnassign(t *testing.T) {
	svc := newThingAuthService()
	ts := newThingsAuthServer(svc)
	defer ts.Close()
	sdkConf := sdk.Config{
		AuthURL:         ts.URL,
		MsgContentType:  contentType,
		TLSVerification: true,
	}
	mainfluxSDK := sdk.NewSDK(sdkConf)

	_, loginSecret, err := svc.Issue(context.Background(), "", auth.Key{Type: auth.LoginKey, IssuedAt: time.Now(), IssuerID: groupID, Subject: email})
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))

	id, err := mainfluxSDK.CreateGroup(group, loginSecret)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	err = mainfluxSDK.Assign(memberIDs, membersUserType, id, loginSecret)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	cases := []struct {
		desc        string
		memberID    string
		membersType string
		groupID     string
		token       string
		err         error
	}{
		{
			desc:        "unassign member from group without token",
			memberID:    memberID,
			membersType: membersUserType,
			groupID:     id,
			token:       "",
			err:         createError(sdk.ErrFailedRemoval, http.StatusInternalServerError),
		},
		{
			desc:        "unassign member from group with invalid token",
			memberID:    memberID,
			membersType: membersUserType,
			groupID:     id,
			token:       invalidToken,
			err:         createError(sdk.ErrFailedRemoval, http.StatusUnauthorized),
		},
		{
			desc:        "unassign member from group without group ID",
			memberID:    memberID,
			membersType: membersUserType,
			groupID:     "",
			token:       loginSecret,
			err:         createError(sdk.ErrFailedRemoval, http.StatusBadRequest),
		},
		{
			desc:        "unassign member from group without member ID",
			memberID:    "",
			membersType: membersUserType,
			groupID:     id,
			token:       loginSecret,
			err:         createError(sdk.ErrFailedRemoval, http.StatusNotFound),
		},
		{
			desc:        "unassign member from group",
			memberID:    memberID,
			membersType: membersUserType,
			groupID:     id,
			token:       loginSecret,
			err:         nil,
		},
		{
			desc:        "unassign member from group which is already unassigned",
			memberID:    memberID,
			membersType: membersUserType,
			groupID:     id,
			token:       loginSecret,
			err:         createError(sdk.ErrFailedRemoval, http.StatusNotFound),
		},
	}
	for _, tc := range cases {
		err := mainfluxSDK.Unassign(tc.token, tc.groupID, tc.memberID)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
	}
}

// TODO: Add test case for multiple groups.
// TODO: Add testcase for offset and limit
func TestMembers(t *testing.T) {
	svc := newThingAuthService()
	ts := newThingsAuthServer(svc)
	defer ts.Close()
	sdkConf := sdk.Config{
		AuthURL:         ts.URL,
		MsgContentType:  contentType,
		TLSVerification: true,
	}
	mainfluxSDK := sdk.NewSDK(sdkConf)

	_, loginSecret, err := svc.Issue(context.Background(), "", auth.Key{Type: auth.LoginKey, IssuedAt: time.Now(), IssuerID: groupID, Subject: email})
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))

	id, err := mainfluxSDK.CreateGroup(group, loginSecret)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	err = mainfluxSDK.Assign(memberIDs, membersUserType, id, loginSecret)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	cases := []struct {
		desc      string
		groupID   string
		groupType string
		token     string
		offset    uint64
		limit     uint64
		response  []string
		err       error
	}{
		{
			desc:     "get list of all members without group ID",
			groupID:  "",
			token:    loginSecret,
			offset:   offset,
			limit:    limit,
			response: nil,
			err:      createError(sdk.ErrFailedFetch, http.StatusBadRequest),
		},
		{
			desc:     "get list of all members with invalid token",
			groupID:  id,
			token:    invalidToken,
			offset:   offset,
			limit:    limit,
			response: nil,
			err:      createError(sdk.ErrFailedFetch, http.StatusUnauthorized),
		},
		{
			desc:     "get list of all members without token",
			groupID:  id,
			token:    "",
			offset:   offset,
			limit:    limit,
			response: nil,
			err:      createError(sdk.ErrFailedFetch, http.StatusInternalServerError),
		},
		{
			desc:     "get list of all members",
			groupID:  id,
			token:    loginSecret,
			offset:   offset,
			limit:    limit,
			response: memberIDs,
			err:      nil,
		},
	}
	for _, tc := range cases {
		page, err := mainfluxSDK.Members(tc.groupID, tc.token, tc.offset, tc.limit)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
		for _, v := range memberIDs {
			var c bool
			for _, h := range page.Members {
				if v == h {
					c = true
					break
				}
			}
			if !c {
				assert.Equal(t, tc.response, page.Members, fmt.Sprintf("%s: expected response member %v, got %s", tc.desc, tc.response, page.Members))

			}
		}
	}
}

////TODO add testcase for offset and limit(?strange behaviour- not respond for change offset and limit?)
func TestGroups(t *testing.T) {
	svc := newThingAuthService()
	ts := newThingsAuthServer(svc)
	defer ts.Close()
	sdkConf := sdk.Config{
		AuthURL:         ts.URL,
		MsgContentType:  contentType,
		TLSVerification: true,
	}
	var groups []sdk.Group
	mainfluxSDK := sdk.NewSDK(sdkConf)

	_, loginSecret, err := svc.Issue(context.Background(), "", auth.Key{Type: auth.LoginKey, IssuedAt: time.Now(), IssuerID: groupID, Subject: email})
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))

	for i := 1; i < 101; i++ {
		name := fmt.Sprintf("testGroupName-%d", i)
		group := sdk.Group{ID: groupID, Name: name}
		id, err := mainfluxSDK.CreateGroup(group, loginSecret)
		require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
		group.ID = id
		groups = append(groups, group)
	}

	cases := []struct {
		desc     string
		token    string
		name     string
		offset   uint64
		limit    uint64
		metadata map[string]interface{}
		response []sdk.Group
		err      error
	}{
		{
			desc:     "get a list of groups with invalid token",
			token:    invalidToken,
			offset:   offset,
			limit:    limit,
			response: nil,
			metadata: make(map[string]interface{}),
			err:      createError(sdk.ErrFailedFetch, http.StatusUnauthorized),
		},
		{
			desc:     "get a list of groups with empty token",
			token:    "",
			offset:   offset,
			limit:    limit,
			response: nil,
			metadata: make(map[string]interface{}),
			err:      createError(sdk.ErrFailedFetch, http.StatusInternalServerError),
		},
		{
			desc:     "get a list of groups without limit",
			token:    loginSecret,
			offset:   offset,
			limit:    0,
			response: nil,
			metadata: make(map[string]interface{}),
			err:      nil,
		},
		{
			desc:     "get a list of groups with limit greater then max",
			token:    loginSecret,
			offset:   offset,
			limit:    1000,
			response: nil,
			metadata: make(map[string]interface{}),
			err:      nil,
		},
		{
			desc:     "get a list of groups with offset greater then max",
			token:    loginSecret,
			offset:   1000,
			limit:    limit,
			response: nil,
			metadata: make(map[string]interface{}),
			err:      nil,
		},
		{
			desc:     "get a list of groups",
			token:    loginSecret,
			offset:   offset,
			limit:    limit,
			response: groups,
			metadata: make(map[string]interface{}),
			err:      nil,
		},
	}
	for _, tc := range cases {
		filter := sdk.PageMetadata{
			Name:     tc.name,
			Total:    total,
			Offset:   tc.offset,
			Limit:    tc.limit,
			Metadata: tc.metadata,
		}
		page, err := mainfluxSDK.Groups(filter, tc.token)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
		for _, v := range groups {
			var c bool
			for _, h := range page.Groups {
				if v.ID == h.ID {
					c = true
					break
				}
			}
			if !c {
				assert.Equal(t, tc.response, page.Groups, fmt.Sprintf("%s: expected response member %s, got %s", tc.desc, tc.response, page.Groups))

			}
		}
	}
}

//TODO check for loop for ids
//?Strange behavior with returning errors
func TestParents(t *testing.T) {
	svc := newThingAuthService()
	ts := newThingsAuthServer(svc)
	defer ts.Close()
	var groups []sdk.Group
	var ids []string

	sdkConf := sdk.Config{
		AuthURL:         ts.URL,
		MsgContentType:  contentType,
		TLSVerification: true,
	}

	mainfluxSDK := sdk.NewSDK(sdkConf)
	_, loginSecret, err := svc.Issue(context.Background(), "", auth.Key{Type: auth.LoginKey, IssuedAt: time.Now(), IssuerID: groupID, Subject: email})
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))

	for i := 1; i < 2; i++ {
		name := fmt.Sprintf("testGroupName-%d", i)
		group := sdk.Group{ID: groupID, Name: name}
		id, err := mainfluxSDK.CreateGroup(group, loginSecret)
		require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
		group.ID = id
		ids = append(ids, id)
		groups = append(groups, group)
	}

	cases := []struct {
		desc     string
		id       string
		token    string
		offset   uint64
		limit    uint64
		response []sdk.Group
		err      error
	}{
		{
			desc:     "get a non existing group",
			token:    loginSecret,
			id:       invalidGroupID,
			offset:   offset,
			limit:    limit,
			response: nil,
			err:      createError(sdk.ErrFailedFetch, http.StatusNotFound),
		},
		{
			desc:     "get a list of parent groups without token",
			token:    "",
			id:       ids[0],
			offset:   offset,
			limit:    limit,
			response: nil,
			err:      createError(sdk.ErrFailedFetch, http.StatusInternalServerError),
		},
		{
			desc:     "get a list of parent groups with invalid token",
			token:    invalidToken,
			id:       ids[0],
			offset:   offset,
			limit:    limit,
			response: nil,
			err:      createError(sdk.ErrFailedFetch, http.StatusUnauthorized),
		},
		{
			desc:   "get a list of all parent groups",
			id:     ids[0],
			token:  loginSecret,
			offset: offset,
			limit:  limit,
			err:    nil,
		},
	}
	for _, tc := range cases {

		page, err := mainfluxSDK.Parents(tc.id, tc.offset, tc.limit, tc.token)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
		for _, v := range groups {
			var c bool
			for _, h := range page.Groups {
				if v.ID == h.ID {
					c = true
					break
				}
			}
			if !c {
				assert.Equal(t, tc.response, page.Groups, fmt.Sprintf("%s: expected response member %s, got %s", tc.desc, tc.response, page.Groups))

			}
		}
	}
}

// ?Strange error behaviour. Not returning err
func TestChildren(t *testing.T) {
	svc := newThingAuthService()
	ts := newThingsAuthServer(svc)
	defer ts.Close()
	sdkConf := sdk.Config{
		AuthURL:         ts.URL,
		MsgContentType:  contentType,
		TLSVerification: true,
	}
	var groups []sdk.Group
	mainfluxSDK := sdk.NewSDK(sdkConf)

	_, loginSecret, err := svc.Issue(context.Background(), "", auth.Key{Type: auth.LoginKey, IssuedAt: time.Now(), IssuerID: groupID, Subject: email})
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))

	parentID, err := mainfluxSDK.CreateGroup(group, loginSecret)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	for i := 1; i < 101; i++ {
		name := fmt.Sprintf("testChildGroupName-%d", i)
		group := sdk.Group{ID: groupID, Name: name, ParentID: parentID}
		id, err := mainfluxSDK.CreateGroup(group, loginSecret)
		require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
		group.ID = id
		groups = append(groups, group)
	}

	cases := []struct {
		desc     string
		id       string
		token    string
		offset   uint64
		limit    uint64
		response []sdk.Group
		err      error
	}{
		{
			desc:     "get all children from no existing group",
			id:       "",
			token:    loginSecret,
			offset:   offset,
			limit:    limit,
			response: []sdk.Group{},
			err:      nil,
		},
		{
			desc:     "get all children from invalid group",
			id:       invalidGroupID,
			token:    loginSecret,
			offset:   offset,
			limit:    limit,
			response: []sdk.Group{},
			err:      nil,
		},
		{
			desc:     "get all children groups without token",
			id:       parentID,
			token:    "",
			offset:   offset,
			limit:    limit,
			response: nil,
			err:      createError(sdk.ErrFailedFetch, http.StatusInternalServerError),
		},
		{
			desc:     "get all children groups with invalid token",
			id:       parentID,
			token:    invalidToken,
			offset:   offset,
			limit:    limit,
			response: nil,
			err:      createError(sdk.ErrFailedFetch, http.StatusUnauthorized),
		},
		{
			desc:     "get all children groups with 0 limit",
			id:       parentID,
			token:    loginSecret,
			offset:   offset,
			limit:    110,
			response: groups,
			err:      nil,
		},
		{
			desc:     "get all children groups",
			id:       parentID,
			token:    loginSecret,
			offset:   offset,
			limit:    limit,
			response: groups,
			err:      nil,
		},
	}
	for _, tc := range cases {
		page, err := mainfluxSDK.Children(tc.id, tc.offset, tc.limit, tc.token)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
		for _, v := range groups {
			var c bool
			for _, h := range page.Groups {
				if v.Name == h.Name {
					c = true
					break
				}
			}
			if !c {
				assert.Equal(t, tc.response, page.Groups, fmt.Sprintf("%s: expected response member %s, got %s", tc.desc, tc.response, page.Groups))

			}
		}
	}
}

func TestGroup(t *testing.T) {
	svc := newThingAuthService()
	ts := newThingsAuthServer(svc)
	defer ts.Close()
	sdkConf := sdk.Config{
		AuthURL:         ts.URL,
		MsgContentType:  contentType,
		TLSVerification: true,
	}
	mainfluxSDK := sdk.NewSDK(sdkConf)
	_, loginSecret, err := svc.Issue(context.Background(), "", auth.Key{Type: auth.LoginKey, IssuedAt: time.Now(), IssuerID: groupID, Subject: email})
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))

	id, err := mainfluxSDK.CreateGroup(group, loginSecret)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	cases := []struct {
		desc  string
		id    string
		token string
		err   error
	}{
		{
			desc:  "get existing group",
			id:    id,
			token: loginSecret,
			err:   nil,
		},
		{
			desc:  "get group without id",
			id:    "",
			token: loginSecret,
			err:   createError(sdk.ErrFailedFetch, http.StatusBadRequest),
		},
		{
			desc:  "get non existing group",
			id:    invalidGroupID,
			token: loginSecret,
			err:   createError(sdk.ErrFailedFetch, http.StatusNotFound),
		},
		{
			desc:  "get group without token",
			id:    id,
			token: "",
			err:   createError(sdk.ErrFailedFetch, http.StatusInternalServerError),
		},
		{
			desc:  "get group with invalid token",
			id:    id,
			token: invalidToken,
			err:   createError(sdk.ErrFailedFetch, http.StatusUnauthorized),
		},
	}
	for _, tc := range cases {
		_, err := mainfluxSDK.Group(tc.id, tc.token)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
	}
}

func TestUpdateGroup(t *testing.T) {
	svc := newThingAuthService()
	ts := newThingsAuthServer(svc)
	defer ts.Close()
	sdkConf := sdk.Config{
		AuthURL:         ts.URL,
		MsgContentType:  contentType,
		TLSVerification: true,
	}
	mainfluxSDK := sdk.NewSDK(sdkConf)

	_, loginSecret, err := svc.Issue(context.Background(), "", auth.Key{Type: auth.LoginKey, IssuedAt: time.Now(), IssuerID: groupID, Subject: email})
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))

	id, err := mainfluxSDK.CreateGroup(group, loginSecret)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	updatedGroup := sdk.Group{ID: id, Name: "updatedGroup", Description: "testDesc"}

	cases := []struct {
		desc  string
		group sdk.Group
		token string
		err   error
	}{
		{
			desc:  "update existing group",
			group: updatedGroup,
			token: loginSecret,
			err:   nil,
		},
		{
			desc:  "update non-existing group",
			group: sdk.Group{ID: "0", Name: "updatedGroup", Description: "testDesc"},
			token: loginSecret,
			err:   createError(sdk.ErrFailedUpdate, http.StatusNotFound),
		},
		{
			desc:  "update group with invalid id",
			group: sdk.Group{ID: "", Name: "updatedGroup", Description: "testDesc"},
			token: loginSecret,
			err:   createError(sdk.ErrFailedUpdate, http.StatusBadRequest),
		},
		{
			desc:  "update group with invalid token",
			group: updatedGroup,
			token: invalidToken,
			err:   createError(sdk.ErrFailedUpdate, http.StatusUnauthorized),
		},
		{
			desc:  "update group with empty token",
			group: updatedGroup,
			token: "",
			err:   createError(sdk.ErrFailedUpdate, http.StatusInternalServerError),
		},
	}
	for _, tc := range cases {
		err := mainfluxSDK.UpdateGroup(tc.group, tc.token)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
	}
}

func TestMemberships(t *testing.T) {
	svc := newThingAuthService()
	ts := newThingsAuthServer(svc)
	defer ts.Close()
	var secondMemberID = "testID1"
	sdkConf := sdk.Config{
		AuthURL:         ts.URL,
		MsgContentType:  contentType,
		TLSVerification: true,
	}
	mainfluxSDK := sdk.NewSDK(sdkConf)
	_, loginSecret, err := svc.Issue(context.Background(), "", auth.Key{Type: auth.LoginKey, IssuedAt: time.Now(), IssuerID: groupID, Subject: email})
	assert.Nil(t, err, fmt.Sprintf("Issuing login key expected to succeed: %s", err))

	var ids []string
	var groups []sdk.Group
	for i := 1; i < 5; i++ {
		name := fmt.Sprintf("testGroupName-%d", i)
		group := sdk.Group{ID: groupID, Name: name}
		id, err := mainfluxSDK.CreateGroup(group, loginSecret)
		require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
		group.ID = id
		ids = append(ids, id)
		groups = append(groups, group)

	}
	for _, id := range ids {
		err := mainfluxSDK.Assign(memberIDs, membersUserType, id, loginSecret)
		require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	}

	cases := []struct {
		desc     string
		userID   string
		token    string
		offset   uint64
		limit    uint64
		response []sdk.Group
		err      error
	}{
		{
			desc:     "get memberships for non existing member",
			userID:   "",
			token:    loginSecret,
			offset:   offset,
			limit:    limit,
			response: nil,
			err:      createError(sdk.ErrFailedFetch, http.StatusBadRequest),
		},
		{
			desc:     "get all existing memberships without token",
			userID:   memberID,
			token:    "",
			offset:   offset,
			limit:    limit,
			response: nil,
			err:      createError(sdk.ErrFailedFetch, http.StatusInternalServerError),
		},
		{
			desc:     "get all existing memberships with invalid token",
			userID:   memberID,
			token:    invalidToken,
			offset:   offset,
			limit:    limit,
			response: nil,
			err:      createError(sdk.ErrFailedFetch, http.StatusUnauthorized),
		},
		{
			desc:     "get all existing memberships with offset greater than max",
			userID:   memberID,
			token:    loginSecret,
			offset:   100,
			limit:    limit,
			response: []sdk.Group{},
			err:      nil,
		},
		{
			desc:     "get all existing memberships with zero limit",
			userID:   memberID,
			token:    loginSecret,
			offset:   offset,
			limit:    0,
			response: []sdk.Group{},
			err:      nil,
		},
		{
			desc:     "get all first member existing memberships",
			userID:   memberID,
			token:    loginSecret,
			offset:   offset,
			limit:    limit,
			response: groups,
			err:      nil,
		},
		{
			desc:     "get all second member existing memberships",
			userID:   secondMemberID,
			token:    loginSecret,
			offset:   offset,
			limit:    limit,
			response: groups,
			err:      nil,
		},
	}
	for _, tc := range cases {
		page, err := mainfluxSDK.Memberships(tc.userID, tc.token, tc.offset, tc.limit)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
		for _, v := range groups {
			var c bool
			for _, h := range page.Groups {
				if v.ID == h.ID {
					c = true
					break
				}
			}
			if !c {
				assert.Equal(t, tc.response, page.Groups, fmt.Sprintf("%s: expected response member %s, got %s", tc.desc, tc.response, page.Groups))

			}
		}
	}
}
