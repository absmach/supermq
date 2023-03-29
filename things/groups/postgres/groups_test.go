package postgres_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/mainflux/mainflux/internal/postgres"
	"github.com/mainflux/mainflux/internal/testsutil"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/mainflux/mainflux/things/clients"
	cpostgres "github.com/mainflux/mainflux/things/clients/postgres"
	"github.com/mainflux/mainflux/things/groups"
	gpostgres "github.com/mainflux/mainflux/things/groups/postgres"
	"github.com/mainflux/mainflux/things/policies"
	ppostgres "github.com/mainflux/mainflux/things/policies/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mfclients "github.com/mainflux/mainflux/internal/mainflux/clients"
)

const (
	maxNameSize = 1024
	maxDescSize = 1024
	maxLevel    = uint64(5)
	groupName   = "group"
	description = "description"
)

var (
	wrongID     = "wrong-id"
	invalidName = strings.Repeat("m", maxNameSize+10)
	validDesc   = strings.Repeat("m", 100)
	invalidDesc = strings.Repeat("m", maxDescSize+1)
	metadata    = groups.Metadata{
		"admin": "true",
	}
	idProvider = uuid.New()
)

func TestGroupSave(t *testing.T) {
	t.Cleanup(func() { testsutil.CleanUpDB(t, db) })
	groupRepo := gpostgres.NewRepository(database)

	usrID := testsutil.GenerateUUID(t, idProvider)
	grpID := testsutil.GenerateUUID(t, idProvider)

	cases := []struct {
		desc  string
		group groups.Group
		err   error
	}{
		{
			desc: "create new group successfully",
			group: groups.Group{
				ID:     grpID,
				Name:   groupName,
				Status: groups.EnabledStatus,
			},
			err: nil,
		},
		{
			desc: "create a new group with an existing name",
			group: groups.Group{
				ID:     grpID,
				Name:   groupName,
				Status: groups.EnabledStatus,
			},
			err: errors.ErrConflict,
		},
		{
			desc: "create group with an invalid name",
			group: groups.Group{
				ID:     testsutil.GenerateUUID(t, idProvider),
				Name:   invalidName,
				Status: groups.EnabledStatus,
			},
			err: errors.ErrMalformedEntity,
		},
		{
			desc: "create a group with invalid ID",
			group: groups.Group{
				ID:          usrID,
				Name:        "withInvalidDescription",
				Description: invalidDesc,
				Status:      groups.EnabledStatus,
			},
			err: errors.ErrMalformedEntity,
		},
		{
			desc: "create group with description",
			group: groups.Group{
				ID:          testsutil.GenerateUUID(t, idProvider),
				Name:        "withDescription",
				Description: validDesc,
				Status:      groups.EnabledStatus,
			},
			err: nil,
		},
		{
			desc: "create group with invalid description",
			group: groups.Group{
				ID:          testsutil.GenerateUUID(t, idProvider),
				Name:        "withInvalidDescription",
				Description: invalidDesc,
				Status:      groups.EnabledStatus,
			},
			err: errors.ErrMalformedEntity,
		},
		{
			desc: "create group with parent",
			group: groups.Group{
				ID:     testsutil.GenerateUUID(t, idProvider),
				Parent: grpID,
				Name:   "withParent",
				Status: groups.EnabledStatus,
			},
			err: nil,
		},
		{
			desc: "create a group with an invalid parent",
			group: groups.Group{
				ID:     testsutil.GenerateUUID(t, idProvider),
				Parent: invalidName,
				Name:   "withInvalidParent",
				Status: groups.EnabledStatus,
			},
			err: errors.ErrMalformedEntity,
		},
		{
			desc: "create a group with an owner",
			group: groups.Group{
				ID:     testsutil.GenerateUUID(t, idProvider),
				Owner:  usrID,
				Name:   "withOwner",
				Status: groups.EnabledStatus,
			},
			err: nil,
		},
		{
			desc: "create a group with an invalid owner",
			group: groups.Group{
				ID:     testsutil.GenerateUUID(t, idProvider),
				Owner:  invalidName,
				Name:   "withInvalidOwner",
				Status: groups.EnabledStatus,
			},
			err: errors.ErrMalformedEntity,
		},
		{
			desc: "create a group with metadata",
			group: groups.Group{
				ID:       testsutil.GenerateUUID(t, idProvider),
				Name:     "withMetadata",
				Metadata: metadata,
				Status:   groups.EnabledStatus,
			},
			err: nil,
		},
	}

	for _, tc := range cases {
		_, err := groupRepo.Save(context.Background(), tc.group)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}

}

