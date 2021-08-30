// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package influxdb

import (
	"math"
	"time"

	influxdata "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/mainflux/mainflux/consumers"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/transformers/senml"
)

var errSaveMessage = errors.New("failed to save message to influxdb database")

var _ consumers.Consumer = (*influxRepo)(nil)

type influxRepo struct {
	client           influxdata.Client
	bucket           string
	org              string
	writeAPI         api.WriteAPI
	mainfluxApiToken string
	thingsService    *ThingsService
}

// New returns new InfluxDB writer.
func New(client influxdata.Client, org, bucket, mainfluxApiToken, mainfluxUrl string) consumers.Consumer {
	thingsServiceConfig := &ThingsServiceConfig{
		Token: mainfluxApiToken,
		Url:   mainfluxUrl,
	}
	return &influxRepo{
		client:           client,
		bucket:           bucket,
		org:              org,
		writeAPI:         client.WriteAPI(org, bucket),
		mainfluxApiToken: mainfluxApiToken,
		thingsService:    NewThingsService(thingsServiceConfig),
	}
}

func (repo *influxRepo) Consume(message interface{}) error {
	switch m := message.(type) {
	// TODO: Bring back json support
	// case json.Messages:
	// 	pts, err := repo.jsonPoints(pts, m)
	default:
		repo.senmlPoints(m)
	}

	return nil
}

func (repo *influxRepo) senmlPoints(messages interface{}) error {
	msgs, ok := messages.([]senml.Message)
	if !ok {
		return errSaveMessage
	}

	for _, msg := range msgs {
		deviceName, measurement, err := senmlBasename(msg)

		if err != nil {
			deviceName = msg.Name
			measurement = msg.Name
		}
		meta, err := repo.thingsService.GetThingMetaById(msg.Publisher)

		tgs := senmlTags(msg, deviceName, meta)

		flds := senmlFields(msg)

		sec, dec := math.Modf(msg.Time)
		t := time.Unix(int64(sec), int64(dec*(1e9)))
		pt := influxdata.NewPoint(measurement, tgs, flds, t)
		repo.writeAPI.WritePoint(pt)
	}
	repo.writeAPI.Flush()
	return nil
}

// TODO: Bring back json support
// func (repo *influxRepo) jsonPoints(pts influxdata.BatchPoints, msgs json.Messages) (influxdata.BatchPoints, error) {
// 	for i, m := range msgs.Data {
// 		t := time.Unix(0, m.Created+int64(i))

// 		flat, err := json.Flatten(m.Payload)
// 		if err != nil {
// 			return nil, errors.Wrap(json.ErrTransform, err)
// 		}
// 		m.Payload = flat

// 		// Copy first-level fields so that the original Payload is unchanged.
// 		fields := make(map[string]interface{})
// 		for k, v := range m.Payload {
// 			fields[k] = v
// 		}
// 		// At least one known field need to exist so that COUNT can be performed.
// 		fields["protocol"] = m.Protocol
// 		pt, err := influxdata.NewPoint(msgs.Format, jsonTags(m), fields, t)
// 		if err != nil {
// 			return nil, errors.Wrap(errSaveMessage, err)
// 		}
// 		pts.AddPoint(pt)
// 	}

// return pts, nil
// }
