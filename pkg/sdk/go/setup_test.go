package sdk_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/mainflux/mainflux/clients/clients"
	cmocks "github.com/mainflux/mainflux/clients/clients/mocks"
	"github.com/mainflux/mainflux/clients/hasher"
	"github.com/mainflux/mainflux/pkg/errors"
	sdk "github.com/mainflux/mainflux/pkg/sdk/go"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	invalidIdentity = "invalididentity"
	Identity        = "identity"
	secret          = "strongsecret"
	token           = "token"
	invalidToken    = "invalidtoken"
)

var (
	idProvider    = uuid.New()
	phasher       = hasher.New()
	validMetadata = sdk.Metadata{"role": "client"}
	client        = sdk.User{
		Name:        "clientname",
		Tags:        []string{"tag1", "tag2"},
		Credentials: sdk.Credentials{Identity: "clientidentity", Secret: secret},
		Metadata:    validMetadata,
		Status:      clients.EnabledStatus.String(),
	}
	description    = "shortdescription"
	gName          = "groupname"
	authoritiesObj = "authorities"
	subject        = generateUUID(&testing.T{})
	object         = generateUUID(&testing.T{})
)

func generateValidToken(t *testing.T, svc clients.Service, cRepo *cmocks.ClientRepository) string {
	client := clients.Client{
		ID:   generateUUID(t),
		Name: "validtoken",
		Credentials: clients.Credentials{
			Identity: "validtoken",
			Secret:   secret,
		},
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

func TestMain(m *testing.M) {
	exitCode := m.Run()
	os.Exit(exitCode)
}
