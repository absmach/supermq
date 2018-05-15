package mocks

import (
	"github.com/mainflux/mainflux/clients"
	uuid "github.com/satori/go.uuid"
)

var _ clients.IdentityProvider = (*identityProviderMock)(nil)

type identityProviderMock struct{}

// NewIdentityProvider creates "mirror" identity provider, i.e. generated
// token will hold value provided by the caller.
func NewIdentityProvider() clients.IdentityProvider {
	return &identityProviderMock{}
}

func (idp *identityProviderMock) Key() string {
	return uuid.NewV4().String()
}

func (idp *identityProviderMock) Identity(key string) (string, error) {
	return idp.Key(), nil
}
