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
	"github.com/mainflux/mainflux/pkg/transformers/senml"
	"github.com/stretchr/testify/assert"
)

func TestTransformJSON(t *testing.T) {
	now := time.Now().Unix()
	tr := json.New()
	msg := messaging.Message{
		Channel:   "channel-1",
		Subtopic:  "subtopic-1",
		Publisher: "publisher-1",
		Protocol:  "protocol",
		Payload:   []byte(`{"key1": "val1", "key2": "val2", "key3": "val3"}`),
		Created:   now,
	}

	val1 := "val1"
	val2 := "val2"
	msgs := []senml.Message{
		{
			Channel:     "channel-1",
			Subtopic:    "subtopic-1",
			Publisher:   "publisher-1",
			Protocol:    "protocol",
			Name:        "key1",
			Time:        float64(now) / float64(1e9),
			StringValue: &val1,
		},
		{
			Channel:     "channel-1",
			Subtopic:    "subtopic-1",
			Publisher:   "publisher-1",
			Protocol:    "protocol",
			Name:        "key2",
			Time:        float64(now) / float64(1e9),
			StringValue: &val2,
		},
	}

	cases := []struct {
		desc string
		msg  messaging.Message
		msgs interface{}
		err  error
	}{
		{
			desc: "test normalize JSON",
			msg:  msg,
			msgs: msgs,
			err:  nil,
		},
	}

	for _, tc := range cases {
		msgs, err := tr.Transform(tc.msg)
		assert.Equal(t, tc.msgs, msgs, fmt.Sprintf("%s expected %v, got %v", tc.desc, tc.msgs, msgs))
		assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s expected %s, got %s", tc.desc, tc.err, err))
	}
}