func TestGroupRetrieveByID(t *testing.T) {
	t.Cleanup(func() { testsutil.CleanUpDB(t, db) })
	groupRepo := gpostgres.NewRepository(database)

	uid := testsutil.GenerateUUID(t, idProvider)
	group1 := groups.Group{
		ID:     testsutil.GenerateUUID(t, idProvider),
		Name:   groupName + "TestGroupRetrieveByID1",
		Owner:  uid,
		Status: groups.EnabledStatus,
	}

	_, err := groupRepo.Save(context.Background(), group1)
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	retrieved, err := groupRepo.RetrieveByID(context.Background(), group1.ID)
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	assert.True(t, retrieved.ID == group1.ID, fmt.Sprintf("Save group, ID: expected %s got %s\n", group1.ID, retrieved.ID))

	// Round to milliseconds as otherwise saving and retrieving from DB
	// adds rounding error.
	creationTime := time.Now().UTC().Round(time.Millisecond)
	group2 := groups.Group{
		ID:          testsutil.GenerateUUID(t, idProvider),
		Name:        groupName + "TestGroupRetrieveByID",
		Owner:       uid,
		Parent:      group1.ID,
		CreatedAt:   creationTime,
		UpdatedAt:   creationTime,
		Description: description,
		Metadata:    metadata,
		Status:      groups.EnabledStatus,
	}

	_, err = groupRepo.Save(context.Background(), group2)
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))

	retrieved, err = groupRepo.RetrieveByID(context.Background(), group2.ID)
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	assert.True(t, retrieved.ID == group2.ID, fmt.Sprintf("Save group, ID: expected %s got %s\n", group2.ID, retrieved.ID))
	assert.True(t, retrieved.CreatedAt.Equal(creationTime), fmt.Sprintf("Save group, CreatedAt: expected %s got %s\n", creationTime, retrieved.CreatedAt))
	assert.True(t, retrieved.UpdatedAt.Equal(creationTime), fmt.Sprintf("Save group, UpdatedAt: expected %s got %s\n", creationTime, retrieved.UpdatedAt))
	assert.True(t, retrieved.Parent == group1.ID, fmt.Sprintf("Save group, Level: expected %s got %s\n", group1.ID, retrieved.Parent))
	assert.True(t, retrieved.Description == description, fmt.Sprintf("Save group, Description: expected %v got %v\n", retrieved.Description, description))

	retrieved, err = groupRepo.RetrieveByID(context.Background(), testsutil.GenerateUUID(t, idProvider))
	assert.True(t, errors.Contains(err, errors.ErrNotFound), fmt.Sprintf("Retrieve group: expected %s got %s\n", errors.ErrNotFound, err))
}

