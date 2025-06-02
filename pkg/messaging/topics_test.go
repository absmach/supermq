// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package messaging_test

import (
	"testing"

	"github.com/absmach/supermq/pkg/messaging"
	"github.com/stretchr/testify/assert"
)

func TestParsePublishTopic(t *testing.T) {
	cases := []struct {
		desc      string
		topic     string
		domainID  string
		channelID string
		subtopic  string
		expectErr bool
	}{
		{
			desc:      "valid topic with subtopic",
			topic:     "/m/domain123/c/channel456/devices/temp",
			domainID:  "domain123",
			channelID: "channel456",
			subtopic:  "devices.temp",
		},
		{
			desc:      "valid topic with URL encoded subtopic",
			topic:     "/m/domain123/c/channel456/devices%2Ftemp%2Fdata",
			domainID:  "domain123",
			channelID: "channel456",
			subtopic:  "devices.temp.data",
		},
		{
			desc:      "valid topic with subtopic",
			topic:     "/m/domain/c/channel/extra/extra2",
			domainID:  "domain",
			channelID: "channel",
			subtopic:  "extra.extra2",
		},
		{
			desc:      "valid topic without subtopic",
			topic:     "/m/domain123/c/channel456",
			domainID:  "domain123",
			channelID: "channel456",
			subtopic:  "",
		},
		{
			desc:      "invalid topic format (missing parts)",
			topic:     "/m/domain123/c/",
			domainID:  "domain123",
			channelID: "",
			subtopic:  "",
			expectErr: true,
		},
		{
			desc:      "invalid topic format (missing domain)",
			topic:     "/m//c/channel123",
			domainID:  "",
			channelID: "",
			subtopic:  "",
			expectErr: true,
		},
		{
			desc:      "topic with wildcards + and #",
			topic:     "/m/domain123/c/channel456/devices/+/temp/#",
			domainID:  "domain123",
			channelID: "channel456",
			subtopic:  "",
			expectErr: true,
		},
		{
			desc:      "invalid domain name",
			topic:     "m/domain*123/c/channel456/devices/+/temp/#",
			domainID:  "",
			channelID: "channel456",
			subtopic:  "devices.*.temp.>",
			expectErr: true,
		},
		{
			desc:      "invalid subtopic",
			topic:     "/m/domain123/c/channel456/sub/a*b/topic",
			domainID:  "domain123",
			channelID: "channel456",
			subtopic:  "",
			expectErr: true,
		},
		{
			desc:      "invalid topic regex",
			topic:     "not-a-topic",
			domainID:  "",
			channelID: "",
			subtopic:  "",
			expectErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			domainID, channelID, subtopic, err := messaging.ParsePublishTopic(tc.topic)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.domainID, domainID)
				assert.Equal(t, tc.channelID, channelID)
				assert.Equal(t, tc.subtopic, subtopic)
			}
		})
	}
}

