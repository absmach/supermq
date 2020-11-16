// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package json

// Message represents a JSON messages.
type Message struct {
	Channel   string                 `json:"channel,omitempty" db:"channel" bson:"channel"`
	Created   int64                  `json:"created,omitempty" db:"created" bson:"created"`
	Subtopic  string                 `json:"subtopic,omitempty" db:"subtopic" bson:"subtopic,omitempty"`
	Publisher string                 `json:"publisher,omitempty" db:"publisher" bson:"publisher"`
	Protocol  string                 `json:"protocol,omitempty" db:"protocol" bson:"protocol"`
	Payload   map[string]interface{} `json:"payload,omitempty" db:"payload" bson:"payload,omitempty"`
}

// Messages represents a list of JSON messages.
type Messages struct {
	Messages []Message
	Format   string
}
