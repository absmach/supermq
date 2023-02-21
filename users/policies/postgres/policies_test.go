package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/internal/testsutil"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/mainflux/mainflux/users/clients"
	cpostgres "github.com/mainflux/mainflux/users/clients/postgres"
	"github.com/mainflux/mainflux/users/groups"
	gpostgres "github.com/mainflux/mainflux/users/groups/postgres"
	"github.com/mainflux/mainflux/users/policies"
	ppostgres "github.com/mainflux/mainflux/users/policies/postgres"
	"github.com/mainflux/mainflux/users/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	idProvider = uuid.New()
)

func TestPoliciesSave(t *testing.T) {
	t.Cleanup(func() { testsutil.CleanUpDB(t, db) })
	repo := ppostgres.NewPolicyRepo(database)
	crepo := cpostgres.NewClientRepo(database)

	uid := testsutil.GenerateUUID(t, idProvider)

	client := clients.Client{
		ID:   uid,
		Name: "policy-save@example.com",
		Credentials: clients.Credentials{
			Identity: "policy-save@example.com",
			Secret:   "pass",
		},
		Status: clients.EnabledStatus,
	}

	client, err := crepo.Save(context.Background(), client)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	uid = testsutil.GenerateUUID(t, idProvider)

	cases := []struct {
		desc   string
		policy policies.Policy
		err    error
	}{
		{
			desc: "add new policy successfully",
			policy: policies.Policy{
				OwnerID: client.ID,
				Subject: client.ID,
				Object:  uid,
				Actions: []string{"c_delete"},
			},
			err: nil,
		},
		{
			desc: "add policy with duplicate subject, object and action",
			policy: policies.Policy{
				OwnerID: client.ID,
				Subject: client.ID,
				Object:  uid,
				Actions: []string{"c_delete"},
			},
			err: errors.ErrConflict,
		},
	}

	for _, tc := range cases {
		err := repo.Save(context.Background(), tc.policy)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestPoliciesEvaluate(t *testing.T) {
	t.Cleanup(func() { testsutil.CleanUpDB(t, db) })
	repo := ppostgres.NewPolicyRepo(database)
	crepo := cpostgres.NewClientRepo(database)
	grepo := gpostgres.NewGroupRepo(database)

	client1 := clients.Client{
		ID:   testsutil.GenerateUUID(t, idProvider),
		Name: "connectedclients-clientA@example.com",
		Credentials: clients.Credentials{
			Identity: "connectedclients-clientA@example.com",
			Secret:   "pass",
		},
		Status: clients.EnabledStatus,
	}
	client2 := clients.Client{
		ID:   testsutil.GenerateUUID(t, idProvider),
		Name: "connectedclients-clientB@example.com",
		Credentials: clients.Credentials{
			Identity: "connectedclients-clientB@example.com",
			Secret:   "pass",
		},
		Status: clients.EnabledStatus,
	}
	group := groups.Group{
		ID:   testsutil.GenerateUUID(t, idProvider),
		Name: "connecting-group@example.com",
	}

	client1, err := crepo.Save(context.Background(), client1)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
	client2, err = crepo.Save(context.Background(), client2)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
	group, err = grepo.Save(context.Background(), group)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	policy1 := policies.Policy{
		OwnerID: client1.ID,
		Subject: client1.ID,
		Object:  group.ID,
		Actions: []string{"c_update", "g_update"},
	}
	policy2 := policies.Policy{
		OwnerID: client2.ID,
		Subject: client2.ID,
		Object:  group.ID,
		Actions: []string{"c_update", "g_update"},
	}
	err = repo.Save(context.Background(), policy1)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
	err = repo.Save(context.Background(), policy2)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	cases := map[string]struct {
		Subject string
		Object  string
		Action  string
		Domain  string
		err     error
	}{
		"evaluate valid client update":   {client1.ID, client2.ID, "c_update", "client", nil},
		"evaluate valid group update":    {client1.ID, group.ID, "g_update", "group", nil},
		"evaluate valid client list":     {client1.ID, client2.ID, "c_list", "client", errors.ErrAuthorization},
		"evaluate valid group list":      {client1.ID, group.ID, "g_list", "group", errors.ErrAuthorization},
		"evaluate invalid client delete": {client1.ID, client2.ID, "c_delete", "client", errors.ErrAuthorization},
		"evaluate invalid group delete":  {client1.ID, group.ID, "g_delete", "group", errors.ErrAuthorization},
		"evaluate invalid client update": {"unknown", "unknown", "c_update", "client", errors.ErrAuthorization},
		"evaluate invalid group update":  {"unknown", "unknown", "c_update", "group", errors.ErrAuthorization},
	}

	for desc, tc := range cases {
		p := policies.Policy{
			Subject: tc.Subject,
			Object:  tc.Object,
			Actions: []string{tc.Action},
		}
		err := repo.Evaluate(context.Background(), tc.Domain, p)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
	}
}

func TestPoliciesRetrieve(t *testing.T) {
	t.Cleanup(func() { testsutil.CleanUpDB(t, db) })
	repo := ppostgres.NewPolicyRepo(database)
	crepo := cpostgres.NewClientRepo(database)

	uid := testsutil.GenerateUUID(t, idProvider)

	client := clients.Client{
		ID:   uid,
		Name: "single-policy-retrieval@example.com",
		Credentials: clients.Credentials{
			Identity: "single-policy-retrieval@example.com",
			Secret:   "pass",
		},
		Status: clients.EnabledStatus,
	}

	client, err := crepo.Save(context.Background(), client)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	uid, err = idProvider.ID()
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	policy := policies.Policy{
		OwnerID: client.ID,
		Subject: client.ID,
		Object:  uid,
		Actions: []string{"c_delete"},
	}

	err = repo.Save(context.Background(), policy)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	cases := map[string]struct {
		Subject string
		Object  string
		err     error
	}{
		"retrieve existing policy":     {uid, uid, nil},
		"retrieve non-existing policy": {"unknown", "unknown", nil},
	}

	for desc, tc := range cases {
		pm := policies.Page{
			Subject: tc.Subject,
			Object:  tc.Object,
		}
		_, err := repo.Retrieve(context.Background(), pm)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
	}
}

func TestPoliciesUpdate(t *testing.T) {
	t.Cleanup(func() { testsutil.CleanUpDB(t, db) })
	repo := ppostgres.NewPolicyRepo(database)
	crepo := cpostgres.NewClientRepo(database)

	cid := testsutil.GenerateUUID(t, idProvider)
	pid := testsutil.GenerateUUID(t, idProvider)

	client := clients.Client{
		ID:   cid,
		Name: "policy-update@example.com",
		Credentials: clients.Credentials{
			Identity: "policy-update@example.com",
			Secret:   "pass",
		},
		Status: clients.EnabledStatus,
	}

	_, err := crepo.Save(context.Background(), client)
	require.Nil(t, err, fmt.Sprintf("unexpected error during saving client: %s", err))

	policy := policies.Policy{
		OwnerID: cid,
		Subject: cid,
		Object:  pid,
		Actions: []string{"c_delete"},
	}
	err = repo.Save(context.Background(), policy)
	require.Nil(t, err, fmt.Sprintf("unexpected error during saving policy: %s", err))

	cases := []struct {
		desc   string
		policy policies.Policy
		resp   policies.Policy
		err    error
	}{
		{
			desc: "update policy successfully",
			policy: policies.Policy{
				OwnerID: cid,
				Subject: cid,
				Object:  pid,
				Actions: []string{"c_update"},
			},
			resp: policies.Policy{
				OwnerID: cid,
				Subject: cid,
				Object:  pid,
				Actions: []string{"c_update"},
			},
			err: nil,
		},
		{
			desc: "update policy with missing owner id",
			policy: policies.Policy{
				OwnerID: "",
				Subject: cid,
				Object:  pid,
				Actions: []string{"c_delete"},
			},
			resp: policies.Policy{
				OwnerID: cid,
				Subject: cid,
				Object:  pid,
				Actions: []string{"c_delete"},
			},
			err: nil,
		},
		{
			desc: "update policy with missing subject",
			policy: policies.Policy{
				OwnerID: cid,
				Subject: "",
				Object:  pid,
				Actions: []string{"c_add"},
			},
			resp: policies.Policy{
				OwnerID: cid,
				Subject: cid,
				Object:  pid,
				Actions: []string{"c_delete"},
			},
			err: apiutil.ErrMissingPolicySub,
		},
		{
			desc: "update policy with missing object",
			policy: policies.Policy{
				OwnerID: cid,
				Subject: cid,
				Object:  "",
				Actions: []string{"c_add"},
			},
			resp: policies.Policy{
				OwnerID: cid,
				Subject: cid,
				Object:  pid,
				Actions: []string{"c_delete"},
			},

			err: apiutil.ErrMissingPolicyObj,
		},
		{
			desc: "update policy with missing action",
			policy: policies.Policy{
				OwnerID: cid,
				Subject: cid,
				Object:  pid,
				Actions: []string{""},
			},
			resp: policies.Policy{
				OwnerID: cid,
				Subject: cid,
				Object:  pid,
				Actions: []string{"c_delete"},
			},
			err: apiutil.ErrMalformedPolicyAct,
		},
	}

	for _, tc := range cases {
		err := repo.Update(context.Background(), tc.policy)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		policPage, err := repo.Retrieve(context.Background(), policies.Page{
			Offset:  uint64(0),
			Limit:   uint64(10),
			Subject: tc.policy.Subject,
		})
		if err == nil {
			assert.Equal(t, tc.resp, policPage.Policies[0], fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		}
	}
}

func TestPoliciesRetrievalAll(t *testing.T) {
	t.Cleanup(func() { testsutil.CleanUpDB(t, db) })
	postgres.NewDatabase(db, tracer)
	repo := ppostgres.NewPolicyRepo(database)
	crepo := cpostgres.NewClientRepo(database)

	var nPolicies = uint64(10)

	clientA := clients.Client{
		ID:   testsutil.GenerateUUID(t, idProvider),
		Name: "policyA-retrievalall@example.com",
		Credentials: clients.Credentials{
			Identity: "policyA-retrievalall@example.com",
			Secret:   "pass",
		},
		Status: clients.EnabledStatus,
	}
	clientB := clients.Client{
		ID:   testsutil.GenerateUUID(t, idProvider),
		Name: "policyB-retrievalall@example.com",
		Credentials: clients.Credentials{
			Identity: "policyB-retrievalall@example.com",
			Secret:   "pass",
		},
		Status: clients.EnabledStatus,
	}

	clientA, err := crepo.Save(context.Background(), clientA)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
	clientB, err = crepo.Save(context.Background(), clientB)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	for i := uint64(0); i < nPolicies; i++ {
		obj := fmt.Sprintf("TestRetrieveAll%d@example.com", i)
		if i%2 == 0 {
			policy := policies.Policy{
				OwnerID: clientA.ID,
				Subject: clientA.ID,
				Object:  obj,
				Actions: []string{"c_delete"},
			}
			err = repo.Save(context.Background(), policy)
			require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
		}
		policy := policies.Policy{
			Subject: clientB.ID,
			Object:  obj,
			Actions: []string{"c_add", "c_update"},
		}
		err = repo.Save(context.Background(), policy)
		require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
	}

	cases := map[string]struct {
		size uint64
		pm   policies.Page
	}{
		"retrieve all policies with limit and offset": {
			pm: policies.Page{
				Offset: 5,
				Limit:  nPolicies,
			},
			size: 10,
		},
		"retrieve all policies by owner id": {
			pm: policies.Page{
				Offset:  0,
				Limit:   nPolicies,
				Total:   nPolicies,
				OwnerID: clientA.ID,
			},
			size: 5,
		},
		"retrieve policies by wrong owner id": {
			pm: policies.Page{
				Offset:  0,
				Limit:   nPolicies,
				Total:   nPolicies,
				OwnerID: clientB.ID,
			},
			size: 0,
		},
		"retrieve all policies by Subject": {
			pm: policies.Page{
				Offset:  0,
				Limit:   nPolicies,
				Total:   nPolicies,
				Subject: clientA.ID,
			},
			size: 5,
		},
		"retrieve policies by wrong Subject": {
			pm: policies.Page{
				Offset:  0,
				Limit:   nPolicies,
				Total:   nPolicies,
				Subject: "wrongSubject",
			},
			size: 0,
		},

		"retrieve all policies by Object": {
			pm: policies.Page{
				Offset: 0,
				Limit:  nPolicies,
				Total:  nPolicies,
				Object: "TestRetrieveAll1@example.com",
			},
			size: 1,
		},
		"retrieve policies by wrong Object": {
			pm: policies.Page{
				Offset: 0,
				Limit:  nPolicies,
				Total:  nPolicies,
				Object: "TestRetrieveAll45@example.com",
			},
			size: 0,
		},
		"retrieve all policies by Action": {
			pm: policies.Page{
				Offset: 0,
				Limit:  nPolicies,
				Total:  nPolicies,
				Action: "c_delete",
			},
			size: 5,
		},
		"retrieve policies by wrong Action": {
			pm: policies.Page{
				Offset: 0,
				Limit:  nPolicies,
				Total:  nPolicies,
				Action: "wrongAction",
			},
			size: 0,
		},
		"retrieve all policies by owner id and subject": {
			pm: policies.Page{
				Offset:  0,
				Limit:   nPolicies,
				Total:   nPolicies,
				OwnerID: clientA.ID,
				Subject: clientA.ID,
			},
			size: 5,
		},
		"retrieve policies by wrong owner id and correct subject": {
			pm: policies.Page{
				Offset:  0,
				Limit:   nPolicies,
				Total:   nPolicies,
				OwnerID: clientB.ID,
				Subject: clientA.ID,
			},
			size: 0,
		},
		"retrieve policies by correct owner id and wrong subject": {
			pm: policies.Page{
				Offset:  0,
				Limit:   nPolicies,
				Total:   nPolicies,
				OwnerID: clientA.ID,
				Subject: "wrongSubject",
			},
			size: 0,
		},
		"retrieve policies by wrong owner id and wrong subject": {
			pm: policies.Page{
				Offset:  0,
				Limit:   nPolicies,
				Total:   nPolicies,
				OwnerID: clientB.ID,
			},
			size: 0,
		},
		"retrieve all policies by owner id and object": {
			pm: policies.Page{
				Offset:  0,
				Limit:   nPolicies,
				Total:   nPolicies,
				OwnerID: clientA.ID,
				Object:  "TestRetrieveAll2@example.com",
			},
			size: 1,
		},
		"retrieve policies by wrong owner id and correct object": {
			pm: policies.Page{
				Offset:  0,
				Limit:   nPolicies,
				Total:   nPolicies,
				OwnerID: clientB.ID,
				Object:  "TestRetrieveAll1@example.com",
			},
			size: 0,
		},
		"retrieve policies by correct owner id and wrong object": {
			pm: policies.Page{
				Offset:  0,
				Limit:   nPolicies,
				Total:   nPolicies,
				OwnerID: clientA.ID,
				Object:  "TestRetrieveAll45@example.com",
			},
			size: 0,
		},
		"retrieve policies by wrong owner id and wrong object": {
			pm: policies.Page{
				Offset:  0,
				Limit:   nPolicies,
				Total:   nPolicies,
				OwnerID: clientB.ID,
				Object:  "TestRetrieveAll45@example.com",
			},
			size: 0,
		},
		"retrieve all policies by owner id and action": {
			pm: policies.Page{
				Offset:  0,
				Limit:   nPolicies,
				Total:   nPolicies,
				OwnerID: clientA.ID,
				Action:  "c_delete",
			},
			size: 5,
		},
		"retrieve policies by wrong owner id and correct action": {
			pm: policies.Page{
				Offset:  0,
				Limit:   nPolicies,
				Total:   nPolicies,
				OwnerID: clientB.ID,
				Action:  "c_delete",
			},
			size: 0,
		},
		"retrieve policies by correct owner id and wrong action": {
			pm: policies.Page{
				Offset:  0,
				Limit:   nPolicies,
				Total:   nPolicies,
				OwnerID: clientA.ID,
				Action:  "wrongAction",
			},
			size: 0,
		},
		"retrieve policies by wrong owner id and wrong action": {
			pm: policies.Page{
				Offset:  0,
				Limit:   nPolicies,
				Total:   nPolicies,
				OwnerID: clientB.ID,
				Action:  "wrongAction",
			},
			size: 0,
		},
	}
	for desc, tc := range cases {
		page, err := repo.Retrieve(context.Background(), tc.pm)
		size := uint64(len(page.Policies))
		assert.Equal(t, tc.size, size, fmt.Sprintf("%s: expected size %d got %d\n", desc, tc.size, size))
		assert.Nil(t, err, fmt.Sprintf("%s: expected no error got %d\n", desc, err))
	}
}

func TestPoliciesDelete(t *testing.T) {
	t.Cleanup(func() { testsutil.CleanUpDB(t, db) })
	repo := ppostgres.NewPolicyRepo(database)
	crepo := cpostgres.NewClientRepo(database)

	client := clients.Client{
		ID:   testsutil.GenerateUUID(t, idProvider),
		Name: "policy-delete@example.com",
		Credentials: clients.Credentials{
			Identity: "policy-delete@example.com",
			Secret:   "pass",
		},
		Status: clients.EnabledStatus,
	}

	subject, err := crepo.Save(context.Background(), client)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	objectID := testsutil.GenerateUUID(t, idProvider)

	policy := policies.Policy{
		OwnerID: subject.ID,
		Subject: subject.ID,
		Object:  objectID,
		Actions: []string{"c_delete"},
	}

	err = repo.Save(context.Background(), policy)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	cases := map[string]struct {
		Subject string
		Object  string
		err     error
	}{
		"delete non-existing policy":                      {"unknown", "unknown", nil},
		"delete non-existing policy with correct subject": {subject.ID, "unknown", nil},
		"delete non-existing policy with correct object":  {"unknown", objectID, nil},
		"delete existing policy":                          {subject.ID, objectID, nil},
	}

	for desc, tc := range cases {
		policy := policies.Policy{
			Subject: tc.Subject,
			Object:  tc.Object,
		}
		err := repo.Delete(context.Background(), policy)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
	}
	pm := policies.Page{
		OwnerID: subject.ID,
		Subject: subject.ID,
		Object:  objectID,
		Action:  "c_delete",
	}
	policyPage, err := repo.Retrieve(context.Background(), pm)
	assert.Equal(t, uint64(0), policyPage.Total, fmt.Sprintf("retrieve policies unexpected total %d\n", policyPage.Total))
	require.Nil(t, err, fmt.Sprintf("retrieve policies unexpected error: %s", err))
}
