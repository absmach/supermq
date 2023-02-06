package policies_test

import (
	context "context"
	fmt "fmt"
	"regexp"
	"testing"

	"github.com/mainflux/mainflux/clients/clients"
	cmocks "github.com/mainflux/mainflux/clients/clients/mocks"
	"github.com/mainflux/mainflux/clients/hasher"
	"github.com/mainflux/mainflux/clients/jwt"
	"github.com/mainflux/mainflux/clients/policies"
	pmocks "github.com/mainflux/mainflux/clients/policies/mocks"
	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/internal/testsutil"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	idProvider     = uuid.New()
	phasher        = hasher.New()
	secret         = "strongsecret"
	inValidToken   = "invalidToken"
	memberActions  = []string{"g_list"}
	authoritiesObj = "authorities"
	passRegex      = regexp.MustCompile("^.{8,}$")
)

func generateValidToken(t *testing.T, clientID string, svc clients.Service, cRepo *cmocks.ClientRepository) string {
	client := clients.Client{
		ID:   clientID,
		Name: "validtoken",
		Credentials: clients.Credentials{
			Identity: "validtoken",
			Secret:   secret,
		},
		Status: clients.EnabledStatus,
	}
	rClient := client
	rClient.Credentials.Secret, _ = phasher.Hash(client.Credentials.Secret)

	repoCall := cRepo.On("RetrieveByIdentity", context.Background(), mock.Anything).Return(rClient, nil)
	token, err := svc.IssueToken(context.Background(), client.Credentials.Identity, client.Credentials.Secret)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("Create token expected nil got %s\n", err))
	repoCall.Unset()
	return token.AccessToken
}

func TestAddPolicy(t *testing.T) {
	cRepo := new(cmocks.ClientRepository)
	pRepo := new(pmocks.PolicyRepository)
	tokenizer := jwt.NewTokenRepo([]byte(secret))
	csvc := clients.NewService(cRepo, pRepo, tokenizer, phasher, idProvider, passRegex)
	svc := policies.NewService(pRepo, tokenizer, idProvider)

	policy := policies.Policy{Object: "obj1", Actions: []string{"m_read"}, Subject: "sub1"}

	cases := []struct {
		desc   string
		policy policies.Policy
		page   policies.PolicyPage
		token  string
		err    error
	}{
		{
			desc:   "add new policy",
			policy: policy,
			page:   policies.PolicyPage{},
			token:  generateValidToken(t, testsutil.GenerateUUID(t, idProvider), csvc, cRepo),
			err:    nil,
		},
		{
			desc:   "add existing policy",
			policy: policy,
			page:   policies.PolicyPage{Policies: []policies.Policy{policy}},
			token:  generateValidToken(t, testsutil.GenerateUUID(t, idProvider), csvc, cRepo),
			err:    errors.ErrConflict,
		},
		{
			desc: "add a new policy with owner",
			page: policies.PolicyPage{},
			policy: policies.Policy{
				OwnerID: testsutil.GenerateUUID(t, idProvider),
				Object:  "objwithowner",
				Actions: []string{"m_read"},
				Subject: "subwithowner",
			},
			err:   nil,
			token: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), csvc, cRepo),
		},
		{
			desc: "add a new policy with more actions",
			page: policies.PolicyPage{},
			policy: policies.Policy{
				Object:  "obj2",
				Actions: []string{"c_delete", "c_update", "c_add", "c_list"},
				Subject: "sub2",
			},
			err:   nil,
			token: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), csvc, cRepo),
		},
		{
			desc: "add a new policy with wrong action",
			page: policies.PolicyPage{},
			policy: policies.Policy{
				Object:  "obj3",
				Actions: []string{"wrong"},
				Subject: "sub3",
			},
			err:   apiutil.ErrMissingPolicyAct,
			token: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), csvc, cRepo),
		},
		{
			desc: "add a new policy with empty object",
			page: policies.PolicyPage{},
			policy: policies.Policy{
				Actions: []string{"c_delete"},
				Subject: "sub4",
			},
			err:   apiutil.ErrMissingPolicyObj,
			token: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), csvc, cRepo),
		},
		{
			desc: "add a new policy with empty subject",
			page: policies.PolicyPage{},
			policy: policies.Policy{
				Actions: []string{"c_delete"},
				Object:  "obj4",
			},
			err:   apiutil.ErrMissingPolicySub,
			token: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), csvc, cRepo),
		},
		{
			desc: "add a new policy with empty action",
			page: policies.PolicyPage{},
			policy: policies.Policy{
				Subject: "sub5",
				Object:  "obj5",
			},
			err:   apiutil.ErrMissingPolicyAct,
			token: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), csvc, cRepo),
		},
	}

	for _, tc := range cases {
		repo1Call := pRepo.On("Evaluate", context.Background(), "client", mock.Anything).Return(nil)
		repoCall := pRepo.On("Update", context.Background(), tc.policy).Return(tc.err)
		repoCall1 := pRepo.On("Save", context.Background(), mock.Anything).Return(tc.err)
		repoCall2 := pRepo.On("Retrieve", context.Background(), mock.Anything).Return(tc.page, nil)
		err := svc.AddPolicy(context.Background(), tc.token, tc.policy)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		if err == nil {
			tc.policy.Subject = tc.token
			err = svc.Authorize(context.Background(), "client", tc.policy)
			require.Nil(t, err, fmt.Sprintf("checking shared %v policy expected to be succeed: %#v", tc.policy, err))
		}
		repoCall1.Parent.AssertCalled(t, "Save", context.Background(), mock.Anything)
		repoCall.Unset()
		repoCall1.Unset()
		repoCall2.Unset()
		repo1Call.Unset()
	}

}

