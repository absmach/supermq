// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package influxdb

import (
	"math"
	"strings"
	"time"

	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/transformers/json"
	"github.com/mainflux/mainflux/pkg/transformers/senml"
	"github.com/mainflux/mainflux/writers"

	influxdata "github.com/influxdata/influxdb/client/v2"
)

const (
	senmlPoints = "messages"
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

func (repo *influxRepo) Save(message interface{}) error {
	pts, err := influxdata.NewBatchPoints(repo.cfg)
	if err != nil {
		return errors.Wrap(errSaveMessage, err)
	}
	switch m := message.(type) {
	case []json.Message:
		pts, err = repo.jsonPoints(pts, m)
	default:
		pts, err = repo.senmlPoints(pts, m)
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

func (repo *influxRepo) jsonPoints(pts influxdata.BatchPoints, msgs []json.Message) (influxdata.BatchPoints, error) {
	if len(msgs) == 0 {
		return pts, nil
	}
	name := strings.Split(msgs[0].Subtopic, ".")[0]
	for i, m := range msgs {
		t := time.Unix(0, m.Created+int64(i))
		pt, err := influxdata.NewPoint(name, jsonTags(m), m.Payload, t)
		if err != nil {
			return nil, errors.Wrap(errSaveMessage, err)
		}
		pts.AddPoint(pt)
	}

	return pts, nil
}

type message struct {
	json.Message
	payload []map[string]interface{}
}
