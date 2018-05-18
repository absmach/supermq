package influxdb

import "github.com/mainflux/mainflux/writers"

// Writer represents writer implementation to the InfluxDB.
type Writer interface {
	writers.MessageRepository

	// Close method closes InfluxDB client.
	Close() error
}
