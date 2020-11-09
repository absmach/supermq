// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package json

// Message represents a message emitted by the Mainflux adapters layer.
type Message struct {
	Channel   string                 `json:"channel,omitempty"`
	Subtopic  string                 `json:"subtopic,omitempty"`
	Publisher string                 `json:"publisher,omitempty"`
	Protocol  string                 `json:"protocol,omitempty"`
	Created   int64                  `json:"created,omitempty"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
}
