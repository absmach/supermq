package influxdb_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/influxdata/influxdb/models"
	client "github.com/influxdata/influxdb/client/v2"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/influxdb"
	log "github.com/mainflux/mainflux/logger"
	"github.com/stretchr/testify/assert"
)

var (
	port      string
	testLog   = log.New(os.Stdout)
	testDB    = "test"
	cl        client.Client
	clientCfg = client.HTTPConfig{
		Username: "test",
		Password: "test",
	}
)

func TestSave(t *testing.T) {
	msg := mainflux.Message{
		Channel:     "ch",
		Publisher:   "pub",
		Protocol:    "http",
		Name:        "test name",
		Unit:        "km",
		Value:       24,
		StringValue: "24",
		BoolValue:   false,
		DataValue:   "dataValue",
		ValueSum:    24,
		Time:        13451312,
		UpdateTime:  5456565466,
		Link:        "link",
	}

	q := fmt.Sprintf("SELECT * FROM %s", "test..messages\n")

	repo, err := influxdb.New(clientCfg, testDB, "messages")
	assert.Nil(t, err, fmt.Sprintf("InfluxDB repo creation expected to succeed.\n"))
	err = repo.Save(msg)
	assert.Nil(t, err, fmt.Sprintf("Save operation expected to succeed.\n"))

	row, err := queryDB(q)
	assert.Nil(t, err, fmt.Sprintf("Querying InfluxDB to retrieve data count expected to succeed.\n"))
	count := len(row)
	assert.Equal(t, 1, count, fmt.Sprintf("Expected to have 1 value, found %d instead.\n", count))
}

// This is utility function to query the database.
func queryDB(cmd string) ([]models.Row, error) {
	q := client.Query{
		Command:  cmd,
		Database: testDB,
	}
	response, err := cl.Query(q)
	if err != nil {
		return nil, err
	}
	if response.Error() != nil {
		return nil, response.Error()
	}
	// There is only one query, so only one result and
	// all data are stored in the same series.
	return response.Results[0].Series, nil
}
