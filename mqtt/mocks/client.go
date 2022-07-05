package mocks

import (
	"context"

	"github.com/mainflux/mainflux/pkg/auth"
	"github.com/mainflux/mainflux/pkg/errors"
)

type MockClient struct {
	key map[string]string
}

func NewClient(key map[string]string) auth.Client {
	return MockClient{key: key}
}

func (cli MockClient) Authorize(ctx context.Context, chanID, thingID string) error {
	return nil
}

func (cli MockClient) Identify(ctx context.Context, thingKey string) (string, error) {
	if id, ok := cli.key[thingKey]; ok {
		return id, nil
	}
	return "", errors.ErrAuthentication
}
