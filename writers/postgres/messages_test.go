// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package postgres_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/mainflux/mainflux/transformers/senml"
	"github.com/mainflux/mainflux/writers/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gofrs/uuid"
)

var (
	msg         = senml.Message{}
	msgsNum     = 42
	valueFields = 6
)

func TestMessageSave(t *testing.T) {
	messageRepo := postgres.New(db)

	chid, err := uuid.NewV4()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	msg.Channel = chid.String()

	pubid, err := uuid.NewV4()
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
	msg.Publisher = pubid.String()

	now := time.Now().Unix()
	var msgs []senml.Message
	for i := 0; i < msgsNum; i++ {
		// Mix possible values as well as value sum.
		count := i % valueFields
		switch count {
		case 0:
			msg.Value = &senml.Message_FloatValue{FloatValue: 5}
		case 1:
			msg.Value = &senml.Message_BoolValue{BoolValue: false}
		case 2:
			msg.Value = &senml.Message_StringValue{StringValue: "value"}
		case 3:
			msg.Value = &senml.Message_DataValue{DataValue: "base64data"}
		case 4:
			msg.ValueSum = nil
		case 5:
			msg.ValueSum = &senml.SumValue{Value: 45}
		}
		msg.Time = float64(now + int64(i))
		msgs = append(msgs, msg)
	}

	err = messageRepo.Save(msgs...)
	assert.Nil(t, err, fmt.Sprintf("expected no error got %s\n", err))
}
