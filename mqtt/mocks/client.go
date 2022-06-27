package mocks

import (
	"context"
	"errors"

	"github.com/mainflux/mainflux/pkg/auth"
)

var errIdentify = errors.New("thing identify error")

type MockClient struct{}

func NewClient() auth.Client {
	return MockClient{}
}

func (cli MockClient) Authorize(ctx context.Context, chanID, thingID string) error {
	return nil
}

func (cli MockClient) Identify(ctx context.Context, thingKey string) (string, error) {

	if thingKey == "" {
		return "", errIdentify
	}
	return "ok", nil
}
