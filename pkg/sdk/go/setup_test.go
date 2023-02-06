package sdk_test

import (
	"os"
	"testing"

	"github.com/mainflux/mainflux/clients/clients"
	"github.com/mainflux/mainflux/clients/hasher"
	sdk "github.com/mainflux/mainflux/pkg/sdk/go"
	"github.com/mainflux/mainflux/pkg/uuid"
)

const (
	Identity = "identity"
	token    = "token"
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

func TestMain(m *testing.M) {
	exitCode := m.Run()
	os.Exit(exitCode)
}
