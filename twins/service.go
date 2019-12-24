// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package twins

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/twins/mqtt"
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
	// AddTwin adds new twin related to user identified by the provided key.
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

var mqttOp = map[string]string{
	"createSucc": "create/success",
	"createFail": "create/failure",
	"updateSucc": "update/success",
	"updateFail": "update/failure",
	"getSucc":    "get/success",
	"getFail":    "get/failure",
	"removeSucc": "remove/success",
	"removeFail": "remove/failure",
}

type twinsService struct {
	natsClient *nats.Conn
	mqttClient mqtt.Mqtt
	auth       mainflux.AuthNServiceClient
	twins      TwinRepository
	states     StateRepository
	idp        IdentityProvider
}

var _ Service = (*twinsService)(nil)

// New instantiates the twins service implementation.
func New(nc *nats.Conn, mc mqtt.Mqtt, auth mainflux.AuthNServiceClient, twins TwinRepository, sr StateRepository, idp IdentityProvider) Service {
	return &twinsService{
		natsClient: nc,
		mqttClient: mc,
		auth:       auth,
		twins:      twins,
		states:     sr,
		idp:        idp,
	}
}

func (ts *twinsService) AddTwin(ctx context.Context, token string, twin Twin, def Definition) (tw Twin, err error) {
	var id string
	var b []byte
	defer ts.mqttClient.Publish(&id, &err, mqttOp["createSucc"], mqttOp["createFail"], &b)

	res, err := ts.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Twin{}, ErrUnauthorizedAccess
	}

	twin.ID, err = ts.idp.ID()
	if err != nil {
		return Twin{}, err
	}

	twin.Owner = res.GetValue()

	twin.Created = time.Now()
	twin.Updated = time.Now()

	fmt.Println(def)

	if len(def.Attributes) == 0 {
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

func (ts *twinsService) UpdateTwin(ctx context.Context, token string, twin Twin, def Definition) (err error) {
	var b []byte
	var id string
	defer ts.mqttClient.Publish(&id, &err, mqttOp["updateSucc"], mqttOp["updateFail"], &b)

	_, err = ts.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return ErrUnauthorizedAccess
	}

	tw, err := ts.twins.RetrieveByID(ctx, twin.ID)
	if err != nil {
		return err
	}
	tw.Updated = time.Now()
	tw.Revision++

	if twin.Name != "" {
		tw.Name = twin.Name
	}

	if twin.ThingID != "" {
		tw.ThingID = twin.ThingID
	}

	if len(def.Attributes) > 0 {
		def.Created = time.Now()
		def.ID = tw.Definitions[len(tw.Definitions)-1].ID + 1
		tw.Definitions = append(tw.Definitions, def)
	}

	if len(twin.Metadata) == 0 {
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
	var b []byte
	defer ts.mqttClient.Publish(&id, &err, mqttOp["getSucc"], mqttOp["getFail"], &b)

	_, err = ts.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Twin{}, ErrUnauthorizedAccess
	}

	twin, err := ts.twins.RetrieveByID(ctx, id)
	if err != nil {
		return Twin{}, err
	}

	b, err = json.Marshal(twin)

	return twin, nil
}

func (ts *twinsService) ViewTwinByThing(ctx context.Context, token, thingid string) (Twin, error) {
	_, err := ts.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return Twin{}, ErrUnauthorizedAccess
	}

	return ts.twins.RetrieveByThing(ctx, thingid)
}

func (ts *twinsService) RemoveTwin(ctx context.Context, token, id string) (err error) {
	var b []byte
	defer ts.mqttClient.Publish(&id, &err, mqttOp["removeSucc"], mqttOp["removeFail"], &b)

	_, err = ts.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return ErrUnauthorizedAccess
	}

	if err := ts.twins.Remove(ctx, id); err != nil {
		return err
	}

	return nil
}

func (ts *twinsService) ListTwins(ctx context.Context, token string, offset uint64, limit uint64, name string, metadata Metadata) (TwinsPage, error) {
	res, err := ts.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return TwinsPage{}, ErrUnauthorizedAccess
	}

	return ts.twins.RetrieveAll(ctx, res.GetValue(), offset, limit, name, metadata)
}

func (ts *twinsService) ListStates(ctx context.Context, token string, offset uint64, limit uint64, id string) (StatesPage, error) {
	_, err := ts.auth.Identify(ctx, &mainflux.Token{Value: token})
	if err != nil {
		return StatesPage{}, ErrUnauthorizedAccess
	}

	return ts.states.RetrieveAll(ctx, offset, limit, id)
}
