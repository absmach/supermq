// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package readers

import "errors"

const (
	// EqualKey represents the equal comparison operator key
	EqualKey = "eq"
	// LowerThanKey represents the lower-than comparison operator key
	LowerThanKey = "lt"
	// LowerEqualThanKey represents the lower-or-equal-than comparison operator key
	LowerEqualThanKey = "lte"
	// GreaterThanKey represents the greater-or-equal-than comparison operator key
	GreaterThanKey = "gt"
	// GreaterEqualThanKey represents the greater-or-equal-than comparison operator key
	GreaterEqualThanKey = "gte"
)

// ErrNotFound indicates that requested entity doesn't exist.
var ErrNotFound = errors.New("entity not found")

// MessageRepository specifies message reader API.
type MessageRepository interface {
	// ReadAll skips given number of messages for given channel and returns next
	// limited number of messages.
	ReadAll(chanID string, pm PageMetadata) (MessagesPage, error)
}

// Message represents any message format.
type Message interface{}

// MessagesPage contains page related metadata as well as list of messages that
// belong to this page.
type MessagesPage struct {
	PageMetadata
	Total    uint64
	Messages []Message
}

// PageMetadata represents the parameters used to create database queries
type PageMetadata struct {
	Offset      uint64  `json:"offset"`
	Limit       uint64  `json:"limit"`
	Subtopic    string  `json:"subtopic,omitempty"`
	Publisher   string  `json:"publisher,omitempty"`
	Protocol    string  `json:"protocol,omitempty"`
	Name        string  `json:"name,omitempty"`
	Value       float64 `json:"v,omitempty"`
	Comparison  string  `json:"comparison,omitempty"`
	BoolValue   bool    `json:"vb,omitempty"`
	StringValue string  `json:"vs,omitempty"`
	DataValue   string  `json:"vd,omitempty"`
	From        float64 `json:"from,omitempty"`
	To          float64 `json:"to,omitempty"`
	Format      string  `json:"format,omitempty"`
}

// ParseValueComparison convert comparison operator keys into mathematic anotation
func ParseValueComparison(query map[string]interface{}) string {
	comparison := "="
	val, ok := query["comparison"]
	if ok {
		switch val.(string) {
		case EqualKey:
			comparison = "="
		case LowerThanKey:
			comparison = "<"
		case LowerEqualThanKey:
			comparison = "<="
		case GreaterThanKey:
			comparison = ">"
		case GreaterEqualThanKey:
			comparison = ">="
		}
	}

	return comparison
}
