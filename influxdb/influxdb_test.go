package influxdb_test

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	client "github.com/influxdata/influxdb/client/v2"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/influxdb"
	"github.com/mainflux/mainflux/logger"
	"github.com/stretchr/testify/assert"
	dockertest "gopkg.in/ory-am/dockertest.v3"
)

var (
	port      string
	testLog   = logger.New(os.Stdout)
	testDB    = "test"
	cl        client.Client
	clientCfg = client.HTTPConfig{
		Username: "test",
		Password: "test",
	}
)

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	cfg := []string{
		"INFLUXDB_USER=test",
		"INFLUXDB_USER_PASSWORD=test",
		"INFLUXDB_DB=test",
	}
	container, err := pool.Run("influxdb", "1.5.2-alpine", cfg)
	if err != nil {
		testLog.Error(fmt.Sprintf("Could not start container: %s", err))
	}

	port = container.GetPort("8086/tcp")
	clientCfg.Addr = fmt.Sprintf("http://localhost:%s", port)

	if err := pool.Retry(func() error {
		cl, err = client.NewHTTPClient(clientCfg)
		_, _, err = cl.Ping(5 * time.Millisecond)
		return err
	}); err != nil {
		testLog.Error(fmt.Sprintf("Could not connect to docker: %s", err))
	}

	code := m.Run()

	if err := pool.Purge(container); err != nil {
		testLog.Error(fmt.Sprintf("Could not purge container: %s", err))
	}

	os.Exit(code)
}

// queryDB convenience function to query the database
func queryDB(cmd string) (res []client.Result, err error) {
	q := client.Query{
		Command:  cmd,
		Database: testDB,
	}
	if response, err := cl.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	} else {
		return res, err
	}
	return res, nil
}

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

	q := fmt.Sprintf("SELECT count(%s) FROM %s", "Value", "test..messages")

	repo, err := influxdb.New(clientCfg, testDB, "messages")
	assert.Nil(t, err, fmt.Sprintf("InfluxDB repo creation expected to succeed."))
	err = repo.Save(msg)
	assert.Nil(t, err, fmt.Sprintf("Save operation expected to succeed."))

	res, err := queryDB(q)
	assert.Nil(t, err)
	count := res[0].Series[0].Values[0][1]
	assert.Equal(t, json.Number("1"), count, fmt.Sprintf("Expected to have 1 value, found %s instead.\n", count))
}
