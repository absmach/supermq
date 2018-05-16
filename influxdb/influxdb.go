package influxdb

import (
	"strconv"
	"time"

	client "github.com/influxdata/influxdb/client/v2"
	"github.com/mainflux/mainflux"
)

var _ Writer = (*influxRepo)(nil)

type influxRepo struct {
	database  string
	pointName string
	client    client.Client
}

type fields map[string]interface{}
type tags map[string]string

func convertMsg(msg mainflux.Message) (tags, fields) {
	time := strconv.FormatFloat(msg.Time, 'f', -1, 64)
	update := strconv.FormatFloat(msg.UpdateTime, 'f', -1, 64)
	tags := map[string]string{
		"Channel":    msg.Channel,
		"Publisher":  msg.Publisher,
		"Protocol":   msg.Protocol,
		"Name":       msg.Name,
		"Unit":       msg.Unit,
		"Link":       msg.Link,
		"Time":       time,
		"UpdateTime": update,
	}

	fields := map[string]interface{}{
		"Value":       msg.Value,
		"ValueSum":    msg.ValueSum,
		"BoolValue":   msg.BoolValue,
		"StringValue": msg.StringValue,
		"DataValue":   msg.DataValue,
	}
	return tags, fields
}

// New returns new InfluxDB writer.
func New(cfg client.HTTPConfig, database, pointName string) (Writer, error) {
	c, err := client.NewHTTPClient(cfg)
	if err != nil {
		return nil, err
	}
	return &influxRepo{database, pointName, c}, nil
}

func (repo *influxRepo) Save(msg mainflux.Message) error {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database: repo.database,
	})
	if err != nil {
		return err
	}

	tags, fields := convertMsg(msg)
	pt, err := client.NewPoint(repo.pointName, tags, fields, time.Now())
	if err != nil {
		return err
	}

	bp.AddPoint(pt)
	if err := repo.client.Write(bp); err != nil {
		return err
	}
	return nil
}

func (repo *influxRepo) Close() error { return repo.client.Close() }
