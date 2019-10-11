//
// Copyright (c) 2019
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package twins

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/mainflux/mainflux"
	"github.com/nats-io/go-nats"
)

var (
	// ErrMalformedEntity indicates malformed entity specification (e.g.
	// invalid username or password).
	ErrMalformedEntity = errors.New("malformed entity specification")

	// ErrUnauthorizedAccess indicates missing or invalid credentials provided
	// when accessing a protected resource.
	ErrUnauthorizedAccess = errors.New("missing or invalid credentials provided")

	// ErrNotFound indicates a non-existent entity request.
	ErrNotFound = errors.New("non-existent entity")

	// ErrConflict indicates that entity already exists.
	ErrConflict = errors.New("entity already exists")
)

// Service specifies an API that must be fullfiled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	// Ping compares a given string with secret
	Ping(string) (string, error)

	// AddTwin adds new twin to the user identified by the provided key.
	AddTwin(context.Context, string, Twin) (Twin, error)

	// UpdateTwin updates the twin identified by the provided ID, that
	// belongs to the user identified by the provided key.
	UpdateTwin(context.Context, string, Twin) error

	// UpdateKey updates key value of the existing twin. A non-nil error is
	// returned to indicate operation failure.
	UpdateKey(context.Context, string, string, string) error

	// ViewTwin retrieves data about the twin identified with the provided
	// ID, that belongs to the user identified by the provided key.
	ViewTwin(context.Context, string, string) (Twin, error)

	// ListTwinsByChannel retrieves data about subset of twins that are
	// connected to specified channel and belong to the user identified by
	// the provided key.
	ListTwinsByChannel(context.Context, string, string, uint64) (TwinsSet, error)

	// RemoveTwin removes the twin identified with the provided ID, that
	// belongs to the user identified by the provided key.
	RemoveTwin(context.Context, string, string) error
}

type twinsService struct {
	secret     string
	natsClient *nats.Conn
	mqttClient mqtt.Client
	users      mainflux.UsersServiceClient
	twins      TwinRepository
	idp        IdentityProvider
}

var _ Service = (*twinsService)(nil)

// New instantiates the twins service implementation.
func New(secret string, nc *nats.Conn, mc mqtt.Client, users mainflux.UsersServiceClient, twins TwinRepository, idp IdentityProvider) Service {
	return &twinsService{
		secret:     secret,
		mqttClient: mc,
		natsClient: nc,
		users:      users,
		twins:      twins,
		idp:        idp,
	}
}

func (ts *twinsService) Ping(secret string) (string, error) {
	if ts.secret != secret {
		return "", ErrUnauthorizedAccess
	}
	return "Hello World :)", nil
}

func (ts *twinsService) AddTwin(ctx context.Context, token string, twin Twin) (Twin, error) {
	res, err := ts.users.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Twin{}, ErrUnauthorizedAccess
	}

	twin.ID, err = ts.idp.ID()
	if err != nil {
		return Twin{}, err
	}

	twin.Owner = res.GetValue()

	if twin.Key == "" {
		twin.Key, err = ts.idp.ID()
		if err != nil {
			return Twin{}, err
		}
	}

	twin.created = time.Now()
	twin.updated = time.Now()

	id, err := ts.twins.Save(ctx, twin)
	if err != nil {
		return Twin{}, err
	}

	twin.ID = id

	return twin, nil
}

func (ts *twinsService) UpdateTwin(ctx context.Context, token string, twin Twin) error {
	res, err := ts.users.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return ErrUnauthorizedAccess
	}

	twin.Owner = res.GetValue()

	twin.updated = time.Now()
	if err := ts.twins.Update(ctx, twin); err != nil {
		return err
	}

	b, err := json.Marshal(twin)
	if ts.publish(twin.thingID, "update/success", b); err != nil {
		return err
	}

	return nil
}

func (ts *twinsService) UpdateKey(ctx context.Context, token, id, key string) error {
	res, err := ts.users.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return ErrUnauthorizedAccess
	}

	if err := ts.twins.UpdateKey(ctx, res.GetValue(), id, key); err != nil {
		return err
	}

	return nil
}

func (ts *twinsService) ViewTwin(ctx context.Context, token, id string) (Twin, error) {
	res, err := ts.users.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Twin{}, ErrUnauthorizedAccess
	}

	twin, err := ts.twins.RetrieveByID(ctx, res.GetValue(), id)
	if err != nil {
		return Twin{}, err
	}

	b, err := json.Marshal(twin)
	if ts.publish(twin.thingID, "get/success", b); err != nil {
		return Twin{}, err
	}

	return twin, nil
}

func (ts *twinsService) ListTwinsByChannel(ctx context.Context, token, channel string, limit uint64) (TwinsSet, error) {
	_, err := ts.users.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return TwinsSet{}, ErrUnauthorizedAccess
	}

	return ts.twins.RetrieveByChannel(ctx, channel, limit)
}

func (ts *twinsService) RemoveTwin(ctx context.Context, token, id string) error {
	res, err := ts.users.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return ErrUnauthorizedAccess
	}

	if err := ts.twins.Remove(ctx, res.GetValue(), id); err != nil {
		return err
	}

	return nil
}

func (ts *twinsService) publish(thingID, op string, payload []byte) error {
	topic := fmt.Sprintf("$mfx/things/%s/%s", thingID, op)
	token := ts.mqttClient.Publish(topic, 0, false, payload)
	token.Wait()

	return token.Error()
}
