package mongodb_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/writers/mongodb"

	log "github.com/mainflux/mainflux/logger"
	"github.com/mongodb/mongo-go-driver/mongo"
)

var (
	port       string
	addr       string
	testLog    = log.New(os.Stdout)
	testDB     = "test"
	collection = "mainflux"
	db         mongo.Database
)

func TestSave(t *testing.T) {
	msg := mainflux.Message{
		Channel:     45,
		Publisher:   2580,
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

	client, err := mongo.Connect(context.Background(), addr, nil)
	assert.Nil(t, err, fmt.Sprintf("Creating new MongoDB client expected to succeed.\n"))

	db := client.Database(testDB)
	repo, err := mongodb.New(db)
	assert.Nil(t, err, fmt.Sprintf("Creating new MongoDB repo expected to succeed.\n"))

	err = repo.Save(msg)
	assert.Nil(t, err, fmt.Sprintf("Save operation expected to succeed.\n"))

	count, _ := db.Collection(collection).Count(context.Background(), nil)
	assert.Equal(t, int64(1), count)
}
