// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package influxdb

import (
	"math"
	"time"

	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/transformers/json"
	"github.com/mainflux/mainflux/pkg/transformers/senml"
	"github.com/mainflux/mainflux/writers"

	influxdata "github.com/influxdata/influxdb/client/v2"
)

const (
	senmlPoints = "senml"
	jsonPoints  = "json"
)

var (
	errSaveMessage   = errors.New("failed to save message to influxdb database")
	errMessageFormat = errors.New("invalid message format")
)
var _ writers.MessageRepository = (*influxRepo)(nil)

type influxRepo struct {
	client influxdata.Client
	cfg    influxdata.BatchPointsConfig
}

// New returns new InfluxDB writer.
func New(client influxdata.Client, database string) writers.MessageRepository {
	return &influxRepo{
		client: client,
		cfg: influxdata.BatchPointsConfig{
			Database: database,
		},
	}
}

func (repo *influxRepo) Save(messages interface{}) error {
	pts, err := influxdata.NewBatchPoints(repo.cfg)
	if err != nil {
		return errors.Wrap(errSaveMessage, err)
	}
	switch m := messages.(type) {
	case json.Message:
		pts, err = repo.jsonPoints(pts, []json.Message{m})
	case []json.Message:
		pts, err = repo.jsonPoints(pts, m)
	default:
		pts, err = repo.senmlPoints(pts, messages)
	}
	if err != nil {
		return err
	}

	if err := repo.client.Write(pts); err != nil {
		return errors.Wrap(errSaveMessage, err)
	}
	return nil
}

func (repo *influxRepo) senmlPoints(pts influxdata.BatchPoints, messages interface{}) (influxdata.BatchPoints, error) {
	msgs, ok := messages.([]senml.Message)
	if !ok {
		return nil, errSaveMessage
	}

	for _, msg := range msgs {
		tgs, flds := senmlTags(msg), senmlFields(msg)

		sec, dec := math.Modf(msg.Time)
		t := time.Unix(int64(sec), int64(dec*(1e9)))

		pt, err := influxdata.NewPoint(senmlPoints, tgs, flds, t)
		if err != nil {
			return nil, errors.Wrap(errSaveMessage, err)
		}
		pts.AddPoint(pt)
	}

	return pts, nil
}

func (repo *influxRepo) jsonPoints(pts influxdata.BatchPoints, messages []json.Message) (influxdata.BatchPoints, error) {
	for i, m := range messages {
		t := time.Unix(0, m.Created+int64(i))
		pt, err := influxdata.NewPoint(jsonPoints, jsonTags(m), m.Payload, t)
		if err != nil {
			return nil, errors.Wrap(errSaveMessage, err)
		}
		pts.AddPoint(pt)
	}

	return pts, nil
}

type message struct {
	json.Message
	payload map[string]interface{}
}
