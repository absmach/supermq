// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package senml

import (
	"github.com/mainflux/mainflux/broker"
	"github.com/mainflux/mainflux/errors"
	"github.com/mainflux/mainflux/transformers"
	"github.com/mainflux/senml"
)

const (
	// JSON represents SenML in JSON format content type.
	JSON = "application/senml+json"
	// CBOR represents SenML in CBOR format content type.
	CBOR = "application/senml+cbor"
)

var (
	errDecode    = errors.New("failed to decode senml")
	errNormalize = errors.New("faled to normalize senml")
)

var formats = map[string]senml.Format{
	JSON: senml.JSON,
	CBOR: senml.CBOR,
}

type transformer struct {
	format senml.Format
}

// New returns transformer service implementation for SenML messages.
func New(contentType string) transformers.Transformer {
	return transformer{
		format: formats[contentType],
	}
}

func (n transformer) Transform(msg broker.Message) (interface{}, error) {
	raw, err := senml.Decode(msg.Payload, n.format)
	if err != nil {
		return nil, errors.Wrap(errDecode, err)
	}

	normalized, err := senml.Normalize(raw)
	if err != nil {
		return nil, errors.Wrap(errNormalize, err)
	}

	msgs := make([]Message, len(normalized.Records))
	for i, v := range normalized.Records {
		// Use reception timestamp if SenML messsage Time is missing
		time := v.Time
		if time == 0 {
			// Convert the timestamp into float64 with nanoseconds precision
			time = float64(msg.Created.GetSeconds()) + float64(msg.Created.GetNanos())/float64(1e9)
		}

		msgs[i] = Message{
			Channel:     msg.Channel,
			Subtopic:    msg.Subtopic,
			Publisher:   msg.Publisher,
			Protocol:    msg.Protocol,
			Name:        v.Name,
			Unit:        v.Unit,
			Time:        time,
			UpdateTime:  v.UpdateTime,
			Value:       v.Value,
			BoolValue:   v.BoolValue,
			DataValue:   v.DataValue,
			StringValue: v.StringValue,
			Sum:         v.Sum,
		}
	}

	return msgs, nil
}
