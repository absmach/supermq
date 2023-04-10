// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package consumers

// Consumer specifies message consuming API.
type AsyncConsumer interface {
	// Consume method is used to consume received messages.
	// An error channel is used to handle errors, so it supports the async approach.
	ConsumeAsync(messages interface{}, errs chan<- error)
}

type SyncConsumer interface {
	// ConsumeBlocking method is used to consume received messages synchronously
	// A non-nil error is returned to indicate operation failure
	ConsumeBlocking(messages interface{}) error
}
