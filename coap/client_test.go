// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package coap_test

import (
	"testing"

	"github.com/mainflux/mainflux/coap"
)

const expectedCount = uint64(1)

var (
	msgChan = make(chan []byte)
	c       *coap.Client
	count   uint64
)

func TestHandle(t *testing.T) {

}

func TestCancel(t *testing.T) {

}
