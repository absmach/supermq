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
	"reflect"
	"time"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/twins/paho"
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
	AddTwin(context.Context, string, Twin, Definition) (Twin, error)

	// UpdateTwin updates twin identified by the provided Twin that
	// belongs to the user identified by the provided key.
	UpdateTwin(context.Context, string, Twin, Definition) error

	// ViewTwin retrieves data about twin with the provided
	// ID belonging to the user identified by the provided key.
	ViewTwin(context.Context, string, string) (Twin, error)

	// ListTwins retrieves data about subset of twins that belongs to the
	// user identified by the provided key.
	ListTwins(context.Context, string, uint64, uint64, string, Metadata) (TwinsPage, error)

	// ListStates retrieves data about subset of states that belongs to the
	// twin identified by the id.
	ListStates(context.Context, string, uint64, uint64, string) (StatesPage, error)

	// ListTwinsByThing retrieves data about subset of twins that represent
	// specified thing belong to the user identified by
	// the provided key.
	ViewTwinByThing(context.Context, string, string) (Twin, error)

	// RemoveTwin removes the twin identified with the provided ID, that
	// belongs to the user identified by the provided key.
	RemoveTwin(context.Context, string, string) error
}

// type mqtt struct {
// 	client paho.Client
// 	topic  string
// }

type twinsService struct {
	natsClient *nats.Conn
	mqttClient paho.Mqtt
	users      mainflux.UsersServiceClient
	twins      TwinRepository
	states     StateRepository
	idp        IdentityProvider
}

var _ Service = (*twinsService)(nil)

// New instantiates the twins service implementation.
func New(nc *nats.Conn, mc paho.Mqtt, users mainflux.UsersServiceClient, twins TwinRepository, sr StateRepository, idp IdentityProvider) Service {
	return &twinsService{
		natsClient: nc,
		mqttClient: mc,
		users:      users,
		twins:      twins,
		states:     sr,
		idp:        idp,
	}
}

func (ts *twinsService) AddTwin(ctx context.Context, token string, twin Twin, def Definition) (tw Twin, err error) {
	id := ""
	b := []byte{}
	defer ts.mqttClient.Publish(&id, &err, "create/success", "create/failure", &b)

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

	if isZeroOfUnderlyingType(def) {
		def = Definition{}
		def.Attributes = make(map[string]Attribute)
	}
	def.Created = time.Now()
	def.ID = 0
	twin.Definitions = append(twin.Definitions, def)

	twin.Revision = 0
	_, err = ts.twins.Save(ctx, twin)
	if err != nil {
		return Twin{}, err
	}

	id = twin.ID
	b, err = json.Marshal(twin)

	return twin, nil
}

func isZeroOfUnderlyingType(x interface{}) bool {
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}

func (ts *twinsService) UpdateTwin(ctx context.Context, token string, twin Twin, def Definition) (err error) {
	b := []byte{}
	id := ""
	defer ts.mqttClient.Publish(&id, &err, "update/success", "update/failure", &b)

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

	if !isZeroOfUnderlyingType(twin.Key) {
		tw.Key = twin.Key
	}

	if !isZeroOfUnderlyingType(twin.Name) {
		tw.Name = twin.Name
	}

	if !isZeroOfUnderlyingType(twin.ThingID) {
		tw.ThingID = twin.ThingID
	}

	if !isZeroOfUnderlyingType(def) {
		def.Created = time.Now()
		def.ID = tw.Definitions[len(tw.Definitions)-1].ID + 1
		tw.Definitions = append(tw.Definitions, def)
	}

	if !isZeroOfUnderlyingType(twin.Metadata) {
		tw.Metadata = twin.Metadata
	}

	if err := ts.twins.Update(ctx, tw); err != nil {
		return err
	}

	id = twin.ID
	b, err = json.Marshal(tw)

	return nil
}

func (ts *twinsService) ViewTwin(ctx context.Context, token, id string) (tw Twin, err error) {
	b := []byte{}
	defer ts.mqttClient.Publish(&id, &err, "get/success", "get/failure", &b)

	res, err := ts.users.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Twin{}, ErrUnauthorizedAccess
	}

	twin, err := ts.twins.RetrieveByID(ctx, res.GetValue(), id)
	if err != nil {
		return Twin{}, err
	}

	b, err = json.Marshal(twin)

	return twin, nil
}

func (ts *twinsService) ViewTwinByThing(ctx context.Context, token, thingid string) (Twin, error) {
	_, err := ts.users.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Twin{}, ErrUnauthorizedAccess
	}

	return ts.twins.RetrieveByThing(ctx, thingid)
}

func (ts *twinsService) RemoveTwin(ctx context.Context, token, id string) (err error) {
	b := []byte{}
	defer ts.mqttClient.Publish(&id, &err, "remove/success", "remove/failure", &b)

	res, err := ts.users.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return ErrUnauthorizedAccess
	}

	if err := ts.twins.Remove(ctx, res.GetValue(), id); err != nil {
		return err
	}

	return nil
}

func (ts *twinsService) ListTwins(ctx context.Context, token string, offset uint64, limit uint64, name string, metadata Metadata) (TwinsPage, error) {
	res, err := ts.users.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return TwinsPage{}, ErrUnauthorizedAccess
	}

	return ts.twins.RetrieveAll(ctx, res.GetValue(), offset, limit, name, metadata)
}

func (ts *twinsService) ListStates(ctx context.Context, token string, offset uint64, limit uint64, id string) (StatesPage, error) {
	_, err := ts.users.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return StatesPage{}, ErrUnauthorizedAccess
	}

	return ts.states.RetrieveAll(ctx, offset, limit, id)
}
