// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package opc

// Message represent an opc message
type Message struct {
	Namespace string  `json:"namespace"`
	ID        string  `json:"id"`
	Data      float64 `json:"data"`
}