func TestGroupRetrieveAll(t *testing.T) {
	t.Cleanup(func() { testsutil.CleanUpDB(t, db) })
	groupRepo := gpostgres.NewRepository(database)

	var nGroups = uint64(200)
	var ownerID = testsutil.GenerateUUID(t, idProvider)
	var parentID string
	for i := uint64(0); i < nGroups; i++ {
		creationTime := time.Now().UTC()
		group := groups.Group{
			ID:          testsutil.GenerateUUID(t, idProvider),
			Name:        fmt.Sprintf("%s-%d", groupName, i),
			Description: fmt.Sprintf("%s-description-%d", groupName, i),
			CreatedAt:   creationTime,
			UpdatedAt:   creationTime,
			Status:      groups.EnabledStatus,
		}
		if i == 1 {
			parentID = group.ID
		}
		if i%10 == 0 {
			group.Owner = ownerID
			group.Parent = parentID
		}
		if i%50 == 0 {
			group.Status = groups.DisabledStatus
		}
		_, err := groupRepo.Save(context.Background(), group)
		require.Nil(t, err, fmt.Sprintf("unexpected error: %s\n", err))
		parentID = group.ID
	}

	cases := map[string]struct {
		Size     uint64
		Metadata groups.GroupsPage
	}{
		"retrieve all groups": {
			Metadata: groups.GroupsPage{
				Page: groups.Page{
					Total:  nGroups,
					Limit:  nGroups,
					Status: groups.AllStatus,
				},
				Level: maxLevel,
			},
			Size: nGroups,
		},
		"retrieve all groups with offset": {
			Metadata: groups.GroupsPage{
				Page: groups.Page{
					Total:  nGroups,
					Offset: 50,
					Limit:  nGroups,
					Status: groups.AllStatus,
				},
				Level: maxLevel,
			},
			Size: nGroups - 50,
		},
		"retrieve all groups with limit": {
			Metadata: groups.GroupsPage{
				Page: groups.Page{
					Total:  nGroups,
					Offset: 0,
					Limit:  50,
					Status: groups.AllStatus,
				},
				Level: maxLevel,
			},
			Size: 50,
		},
		"retrieve all groups with offset and limit": {
			Metadata: groups.GroupsPage{
				Page: groups.Page{
					Total:  nGroups,
					Offset: 50,
					Limit:  50,
					Status: groups.AllStatus,
				},
				Level: maxLevel,
			},
			Size: 50,
		},
		"retrieve all groups with offset greater than limit": {
			Metadata: groups.GroupsPage{
				Page: groups.Page{
					Total:  nGroups,
					Offset: 250,
					Limit:  nGroups,
					Status: groups.AllStatus,
				},
				Level: maxLevel,
			},
			Size: 0,
		},
		"retrieve all groups with owner id": {
			Metadata: groups.GroupsPage{
				Page: groups.Page{
					Total:   nGroups,
					Limit:   nGroups,
					Subject: ownerID,
					OwnerID: ownerID,
					Status:  groups.AllStatus,
				},
				Level: maxLevel,
			},
			Size: 20,
		},
	}

	for desc, tc := range cases {
		page, err := groupRepo.RetrieveAll(context.Background(), tc.Metadata)
		size := len(page.Groups)
		assert.Equal(t, tc.Size, uint64(size), fmt.Sprintf("%s: expected size %d got %d\n", desc, tc.Size, size))
		assert.Nil(t, err, fmt.Sprintf("%s: expected no error got %d\n", desc, err))
	}
}

