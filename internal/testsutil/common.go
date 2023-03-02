package testsutil

import (
	"context"
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/users/clients"
	cmocks "github.com/mainflux/mainflux/users/clients/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func GenerateUUID(t *testing.T, idProvider mainflux.IDProvider) string {
	ulid, err := idProvider.ID()
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
	return ulid
}

func GenerateValidToken(t *testing.T, clientID string, svc clients.Service, cRepo *cmocks.ClientRepository, phasher clients.Hasher) string {
	client := clients.Client{
		ID:   clientID,
		Name: "validtoken",
		Credentials: clients.Credentials{
			Identity: "validtoken",
			Secret:   "secret",
		},
		Status: clients.EnabledStatus,
	}
	rClient := client
	rClient.Credentials.Secret, _ = phasher.Hash(client.Credentials.Secret)

	repoCall := cRepo.On("RetrieveByIdentity", context.Background(), client.Credentials.Identity).Return(rClient, nil)
	token, err := svc.IssueToken(context.Background(), client.Credentials.Identity, client.Credentials.Secret)
	assert.True(t, errors.Contains(err, nil), fmt.Sprintf("Create token expected nil got %s\n", err))
	ok := repoCall.Parent.AssertCalled(t, "RetrieveByIdentity", context.Background(), client.Credentials.Identity)
	assert.True(t, ok, "RetrieveByIdentity was not called on creating token")
	repoCall.Unset()
	return token.AccessToken
}

func CleanUpDB(t *testing.T, db *sqlx.DB) {
	_, err := db.Exec("DELETE FROM policies")
	require.Nil(t, err, fmt.Sprintf("clean policies unexpected error: %s", err))
	_, err = db.Exec("DELETE FROM groups")
	require.Nil(t, err, fmt.Sprintf("clean groups unexpected error: %s", err))
	_, err = db.Exec("DELETE FROM clients")
	require.Nil(t, err, fmt.Sprintf("clean clients unexpected error: %s", err))
}


curl --request POST --url 'https://transcribe.whisperapi.com' --header 'Authorization: Bearer 9ZHL6QGV31CCIQGNPA2SELYHNIQF6E8Q' -F "file=@YOUR_FILE_PATH" -F "diarization=false" -F "numSpeakers=1" -F "fileType=YOUR_FILE_TYPE" -F "language=en" -F "task=transcribe"
