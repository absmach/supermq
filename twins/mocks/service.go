package mocks

import (
	"github.com/mainflux/mainflux/twins"
)

func NewService(tokens map[string]string) twins.Service {
	auth := NewAuthNServiceClient(tokens)
	twinsRepo := NewTwinRepository()
	statesRepo := NewStateRepository()
	idp := NewIdentityProvider()
	subs := map[string]string{"chanID": "chanID"}
	broker := New(subs)
	return twins.New(broker, auth, twinsRepo, statesRepo, idp, "chanID", nil)
}
