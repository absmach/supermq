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
	"reflect"
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
	// AddTwin adds new twin to the user identified by the provided key.
	AddTwin(context.Context, string, Twin) (Twin, error)

	// UpdateTwin updates twin identified by the provided Twin that
	// belongs to the user identified by the provided key.
	UpdateTwin(context.Context, string, Twin) error

	// ViewTwin retrieves data about twin with the provided
	// ID belonging to the user identified by the provided key.
	ViewTwin(context.Context, string, string) (Twin, error)

	// ListTwins retrieves data about subset of twins that belongs to the
	// user identified by the provided key.
	ListTwins(context.Context, string, uint64, string, Metadata) (TwinsSet, error)

	// ListTwinsByThing retrieves data about subset of twins that represent
	// specified thing belong to the user identified by
	// the provided key.
	ListTwinsByThing(context.Context, string, string, uint64) (TwinsSet, error)

	// RemoveTwin removes the twin identified with the provided ID, that
	// belongs to the user identified by the provided key.
	RemoveTwin(context.Context, string, string) error
}

type twinsService struct {
	natsClient *nats.Conn
	mqttClient mqtt.Client
	mqttTopic  string
	users      mainflux.UsersServiceClient
	twins      TwinRepository
	idp        IdentityProvider
}

var _ Service = (*twinsService)(nil)

// New instantiates the twins service implementation.
func New(nc *nats.Conn, mc mqtt.Client, topic string, users mainflux.UsersServiceClient, twins TwinRepository, idp IdentityProvider) Service {
	return &twinsService{
		natsClient: nc,
		mqttClient: mc,
		mqttTopic:  topic,
		users:      users,
		twins:      twins,
		idp:        idp,
	}

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

	twin.Created = time.Now()
	twin.Updated = time.Now()

	id, err := ts.twins.Save(ctx, twin)
	if err != nil {
		return Twin{}, err
	}

	twin.ID = id
	twin.Revision = 0

	b, err := json.Marshal(twin)
	if ts.publish(twin.ID, "create/success", b); err != nil {
		return Twin{}, err
	}

	return twin, nil
}

func IsZeroOfUnderlyingType(x interface{}) bool {
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}

func (ts *twinsService) UpdateTwin(ctx context.Context, token string, twin Twin) error {
	res, err := ts.users.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return ErrUnauthorizedAccess
	}

	tw, err := ts.twins.RetrieveByID(ctx, res.GetValue(), twin.ID)
	if err != nil {
		return err
	}
	tw.Updated = time.Now()
	tw.Revision++

	if !IsZeroOfUnderlyingType(twin.Key) {
		tw.Key = twin.Key
	}
	if !IsZeroOfUnderlyingType(twin.Name) {
		tw.Name = twin.Name
	}
	if !IsZeroOfUnderlyingType(twin.ThingID) {
		tw.ThingID = twin.ThingID
	}
	if !IsZeroOfUnderlyingType(twin.Attributes) {
		tw.Attributes = twin.Attributes
	}
	if !IsZeroOfUnderlyingType(twin.State) {
		tw.State = twin.State
	}
	if !IsZeroOfUnderlyingType(twin.Metadata) {
		tw.Metadata = twin.Metadata
	}

	if err := ts.twins.Update(ctx, tw); err != nil {
		return err
	}

	b, err := json.Marshal(tw)
	if ts.publish(twin.ID, "update/success", b); err != nil {
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
	if ts.publish(twin.ID, "get/success", b); err != nil {
		return Twin{}, err
	}

	return twin, nil
}

func (ts *twinsService) ListTwins(ctx context.Context, token string, limit uint64, name string, metadata Metadata) (TwinsSet, error) {
	res, err := ts.users.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return TwinsSet{}, ErrUnauthorizedAccess
	}

	return ts.twins.RetrieveAll(ctx, res.GetValue(), limit, name, metadata)
}

func (ts *twinsService) ListTwinsByThing(ctx context.Context, token, thing string, limit uint64) (TwinsSet, error) {
	_, err := ts.users.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return TwinsSet{}, ErrUnauthorizedAccess
	}

	return ts.twins.RetrieveByThing(ctx, thing, limit)
}

func (ts *twinsService) RemoveTwin(ctx context.Context, token, id string) error {
	res, err := ts.users.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return ErrUnauthorizedAccess
	}

	if err := ts.twins.Remove(ctx, res.GetValue(), id); err != nil {
		return err
	}

	if ts.publish(id, "remove/success", []byte{}); err != nil {
		return err
	}

	return nil
}

func (ts *twinsService) publish(id, op string, payload []byte) error {
	topic := fmt.Sprintf("channels/%s/messages/%s/%s", ts.mqttTopic, id, op)

	token := ts.mqttClient.Publish(topic, 0, false, payload)
	token.Wait()

	return token.Error()
}
