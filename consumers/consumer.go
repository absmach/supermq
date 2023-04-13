// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package consumers

// Consumer specifies message consuming API.
type AsyncConsumer interface {
	// Consume method is used to consume received messages.
	ConsumeAsync(messages interface{})

	// An error channel is used to handle errors, so it supports the async approach.
	Errors() <-chan error
}

type SyncConsumer interface {
	// ConsumeBlocking method is used to consume received messages synchronously.
	// A non-nil error is returned to indicate operation failure.
	ConsumeBlocking(messages interface{}) error
}
