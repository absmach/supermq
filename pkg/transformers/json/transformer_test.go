// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package json_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/transformers/json"
	"github.com/stretchr/testify/assert"
)

const (
	validPayload       = `{"key1": "val1", "key2": 123, "key3": "val3", "key4": {"key5": "val5"}}`
	listPayload        = `[{"key1": "val1", "key2": 123, "keylist3": "val3", "key4": {"key5": "val5"}}, {"key1": "val1", "key2": 123, "key3": "val3", "key4": {"key5": "val5"}}]`
	invalidPayload     = `{"key1": "val1", "key2": 123, "key3/1": "val3", "key4": {"key5": "val5"}}`
	invalidFlatPayload = `{"key1"}`
)

func TestTransformJSON(t *testing.T) {
	now := time.Now().Unix()
	trFlatten := json.New(true)
	trNoFlatten := json.New(false)

	msg := messaging.Message{
		Channel:   "channel-1",
		Subtopic:  "subtopic-1",
		Publisher: "publisher-1",
		Protocol:  "protocol",
		Payload:   []byte(validPayload),
		Created:   now,
	}
	invalid := msg
	invalid.Payload = []byte(invalidPayload)

	listMsg := msg
	listMsg.Payload = []byte(listPayload)

	flatJSONMsg := json.Message{
		Channel:   msg.Channel,
		Subtopic:  msg.Subtopic,
		Publisher: msg.Publisher,
		Protocol:  msg.Protocol,
		Created:   msg.Created,
		Payload: map[string]interface{}{
			"key1":      "val1",
			"key2":      float64(123),
			"key3":      "val3",
			"key4/key5": "val5",
		},
	}

	jsonMsgs := json.Messages{
		Data:   []json.Message{flatJSONMsg},
		Format: msg.Subtopic,
	}

	invalidFmt := msg
	invalidFmt.Subtopic = ""

	listJSON := json.Messages{
		Data: []json.Message{
			{
				Channel:   msg.Channel,
				Subtopic:  msg.Subtopic,
				Publisher: msg.Publisher,
				Protocol:  msg.Protocol,
				Created:   msg.Created,
				Payload: map[string]interface{}{
					"key1":      "val1",
					"key2":      float64(123),
					"keylist3":  "val3",
					"key4/key5": "val5",
				},
			},
			{
				Channel:   msg.Channel,
				Subtopic:  msg.Subtopic,
				Publisher: msg.Publisher,
				Protocol:  msg.Protocol,
				Created:   msg.Created,
				Payload: map[string]interface{}{
					"key1":      "val1",
					"key2":      float64(123),
					"key3":      "val3",
					"key4/key5": "val5",
				},
			},
		},
		Format: msg.Subtopic,
	}

	// Test cases with JSON flattening.
	flatCases := []struct {
		desc string
		msg  messaging.Message
		json interface{}
		err  error
	}{
		{
			desc: "test transform JSON",
			msg:  msg,
			json: jsonMsgs,
			err:  nil,
		},
		{
			desc: "test transform JSON with an invalid subtopic",
			msg:  invalidFmt,
			json: nil,
			err:  json.ErrTransform,
		},
		{
			desc: "test transform JSON array",
			msg:  listMsg,
			json: listJSON,
			err:  nil,
		},
		{
			desc: "test transform JSON with invalid payload",
			msg:  invalid,
			json: nil,
			err:  json.ErrTransform,
		},
	}

	noFlatJSONMsg := flatJSONMsg
	noFlatJSONMsg.Payload = map[string]interface{}{
		"key1": "val1",
		"key2": float64(123),
		"key3": "val3",
		"key4": map[string]interface{}{"key5": "val5"},
	}

	noFlatListJSON := json.Messages{
		Data: []json.Message{
			{
				Channel:   msg.Channel,
				Subtopic:  msg.Subtopic,
				Publisher: msg.Publisher,
				Protocol:  msg.Protocol,
				Created:   msg.Created,
				Payload: map[string]interface{}{
					"key1":     "val1",
					"key2":     float64(123),
					"keylist3": "val3",
					"key4":     map[string]interface{}{"key5": "val5"},
				},
			},
			{
				Channel:   msg.Channel,
				Subtopic:  msg.Subtopic,
				Publisher: msg.Publisher,
				Protocol:  msg.Protocol,
				Created:   msg.Created,
				Payload: map[string]interface{}{
					"key1": "val1",
					"key2": float64(123),
					"key3": "val3",
					"key4": map[string]interface{}{"key5": "val5"},
				},
			},
		},
		Format: msg.Subtopic,
	}

	jsonMsgs.Data = []json.Message{noFlatJSONMsg}
	noFlatInvalid := invalid
	noFlatInvalid.Payload = []byte(invalidFlatPayload)

	// Test cases without JSON flattening.
	noFlatCases := []struct {
		desc string
		msg  messaging.Message
		json interface{}
		err  error
	}{
		{
			desc: "test no-flattening transform JSON",
			msg:  msg,
			json: jsonMsgs,
			err:  nil,
		},
		{
			desc: "test no-flattening transform JSON with an invalid subtopic",
			msg:  invalidFmt,
			json: nil,
			err:  json.ErrTransform,
		},
		{
			desc: "test no-flattening transform JSON array",
			msg:  listMsg,
			json: noFlatListJSON,
			err:  nil,
		},
		{
			desc: "test no-flattening transform JSON with invalid payload",
			msg:  noFlatInvalid,
			json: nil,
			err:  json.ErrTransform,
		},
	}

	for _, tc := range flatCases {
		m, err := trFlatten.Transform(tc.msg)
		assert.Equal(t, tc.json, m, fmt.Sprintf("%s expected %v, got %v", tc.desc, tc.json, m))
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s expected %s, got %s", tc.desc, tc.err, err))
	}

	for _, tc := range noFlatCases {
		m, err := trNoFlatten.Transform(tc.msg)
		assert.Equal(t, tc.json, m, fmt.Sprintf("%s expected %v, got %v", tc.desc, tc.json, m))
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s expected %s, got %s", tc.desc, tc.err, err))
	}
}
