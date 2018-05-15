package mocks

import (
	"github.com/mainflux/mainflux/things"
)

var _ things.IdentityProvider = (*identityProviderMock)(nil)

type identityProviderMock struct{}

func (idp *identityProviderMock) ID() string {
	// TODO: prepare mock for usage in tests
	return ""
}

// NewIdentityProvider creates "mirror" identity provider, i.e. generated
// token will hold value provided by the caller.
func NewIdentityProvider() things.IdentityProvider {
	return &identityProviderMock{}
}
