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

	paho "github.com/eclipse/paho.mqtt.golang"
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
	AddTwin(context.Context, string, Twin, Definition) (Twin, error)

	// UpdateTwin updates twin identified by the provided Twin that
	// belongs to the user identified by the provided key.
	UpdateTwin(context.Context, string, Twin, Definition) error

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

type mqtt struct {
	client paho.Client
	topic  string
}

type twinsService struct {
	natsClient *nats.Conn
	mqtt       mqtt
	users      mainflux.UsersServiceClient
	twins      TwinRepository
	idp        IdentityProvider
}

var _ Service = (*twinsService)(nil)

// New instantiates the twins service implementation.
func New(nc *nats.Conn, mc paho.Client, topic string, users mainflux.UsersServiceClient, twins TwinRepository, idp IdentityProvider) Service {
	return &twinsService{
		natsClient: nc,
		mqtt: mqtt{
			client: mc,
			topic:  topic,
		},
		users: users,
		twins: twins,
		idp:   idp,
	}
}

func (ts *twinsService) publish(id *string, err error, succOp, failOp string, payload *[]byte) {
	if err != nil {
		ts.mqtt.publish(*id, succOp, payload)
	} else {
		ts.mqtt.publish(*id, failOp, payload)
	}
}

func (ts *twinsService) AddTwin(ctx context.Context, token string, twin Twin, def Definition) (tw Twin, err error) {
	var b []byte
	var id *string
	defer ts.publish(id, err, "create/success", "create/failure", &b)

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

	def.Created = time.Now()
	def.Revision = 0
	twin.Definitions = append(twin.Definitions, def)

	_, err = ts.twins.Save(ctx, twin)
	if err != nil {
		return Twin{}, err
	}

	twin.Revision = 0
	*id = twin.ID
	b, err = json.Marshal(twin)

	return twin, nil
}

func isZeroOfUnderlyingType(x interface{}) bool {
	return reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface())
}

func (ts *twinsService) UpdateTwin(ctx context.Context, token string, twin Twin, def Definition) error {
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
		def.Revision = tw.Definitions[len(tw.Definitions)-1].Revision + 1
		tw.Definitions = append(tw.Definitions, def)
	}

	if !isZeroOfUnderlyingType(twin.States) {
		tw.States = twin.States
	}

	if !isZeroOfUnderlyingType(twin.Metadata) {
		tw.Metadata = twin.Metadata
	}

	if err := ts.twins.Update(ctx, tw); err != nil {
		return err
	}

	b, err := json.Marshal(tw)
	if ts.mqtt.publish(twin.ID, "update/success", &b); err != nil {
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
	if ts.mqtt.publish(twin.ID, "get/success", &b); err != nil {
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

	if ts.mqtt.publish(id, "remove/success", nil); err != nil {
		return err
	}

	return nil
}

func (mqtt *mqtt) publish(id, op string, payload *[]byte) error {
	topic := fmt.Sprintf("channels/%s/messages/%s/%s", mqtt.topic, id, op)

	token := mqtt.client.Publish(topic, 0, false, &payload)
	token.Wait()

	return token.Error()
}
