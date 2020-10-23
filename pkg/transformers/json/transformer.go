// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package json

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/transformers"
)

const sep = "/"

// ErrInvalidKey represents an invalid JSON key format.
var ErrInvalidKey = errors.New("invalid object key")

type funcTransformer func(messaging.Message) (interface{}, error)

// New returns a new JSON transformer.
func New() transformers.Transformer {
	return funcTransformer(transformer)
}

func (fh funcTransformer) Transform(msg messaging.Message) (interface{}, error) {
	return fh(msg)
}

func transformer(msg messaging.Message) (interface{}, error) {
	ret := Message{
		Publisher: msg.Publisher,
		Created:   msg.Created,
		Protocol:  msg.Protocol,
		Channel:   msg.Channel,
		Subtopic:  msg.Subtopic,
	}
	if err := json.Unmarshal(msg.Payload, &ret.Payload); err != nil {
		return nil, err
	}
	var err error
	var pld = make(map[string]interface{})
	ret.Payload, err = flatten("", pld, ret.Payload)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func flatten(prefix string, m, m1 map[string]interface{}) (map[string]interface{}, error) {
	for k, v := range m1 {
		if k == "publisher" || k == "protocol" || k == "channel" || k == "subtopic" || strings.Contains(k, sep) {
			return nil, ErrInvalidKey
		}
		switch val := v.(type) {
		case map[string]interface{}:
			var err error
			m, err = flatten(prefix+k+sep, m, val)
			if err != nil {
				return nil, err
			}
		default:
			m[prefix+k] = v
		}
	}
	return m, nil
}
