package sdk_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/mainflux/mainflux/pkg/errors"
	sdk "github.com/mainflux/mainflux/pkg/sdk/go"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/mainflux/mainflux/users/clients"
	umocks "github.com/mainflux/mainflux/users/clients/mocks"
	"github.com/mainflux/mainflux/users/groups"
	"github.com/mainflux/mainflux/users/hasher"
	"github.com/mainflux/mainflux/users/policies"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	invalidIdentity = "invalididentity"
	Identity        = "identity"
	secret          = "strongsecret"
	token           = "token"
	invalidToken    = "invalidtoken"
	contentType     = "application/senml+json"
)

var (
	idProvider    = uuid.New()
	phasher       = hasher.New()
	validMetadata = sdk.Metadata{"role": "client"}
	user          = sdk.User{
		Name:        "clientname",
		Tags:        []string{"tag1", "tag2"},
		Credentials: sdk.Credentials{Identity: "clientidentity", Secret: secret},
		Metadata:    validMetadata,
		Status:      clients.EnabledStatus.String(),
	}
	description = "shortdescription"
	gName       = "groupname"

	limit  uint64 = 5
	offset uint64 = 0
	total  uint64 = 200

	authoritiesObj  = "authorities"
	subject         = generateUUID(&testing.T{})
	object          = generateUUID(&testing.T{})
	emailer         = umocks.NewEmailer()
	passRegex       = regexp.MustCompile("^.{8,}$")
	accessDuration  = time.Minute * 1
	refreshDuration = time.Minute * 10
)

func generateValidToken(t *testing.T, svc clients.Service, cRepo *umocks.ClientRepository) string {
	client := clients.Client{
		ID:   generateUUID(t),
		Name: "validtoken",
		Credentials: clients.Credentials{
			Identity: "validtoken",
			Secret:   secret,
		},
		Role:   clients.AdminRole,
		Status: clients.EnabledStatus,
	}
	rclient := client
	rclient.Credentials.Secret, _ = phasher.Hash(client.Credentials.Secret)

	repoCall := cRepo.On("RetrieveByIdentity", context.Background(), mock.Anything).Return(rclient, nil)
	token, err := svc.IssueToken(context.Background(), client.Credentials.Identity, client.Credentials.Secret)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("Create token expected nil got %s\n", err))
	repoCall.Unset()
	return token.AccessToken
}

func generateUUID(t *testing.T) string {
	ulid, err := idProvider.ID()
	assert.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
	return ulid
}

func convertClientsPage(cp sdk.UsersPage) clients.ClientsPage {
	return clients.ClientsPage{
		Clients: convertClients(cp.Users),
	}
}

func convertClients(cs []sdk.User) []clients.Client {
	ccs := []clients.Client{}

	for _, c := range cs {
		ccs = append(ccs, convertClient(c))
	}

	return ccs
}

func convertGroups(cs []sdk.Group) []groups.Group {
	cgs := []groups.Group{}

	for _, c := range cs {
		cgs = append(cgs, convertGroup(c))
	}

	return cgs
}

func convertPolicies(cs []sdk.Policy) []policies.Policy {
	ccs := []policies.Policy{}

	for _, c := range cs {
		ccs = append(ccs, convertPolicy(c))
	}

	return ccs
}

func convertPolicy(sp sdk.Policy) policies.Policy {
	return policies.Policy{
		OwnerID:   sp.OwnerID,
		Subject:   sp.Subject,
		Object:    sp.Object,
		Actions:   sp.Actions,
		CreatedAt: sp.CreatedAt,
		UpdatedAt: sp.UpdatedAt,
	}
}

func convertMembershipsPage(m sdk.MembershipsPage) groups.MembershipsPage {
	return groups.MembershipsPage{
		Page: groups.Page{
			Limit:  m.Limit,
			Total:  m.Total,
			Offset: m.Offset,
		},
		Memberships: convertMemberships(m.Memberships),
	}
}

func convertClientPage(p sdk.PageMetadata) clients.Page {
	if p.Status == "" {
		p.Status = clients.EnabledStatus.String()
	}
	status, err := clients.ToStatus(p.Status)
	if err != nil {
		return clients.Page{}
	}
	return clients.Page{
		Status:   status,
		Total:    p.Total,
		Offset:   p.Offset,
		Limit:    p.Limit,
		Name:     p.Name,
		Action:   p.Action,
		Tag:      p.Tag,
		Metadata: clients.Metadata(p.Metadata),
	}
}

func convertMemberships(gs []sdk.Group) []groups.Group {
	cg := []groups.Group{}
	for _, g := range gs {
		cg = append(cg, convertGroup(g))
	}

	return cg
}

func convertGroup(g sdk.Group) groups.Group {
	if g.Status == "" {
		g.Status = groups.EnabledStatus.String()
	}
	status, err := groups.ToStatus(g.Status)
	if err != nil {
		return groups.Group{}
	}
	return groups.Group{
		ID:          g.ID,
		OwnerID:     g.OwnerID,
		ParentID:    g.ParentID,
		Name:        g.Name,
		Description: g.Description,
		Metadata:    groups.Metadata(g.Metadata),
		Level:       g.Level,
		Path:        g.Path,
		Children:    convertChildren(g.Children),
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
		Status:      status,
	}
}

func convertChildren(gs []*sdk.Group) []*groups.Group {
	cg := []*groups.Group{}

	if len(gs) == 0 {
		return cg
	}

	for _, g := range gs {
		insert := convertGroup(*g)
		cg = append(cg, &insert)
	}

	return cg
}

func convertClient(c sdk.User) clients.Client {
	if c.Status == "" {
		c.Status = clients.EnabledStatus.String()
	}
	status, err := clients.ToStatus(c.Status)
	if err != nil {
		return clients.Client{}
	}
	return clients.Client{
		ID:          c.ID,
		Name:        c.Name,
		Tags:        c.Tags,
		Owner:       c.Owner,
		Credentials: clients.Credentials(c.Credentials),
		Metadata:    clients.Metadata(c.Metadata),
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
		Status:      status,
	}
}

func convertPolicyPage(pp sdk.PolicyPage) policies.PolicyPage {
	return policies.PolicyPage{
		Page: policies.Page{
			Limit:  pp.Limit,
			Total:  pp.Total,
			Offset: pp.Offset,
		},
		Policies: convertPolicies(pp.Policies),
	}
}

func TestMain(m *testing.M) {
	exitCode := m.Run()
	os.Exit(exitCode)
}