func TestParseSubscribeTopic(t *testing.T) {
	cases := []struct {
		desc      string
		topic     string
		domainID  string
		channelID string
		subtopic  string
		expectErr bool
	}{
		{
			desc:      "valid topic with subtopic",
			topic:     "/m/domain123/c/channel456/devices/temp",
			domainID:  "domain123",
			channelID: "channel456",
			subtopic:  "devices.temp",
		},
		{
			desc:      "topic with wildcards + and #",
			topic:     "/m/domain123/c/channel456/devices/+/temp/#",
			domainID:  "domain123",
			channelID: "channel456",
			subtopic:  "devices.*.temp.>",
		},
		{
			desc:      "valid topic without subtopic",
			topic:     "/m/domain123/c/channel456",
			domainID:  "domain123",
			channelID: "channel456",
			subtopic:  "",
		},
		{
			desc:      "invalid topic format (missing channel)",
			topic:     "/m/domain123/c/",
			domainID:  "domain123",
			channelID: "",
			subtopic:  "",
			expectErr: true,
		},
		{
			desc:      "invalid topic format (missing domain)",
			topic:     "/m//c/channel123",
			domainID:  "",
			channelID: "",
			subtopic:  "",
			expectErr: true,
		},
		{
			desc:      "invalid domain name m/domain*123/c/channel456/devices/+/temp/#",
			topic:     "m/domain*123/c/channel456/devices/+/temp/#",
			domainID:  "",
			channelID: "channel456",
			subtopic:  "devices.*.temp.>",
			expectErr: true,
		},
		{
			desc:      "invalid subtopic /m/domain123/c/channel456/sub/a*b/topic",
			topic:     "/m/domain123/c/channel456/sub/a*b/topic",
			domainID:  "domain123",
			channelID: "channel456",
			subtopic:  "",
			expectErr: true,
		},
		{
			desc:      "completely invalid topic",
			topic:     "invalid-topic",
			domainID:  "",
			channelID: "",
			subtopic:  "",
			expectErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			domainID, channelID, subtopic, err := messaging.ParseSubscribeTopic(tc.topic)
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.domainID, domainID)
				assert.Equal(t, tc.channelID, channelID)
				assert.Equal(t, tc.subtopic, subtopic)
			}
		})
	}
}

func TestEncodeTopic(t *testing.T) {
	cases := []struct {
		desc      string
		domainID  string
		channelID string
		subtopic  string
		expected  string
	}{
		{
			desc:      "with subtopic",
			domainID:  "domain1",
			channelID: "chan1",
			subtopic:  "dev.sensor.temp",
			expected:  "m.domain1.c.chan1.dev.sensor.temp",
		},
		{
			desc:      "without subtopic",
			domainID:  "domain1",
			channelID: "chan1",
			subtopic:  "",
			expected:  "m.domain1.c.chan1",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			actual := messaging.EncodeTopic(tc.domainID, tc.channelID, tc.subtopic)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestEncodeTopicSuffix(t *testing.T) {
	cases := []struct {
		desc      string
		domainID  string
		channelID string
		subtopic  string
		expected  string
	}{
		{
			desc:      "with subtopic",
			domainID:  "domain1",
			channelID: "chan1",
			subtopic:  "dev.sensor.temp",
			expected:  "domain1.c.chan1.dev.sensor.temp",
		},
		{
			desc:      "without subtopic",
			domainID:  "domain1",
			channelID: "chan1",
			subtopic:  "",
			expected:  "domain1.c.chan1",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			actual := messaging.EncodeTopicSuffix(tc.domainID, tc.channelID, tc.subtopic)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestMessage_EncodeTopicSuffix(t *testing.T) {
	cases := []struct {
		desc     string
		message  *messaging.Message
		expected string
	}{
		{
			desc: "with subtopic",
			message: &messaging.Message{
				Domain:   "domainX",
				Channel:  "chanX",
				Subtopic: "device.123.status",
			},
			expected: "domainX.c.chanX.device.123.status",
		},
		{
			desc: "without subtopic",
			message: &messaging.Message{
				Domain:  "domainY",
				Channel: "chanY",
			},
			expected: "domainY.c.chanY",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			actual := tc.message.EncodeTopicSuffix()
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestMessage_EncodeToMQTTTopic(t *testing.T) {
	cases := []struct {
		desc     string
		message  *messaging.Message
		expected string
	}{
		{
			desc: "with subtopic",
			message: &messaging.Message{
				Domain:   "domainA",
				Channel:  "chanA",
				Subtopic: "dev.1.temp",
			},
			expected: "m/domainA/c/chanA/dev/1/temp",
		},
		{
			desc: "without subtopic",
			message: &messaging.Message{
				Domain:  "domainB",
				Channel: "chanB",
			},
			expected: "m/domainB/c/chanB",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			actual := tc.message.EncodeToMQTTTopic()
			assert.Equal(t, tc.expected, actual)
		})
	}
}
