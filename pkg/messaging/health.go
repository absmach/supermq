// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package messaging

import "time"

const (
	HealthTopicPrefix = "hc"
	StatusOK          = "ok"
)

type HealthInfo struct {
	// Status contains adapter status.
	Status string `json:"status"`

	// Protocol contains the protocol used.
	Protocol string `json:"protocol"`

	// Timestamp of health check.
	Timestamp time.Time `json:"timestamp"`
}
