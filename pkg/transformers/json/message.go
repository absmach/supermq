// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package json

// Message represents a message emitted by the Mainflux adapters layer.
type Message struct {
	Channel   string      `json:"channel,omitempty" db:"channel" bson:"channel"`
	Created   int64       `json:"created,omitempty" db:"created" bson:"created"`
	Subtopic  string      `json:"subtopic,omitempty" db:"subtopic" bson:"subtopic,omitempty"`
	Publisher string      `json:"publisher,omitempty" db:"publisher" bson:"publisher"`
	Protocol  string      `json:"protocol,omitempty" db:"protocol" bson:"protocol"`
	Payload   interface{} `json:"payload,omitempty" db:"payload" bson:"payload,omitempty"`
}
