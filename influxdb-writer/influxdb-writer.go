package influxdb

import (
	"time"

	client "github.com/influxdata/influxdb/client/v2"
	"github.com/mainflux/mainflux"
)

const precision string = "ns"

type eventFlow struct {
	database  string
	pointName string
	client    client.Client
}

var _ Writer = (*eventFlow)(nil)

type fields map[string]interface{}
type tags map[string]string

func convertMsg(msg mainflux.Message) (tags, fields) {
	tags := map[string]string{
		"Channel":   msg.GetChannel(),
		"Publisher": msg.GetPublisher(),
		"Protocol":  msg.GetProtocol(),
		"Name":      msg.GetName(),
		"Unit":      msg.GetUnit(),
		"Link":      msg.GetLink(),
	}

	fields := map[string]interface{}{
		"Time":        msg.GetTime(),
		"UpdateTime":  msg.GetUpdateTime(),
		"StringValue": msg.GetStringValue(),
		"DataValue":   msg.GetDataValue(),
	}
	return tags, fields
}

// New returns new InfluxDB writer.
func New(cfg client.HTTPConfig, database, pointName string) (Writer, error) {
	c, err := client.NewHTTPClient(cfg)

	if err != nil {
		return nil, err
	}

	return &eventFlow{database, pointName, c}, nil
}

func (ef *eventFlow) Save(msg mainflux.Message) error {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database: ef.database,
	})

	if err != nil {
		return err
	}

	tags, fields := convertMsg(msg)
	pt, err := client.NewPoint(ef.pointName, tags, fields, time.Now())

	if err != nil {
		return err
	}

	bp.AddPoint(pt)

	if err := ef.client.Write(bp); err != nil {
		return err
	}
	return nil
}

func (ef *eventFlow) Close() error { return ef.client.Close() }
