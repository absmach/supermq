//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package mongodb_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	readers "github.com/mainflux/mainflux/readers/mongodb"
	writers "github.com/mainflux/mainflux/writers/mongodb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mainflux/mainflux"

	log "github.com/mainflux/mainflux/logger"
	"github.com/mongodb/mongo-go-driver/mongo"
)

const (
	testDB        = "test"
	collection    = "mainflux"
	chanID        = 1
	numOfMessages = 42
	valueFields   = 6
)

var (
	port string
	addr string
	msg  = mainflux.Message{
		Channel:   chanID,
		Publisher: 1,
		Protocol:  "mqtt",
	}
	testLog, _ = log.New(os.Stdout, log.Info.String())
)

func TestReadAll(t *testing.T) {
	client, err := mongo.Connect(context.Background(), addr, nil)
	require.Nil(t, err, fmt.Sprintf("Creating new MongoDB client expected to succeed: %s.\n", err))

	db := client.Database(testDB)

	writer := writers.New(db)

	messages := []mainflux.Message{}
	for i := 0; i < numOfMessages; i++ {
		// Mix possible values as well as value sum.
		count := i % valueFields
		switch count {
		case 0:
			msg.Values = &mainflux.Message_Value{5}
		case 1:
			msg.Values = &mainflux.Message_BoolValue{false}
		case 2:
			msg.Values = &mainflux.Message_StringValue{"value"}
		case 3:
			msg.Values = &mainflux.Message_DataValue{"base64data"}
		case 4:
			msg.ValueSum = nil
		case 5:
			msg.ValueSum = &mainflux.Sum{Value: 45}
		}

		err := writer.Save(msg)
		require.Nil(t, err, fmt.Sprintf("failed to store message to Cassandra: %s", err))
		messages = append(messages, msg)
	}

	reader := readers.New(db)

	cases := map[string]struct {
		chanID   uint64
		offset   uint64
		limit    uint64
		messages []mainflux.Message
	}{
		"read message page for existing channel": {
			chanID:   chanID,
			offset:   0,
			limit:    10,
			messages: messages[0:10],
		},
		"read message page for non-existent channel": {
			chanID:   2,
			offset:   0,
			limit:    10,
			messages: []mainflux.Message{},
		},
		"read message last page": {
			chanID:   chanID,
			offset:   40,
			limit:    10,
			messages: messages[40:42],
		},
	}

	for desc, tc := range cases {
		result := reader.ReadAll(tc.chanID, tc.offset, tc.limit)
		assert.ElementsMatch(t, tc.messages, result, fmt.Sprintf("%s: expected %v got %v", desc, tc.messages, result))
	}
}
