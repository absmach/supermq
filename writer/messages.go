// Package writer provides message writer concept definitions.
package writer

import "github.com/mainflux/mainflux"

// Message represents a resolved (normalized) raw message.
type Message struct {
	Channel     string
	Publisher   string
	Protocol    string
	Name        string  `json:"n,omitempty"`
	Unit        string  `json:"u,omitempty"`
	Value       float64 `json:"v,omitempty"`
	StringValue string  `json:"vs,omitempty"`
	BoolValue   bool    `json:"vb,omitempty"`
	DataValue   string  `json:"vd,omitempty"`
	ValueSum    float64 `json:"s,omitempty"`
	Time        float64 `json:"t,omitempty"`
	UpdateTime  float64 `json:"ut,omitempty"`
	Link        string  `json:"l,omitempty"`
}

// MessageRepository specifies a message persistence API.
type MessageRepository interface {
	// Save persists the message. A non-nil error is returned to indicate
	// operation failure.
	Save(mainflux.Message) error
}