func TestAuthorize(t *testing.T) {
	cRepo := new(cmocks.ClientRepository)
	pRepo := new(pmocks.PolicyRepository)
	tokenizer := jwt.NewTokenRepo([]byte(secret))
	csvc := clients.NewService(cRepo, pRepo, tokenizer, phasher, idProvider, passRegex)
	svc := policies.NewService(pRepo, tokenizer, idProvider)

	cases := []struct {
		desc   string
		policy policies.Policy
		domain string
		err    error
	}{
		{
			desc:   "check valid policy in client domain",
			policy: policies.Policy{Object: "client1", Actions: []string{"c_update"}, Subject: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), csvc, cRepo)},
			domain: "client",
			err:    nil,
		},
		{
			desc:   "check valid policy in group domain",
			policy: policies.Policy{Object: "client1", Actions: []string{"g_update"}, Subject: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), csvc, cRepo)},
			domain: "group",
			err:    errors.ErrConflict,
		},
		{
			desc:   "check invalid policy in client domain",
			policy: policies.Policy{Object: "client3", Actions: []string{"c_update"}, Subject: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), csvc, cRepo)},
			domain: "client",
			err:    nil,
		},
		{
			desc:   "check invalid policy in group domain",
			policy: policies.Policy{Object: "client3", Actions: []string{"g_update"}, Subject: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), csvc, cRepo)},
			domain: "group",
			err:    nil,
		},
	}

	for _, tc := range cases {
		repoCall := pRepo.On("Evaluate", context.Background(), tc.domain, mock.Anything).Return(tc.err)
		err := svc.Authorize(context.Background(), tc.domain, tc.policy)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		repoCall.Unset()
	}

}

func TestDeletePolicy(t *testing.T) {
	cRepo := new(cmocks.ClientRepository)
	pRepo := new(pmocks.PolicyRepository)
	tokenizer := jwt.NewTokenRepo([]byte(secret))
	csvc := clients.NewService(cRepo, pRepo, tokenizer, phasher, idProvider, passRegex)
	svc := policies.NewService(pRepo, tokenizer, idProvider)

	pr := policies.Policy{Object: authoritiesObj, Actions: memberActions, Subject: testsutil.GenerateUUID(t, idProvider)}

	repoCall := pRepo.On("Delete", context.Background(), mock.Anything).Return(nil)
	repoCall1 := pRepo.On("Retrieve", context.Background(), mock.Anything).Return(policies.PolicyPage{Policies: []policies.Policy{pr}}, nil)
	err := svc.DeletePolicy(context.Background(), generateValidToken(t, testsutil.GenerateUUID(t, idProvider), csvc, cRepo), pr)
	require.Nil(t, err, fmt.Sprintf("deleting %v policy expected to succeed: %s", pr, err))
	repoCall.Unset()
	repoCall1.Unset()
}