func TestGroupUpdate(t *testing.T) {
	t.Cleanup(func() { testsutil.CleanUpDB(t, db) })
	groupRepo := gpostgres.NewRepository(database)

	uid := testsutil.GenerateUUID(t, idProvider)

	creationTime := time.Now().UTC()
	updateTime := time.Now().UTC()
	groupID := testsutil.GenerateUUID(t, idProvider)

	group := groups.Group{
		ID:          groupID,
		Name:        groupName + "TestGroupUpdate",
		Owner:       uid,
		CreatedAt:   creationTime,
		UpdatedAt:   creationTime,
		Description: description,
		Metadata:    metadata,
		Status:      groups.EnabledStatus,
	}
	updatedName := groupName + "Updated"
	updatedMetadata := groups.Metadata{"admin": "false"}
	updatedDescription := description + "updated"
	_, err := groupRepo.Save(context.Background(), group)
	require.Nil(t, err, fmt.Sprintf("group save got unexpected error: %s", err))

	retrieved, err := groupRepo.RetrieveByID(context.Background(), group.ID)
	require.Nil(t, err, fmt.Sprintf("group save got unexpected error: %s", err))

	cases := []struct {
		desc          string
		groupUpdate   groups.Group
		groupExpected groups.Group
		err           error
	}{
		{
			desc: "update group name for existing id",
			groupUpdate: groups.Group{
				ID:        group.ID,
				Name:      updatedName,
				UpdatedAt: updateTime,
				Owner:     uid,
			},
			groupExpected: groups.Group{
				Name:        updatedName,
				Metadata:    retrieved.Metadata,
				Description: retrieved.Description,
			},
			err: nil,
		},
		{
			desc: "update group metadata for existing id",
			groupUpdate: groups.Group{
				ID:        group.ID,
				UpdatedAt: updateTime,
				Metadata:  updatedMetadata,
				Owner:     uid,
			},
			groupExpected: groups.Group{
				Name:        updatedName,
				UpdatedAt:   updateTime,
				Metadata:    updatedMetadata,
				Description: retrieved.Description,
			},
			err: nil,
		},
		{
			desc: "update group description for existing id",
			groupUpdate: groups.Group{
				ID:          group.ID,
				UpdatedAt:   updateTime,
				Description: updatedDescription,
				Owner:       uid,
			},
			groupExpected: groups.Group{
				Name:        updatedName,
				Description: updatedDescription,
				UpdatedAt:   updateTime,
				Metadata:    updatedMetadata,
			},
			err: nil,
		},
		{
			desc: "update group name and metadata for existing id",
			groupUpdate: groups.Group{
				ID:        group.ID,
				Name:      updatedName,
				UpdatedAt: updateTime,
				Metadata:  updatedMetadata,
				Owner:     uid,
			},
			groupExpected: groups.Group{
				Name:        updatedName,
				UpdatedAt:   updateTime,
				Metadata:    updatedMetadata,
				Description: updatedDescription,
			},
			err: nil,
		},
		{
			desc: "update group for invalid name",
			groupUpdate: groups.Group{
				ID:    group.ID,
				Owner: uid,
				Name:  invalidName,
			},
			err: errors.ErrMalformedEntity,
		},
		{
			desc: "update group for invalid description",
			groupUpdate: groups.Group{
				ID:          group.ID,
				Owner:       uid,
				Description: invalidDesc,
			},
			err: errors.ErrMalformedEntity,
		},
	}

	for _, tc := range cases {
		updated, err := groupRepo.Update(context.Background(), tc.groupUpdate)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		if err == nil {
			assert.True(t, updated.Name == tc.groupExpected.Name, fmt.Sprintf("%s:Name: expected %s got %s\n", tc.desc, tc.groupExpected.Name, updated.Name))
			assert.True(t, updated.Description == tc.groupExpected.Description, fmt.Sprintf("%s:Description: expected %s got %s\n", tc.desc, tc.groupExpected.Description, updated.Description))
			assert.True(t, updated.Metadata["admin"] == tc.groupExpected.Metadata["admin"], fmt.Sprintf("%s:Metadata: expected %d got %d\n", tc.desc, tc.groupExpected.Metadata["admin"], updated.Metadata["admin"]))
		}
	}
}

