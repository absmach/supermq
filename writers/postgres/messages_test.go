//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package postgres_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/writers/postgres"
	"github.com/stretchr/testify/assert"
)

var (
	msg         = mainflux.Message{}
	msgsNum     = 42
	valueFields = 6
)

func TestMessageSave(t *testing.T) {
	messageRepo := postgres.New(db)

	now := time.Now().Unix()
	for i := 0; i < msgsNum; i++ {
		// Mix possible values as well as value sum.
		count := i % valueFields
		switch count {
		case 0:
			msg.Value = &mainflux.Message_FloatValue{FloatValue: 5}
		case 1:
			msg.Value = &mainflux.Message_BoolValue{BoolValue: false}
		case 2:
			msg.Value = &mainflux.Message_StringValue{StringValue: "value"}
		case 3:
			msg.Value = &mainflux.Message_DataValue{DataValue: "base64data"}
		case 4:
			msg.ValueSum = nil
		case 5:
			msg.ValueSum = &mainflux.SumValue{Value: 45}
		}
		msg.Time = float64(now + int64(i))

		if err := messageRepo.Save(msg); err != nil {
			assert.Equal(t, err, fmt.Sprintf("expected no error got %s\n", err))
		}
	}
}
