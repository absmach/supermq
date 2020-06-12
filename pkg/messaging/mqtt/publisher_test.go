// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mqtt_test

import (
	"fmt"
	"testing"

	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/stretchr/testify/require"
)

func TestPublish(t *testing.T) {
	err := publisher.Publish("test", messaging.Message{})
	require.Nil(t, err, fmt.Sprintf("got unexpected error: %s", err))
}