func TestClientsMemberships(t *testing.T) {
	t.Cleanup(func() { testsutil.CleanUpDB(t, db) })
	postgres.NewDatabase(db, tracer)
	crepo := cpostgres.NewRepository(database)
	grepo := gpostgres.NewRepository(database)
	prepo := ppostgres.NewRepository(database)

	clientA := clients.Client{
		ID:   testsutil.GenerateUUID(t, idProvider),
		Name: "client-memberships",
		Credentials: clients.Credentials{
			Identity: "client-memberships1@example.com",
			Secret:   testsutil.GenerateUUID(t, idProvider),
		},
		Metadata: clients.Metadata{},
		Status:   mfclients.EnabledStatus,
	}
	clientB := clients.Client{
		ID:   testsutil.GenerateUUID(t, idProvider),
		Name: "client-memberships",
		Credentials: clients.Credentials{
			Identity: "client-memberships2@example.com",
			Secret:   testsutil.GenerateUUID(t, idProvider),
		},
		Metadata: clients.Metadata{},
		Status:   mfclients.EnabledStatus,
	}
	group := groups.Group{
		ID:       testsutil.GenerateUUID(t, idProvider),
		Name:     "group-membership",
		Metadata: groups.Metadata{},
		Status:   groups.EnabledStatus,
	}

	policyA := policies.Policy{
		Subject: clientA.ID,
		Object:  group.ID,
		Actions: []string{"g_list"},
	}
	policyB := policies.Policy{
		Subject: clientB.ID,
		Object:  group.ID,
		Actions: []string{"g_list"},
	}

	_, err := crepo.Save(context.Background(), clientA)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("save client: expected %v got %s\n", nil, err))
	_, err = crepo.Save(context.Background(), clientB)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("save client: expected %v got %s\n", nil, err))
	_, err = grepo.Save(context.Background(), group)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("save group: expected %v got %s\n", nil, err))
	_, err = prepo.Save(context.Background(), policyA)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("save policy: expected %v got %s\n", nil, err))
	_, err = prepo.Save(context.Background(), policyB)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("save policy: expected %v got %s\n", nil, err))

	cases := map[string]struct {
		ID  string
		err error
	}{
		"retrieve membership for existing client":     {clientA.ID, nil},
		"retrieve membership for non-existing client": {wrongID, nil},
	}

	for desc, tc := range cases {
		mp, err := grepo.Memberships(context.Background(), tc.ID, groups.GroupsPage{Page: groups.Page{Total: 10, Offset: 0, Limit: 10, Status: groups.AllStatus, Subject: clientB.ID, Action: "g_list"}})
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
		if tc.ID == clientA.ID {
			assert.ElementsMatch(t, mp.Memberships, []groups.Group{group}, fmt.Sprintf("%s: expected %v got %v\n", desc, []groups.Group{group}, mp.Memberships))
		}
	}
}

func TestGroupChangeStatus(t *testing.T) {
	t.Cleanup(func() { testsutil.CleanUpDB(t, db) })
	dbMiddleware := postgres.NewDatabase(db, tracer)
	repo := gpostgres.NewRepository(dbMiddleware)

	group1 := groups.Group{
		ID:     testsutil.GenerateUUID(t, idProvider),
		Name:   "active-group",
		Status: groups.EnabledStatus,
	}
	group2 := groups.Group{
		ID:     testsutil.GenerateUUID(t, idProvider),
		Name:   "inactive-group",
		Status: groups.DisabledStatus,
	}

	group1, err := repo.Save(context.Background(), group1)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("add new group: expected %v got %s\n", nil, err))
	group2, err = repo.Save(context.Background(), group2)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("add new disabled group: expected %v got %s\n", nil, err))

	cases := []struct {
		desc  string
		group groups.Group
		err   error
	}{
		{
			desc: "change group status for an active group",
			group: groups.Group{
				ID:     group1.ID,
				Status: groups.EnabledStatus,
			},
			err: nil,
		},
		{
			desc: "change group status for a inactive group",
			group: groups.Group{
				ID:     group2.ID,
				Status: groups.EnabledStatus,
			},
			err: nil,
		},
		{
			desc: "change group status for an invalid group",
			group: groups.Group{
				ID:     "invalid",
				Status: groups.DisabledStatus,
			},
			err: errors.ErrNotFound,
		},
	}

	for _, tc := range cases {
		expected, err := repo.ChangeStatus(context.Background(), tc.group)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		if err == nil {
			assert.Equal(t, tc.group.Status, expected.Status, fmt.Sprintf("%s: expected %d got %d\n", tc.desc, tc.group.Status, expected.Status))
		}
	}
}
