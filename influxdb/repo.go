package influxdb

import "github.com/mainflux/mainflux"

// Writer represents writer implementation to the InfluxDB.
type Writer interface {
	mainflux.MessageRepository

	// Close method closes InfluxDB client.
	Close() error
}
