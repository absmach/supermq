package mocks

import (
	"context"

	"github.com/mainflux/mainflux/pkg/auth"
	"github.com/mainflux/mainflux/pkg/errors"
)

type MockClient struct{}

func NewClient() auth.Client {
	return MockClient{}
}

func (cli MockClient) Authorize(ctx context.Context, chanID, thingID string) error {
	return nil
}

func (cli MockClient) Identify(ctx context.Context, thingKey string) (string, error) {
	if thingKey == "" {
		return "", errors.ErrAuthentication
	}
	return "ok", nil
}
