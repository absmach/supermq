// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package senml

import (
	"github.com/cisco/senml"
	"github.com/mainflux/mainflux"
	mfxTransformer "github.com/mainflux/mainflux/transformer"
)

var formats = map[string]senml.Format{
	SenMLJSON: senml.JSON,
	SenMLCBOR: senml.CBOR,
}

type transformer struct{}

// New returns normalizer service implementation.
func New() mfxTransformer.Transformer {
	return transformer{}
}

func (n transformer) Transform(msg mainflux.Message) (interface{}, error) {
	format, ok := formats[msg.ContentType]
	if !ok {
		format = senml.JSON
	}

	raw, err := senml.Decode(msg.Payload, format)
	if err != nil {
		return nil, err
	}

	normalized := senml.Normalize(raw)

	msgs := make([]Message, len(normalized.Records))
	for k, v := range normalized.Records {
		m := Message{
			Channel:    msg.Channel,
			Subtopic:   msg.Subtopic,
			Publisher:  msg.Publisher,
			Protocol:   msg.Protocol,
			Name:       v.Name,
			Unit:       v.Unit,
			Time:       v.Time,
			UpdateTime: v.UpdateTime,
			Link:       v.Link,
		}

		switch {
		case v.Value != nil:
			m.Value = &Message_FloatValue{FloatValue: *v.Value}
		case v.BoolValue != nil:
			m.Value = &Message_BoolValue{BoolValue: *v.BoolValue}
		case v.DataValue != "":
			m.Value = &Message_DataValue{DataValue: v.DataValue}
		case v.StringValue != "":
			m.Value = &Message_StringValue{StringValue: v.StringValue}
		}

		if v.Sum != nil {
			m.ValueSum = &SumValue{Value: *v.Sum}
		}

		msgs[k] = m
	}

	return msgs, nil
}
