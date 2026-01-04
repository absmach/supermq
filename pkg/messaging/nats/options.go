// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package nats

import (
	"errors"
	"time"

	"github.com/absmach/supermq/pkg/messaging"
	"github.com/nats-io/nats.go/jetstream"
)

var (
	// ErrInvalidType is returned when the provided value is not of the expected type.
	ErrInvalidType = errors.New("invalid type")

	jsStreamConfig = jetstream.StreamConfig{
		Name:              "m",
		Description:       "SuperMQ stream for sending and receiving messages in between SuperMQ channels",
		Subjects:          []string{"m.>"},
		Retention:         jetstream.LimitsPolicy,
		MaxMsgsPerSubject: 1e6,
		MaxAge:            time.Hour * 24,
		MaxMsgSize:        1024 * 1024,
		Discard:           jetstream.DiscardOld,
		Storage:           jetstream.FileStorage,
	}
)

const (
	msgPrefix                       = "m"
	defaultMaxPendingMsgs           = 1000
	defaultMaxPendingBytes          = 5 * 1024 * 1024
	defaultEnableDroppedMsgTracking = true
)

type options struct {
	prefix             string
	jsStreamConfig     jetstream.StreamConfig
	slowConsumerConfig *SlowConsumerConfig
}

type SlowConsumerConfig struct {
	// MaxPendingMsgs maps to JetStream ConsumerConfig.MaxAckPending
	// Controls the maximum number of outstanding unacknowledged messages
	MaxPendingMsgs int

	// MaxPendingBytes maps to JetStream ConsumerConfig.MaxRequestMaxBytes
	// Controls the maximum bytes per batch request (closest JetStream equivalent)
	MaxPendingBytes int

	// EnableDroppedMsgTracking enables logging of message redeliveries
	// which can indicate slow consumer behavior in JetStream
	EnableDroppedMsgTracking bool
}

func defaultOptions() options {
	return options{
		prefix:         msgPrefix,
		jsStreamConfig: jsStreamConfig,
		slowConsumerConfig: &SlowConsumerConfig{
			MaxPendingMsgs:           defaultMaxPendingMsgs,
			MaxPendingBytes:          defaultMaxPendingBytes,
			EnableDroppedMsgTracking: defaultEnableDroppedMsgTracking,
		},
	}
}

// Prefix sets the prefix for the publisher or subscriber.
func Prefix(prefix string) messaging.Option {
	return func(val any) error {
		switch v := val.(type) {
		case *publisher:
			v.prefix = prefix
		case *pubsub:
			v.prefix = prefix
		default:
			return ErrInvalidType
		}

		return nil
	}
}

// JSStreamConfig sets the JetStream for the publisher or subscriber.
func JSStreamConfig(jsStreamConfig jetstream.StreamConfig) messaging.Option {
	return func(val any) error {
		switch v := val.(type) {
		case *publisher:
			v.jsStreamConfig = jsStreamConfig
		case *pubsub:
			v.jsStreamConfig = jsStreamConfig
		default:
			return ErrInvalidType
		}

		return nil
	}
}

// WithSlowConsumerConfig sets the slow consumer configuration.
func WithSlowConsumerConfig(config SlowConsumerConfig) messaging.Option {
	return func(val interface{}) error {
		switch v := val.(type) {
		case *publisher:
			v.slowConsumerConfig = &config
		case *pubsub:
			v.slowConsumerConfig = &config
		default:
			return ErrInvalidType
		}

		return nil
	}
}
