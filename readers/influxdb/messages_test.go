package influxdb_test

import (
	"os"

	influxdata "github.com/influxdata/influxdb/client/v2"
	log "github.com/mainflux/mainflux/logger"
)

var (
	port      string
	testLog   = log.New(os.Stdout)
	testDB    = "test"
	client    influxdata.Client
	clientCfg = influxdata.HTTPConfig{
		Username: "test",
		Password: "test",
	}
)