func TestListPolicies(t *testing.T) {
	cRepo := new(cmocks.ClientRepository)
	pRepo := new(pmocks.PolicyRepository)
	tokenizer := jwt.NewTokenRepo([]byte(secret))
	csvc := clients.NewService(cRepo, pRepo, tokenizer, phasher, idProvider, passRegex)
	svc := policies.NewService(pRepo, tokenizer, idProvider)

	id := testsutil.GenerateUUID(t, idProvider)

	readPolicy := "m_read"
	writePolicy := "m_write"

	var nPolicy = uint64(10)
	var aPolicies = []policies.Policy{}
	for i := uint64(0); i < nPolicy; i++ {
		pr := policies.Policy{
			OwnerID: id,
			Actions: []string{readPolicy},
			Subject: fmt.Sprintf("thing_%d", i),
			Object:  fmt.Sprintf("client_%d", i),
		}
		if i%3 == 0 {
			pr.Actions = []string{writePolicy}
		}
		aPolicies = append(aPolicies, pr)
	}

	cases := []struct {
		desc     string
		token    string
		page     policies.Page
		response policies.PolicyPage
		err      error
	}{
		{
			desc:  "list policies with authorized token",
			token: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), csvc, cRepo),
			err:   nil,
			response: policies.PolicyPage{
				Page: policies.Page{
					Offset: 0,
					Total:  nPolicy,
				},
				Policies: aPolicies,
			},
		},
		{
			desc:  "list policies with invalid token",
			token: inValidToken,
			err:   errors.ErrAuthentication,
			response: policies.PolicyPage{
				Page: policies.Page{
					Offset: 0,
				},
			},
		},
		{
			desc:  "list policies with offset and limit",
			token: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), csvc, cRepo),
			page: policies.Page{
				Offset: 6,
				Limit:  nPolicy,
			},
			response: policies.PolicyPage{
				Page: policies.Page{
					Offset: 6,
					Total:  nPolicy,
				},
				Policies: aPolicies[6:10],
			},
		},
		{
			desc:  "list policies with wrong action",
			token: generateValidToken(t, testsutil.GenerateUUID(t, idProvider), csvc, cRepo),
			page: policies.Page{
				Action: "wrong",
			},
			response: policies.PolicyPage{},
			err:      apiutil.ErrMissingPolicyAct,
		},
	}

	for _, tc := range cases {
		repoCall := pRepo.On("Retrieve", context.Background(), tc.page).Return(tc.response, tc.err)
		page, err := svc.ListPolicy(context.Background(), tc.token, tc.page)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		assert.Equal(t, tc.response, page, fmt.Sprintf("%s: expected size %v got %v\n", tc.desc, tc.response, page))
		repoCall.Unset()
	}

}

func TestUpdatePolicies(t *testing.T) {
	cRepo := new(cmocks.ClientRepository)
	pRepo := new(pmocks.PolicyRepository)
	tokenizer := jwt.NewTokenRepo([]byte(secret))
	csvc := clients.NewService(cRepo, pRepo, tokenizer, phasher, idProvider, passRegex)
	svc := policies.NewService(pRepo, tokenizer, idProvider)

	policy := policies.Policy{Object: "obj1", Actions: []string{"m_read"}, Subject: "sub1"}

	cases := []struct {
		desc   string
		action []string
		token  string
		err    error
	}{
		{
			desc:   "update policy actions with valid token",
			action: []string{"m_write"},
			token:  generateValidToken(t, testsutil.GenerateUUID(t, idProvider), csvc, cRepo),
			err:    nil,
		},
		{
			desc:   "update policy action with invalid token",
			action: []string{"m_write"},
			token:  "non-existent",
			err:    errors.ErrAuthentication,
		},
		{
			desc:   "update policy action with wrong policy action",
			action: []string{"wrong"},
			token:  generateValidToken(t, testsutil.GenerateUUID(t, idProvider), csvc, cRepo),
			err:    apiutil.ErrMissingPolicyAct,
		},
	}

	for _, tc := range cases {
		policy.Actions = tc.action
		repoCall := pRepo.On("Retrieve", context.Background(), mock.Anything).Return(policies.PolicyPage{Policies: []policies.Policy{policy}}, nil)
		repoCall1 := pRepo.On("Update", context.Background(), mock.Anything).Return(tc.err)
		err := svc.UpdatePolicy(context.Background(), tc.token, policy)
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
		repoCall1.Parent.AssertCalled(t, "Update", context.Background(), mock.Anything)
		repoCall.Unset()
		repoCall1.Unset()
	}
}
