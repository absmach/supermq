package messaging_test

import (
	"testing"

	"github.com/absmach/supermq/pkg/messaging"
	"github.com/stretchr/testify/assert"
)

func TestParseTopic(t *testing.T) {
	cases := []struct {
		desc        string
		topic       string
		expDomainID string
		expChanID   string
		expSubtopic string
		expectErr   bool
	}{
		{
			desc:        "valid topic with subtopic",
			topic:       "/m/domain123/c/channel456/devices/temp",
			expDomainID: "domain123",
			expChanID:   "channel456",
			expSubtopic: "devices.temp",
		},
		{
			desc:        "valid topic with URL encoded subtopic",
			topic:       "/m/domain123/c/channel456/devices%2Ftemp%2Fdata",
			expDomainID: "domain123",
			expChanID:   "channel456",
			expSubtopic: "devices.temp.data",
		},
		{
			desc:        "valid topic with subtopic",
			topic:       "/m/domain/c/channel/extra/extra2",
			expDomainID: "domain",
			expChanID:   "channel",
			expSubtopic: "extra.extra2",
			expectErr:   false,
		},
		{
			desc:        "valid topic without subtopic",
			topic:       "/m/domain123/c/channel456",
			expDomainID: "domain123",
			expChanID:   "channel456",
			expSubtopic: "",
		},
		{
			desc:      "invalid topic format (missing parts)",
			topic:     "/m/domain123/c/",
			expectErr: true,
		},
		{
			desc:      "invalid topic format (missing domain)",
			topic:     "/m//c/channel123",
			expectErr: true,
		},
		{
			desc:      "invalid topic regex",
			topic:     "not-a-topic",
			expectErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			domainID, chanID, subtopic, err := messaging.ParseTopic(tc.topic)

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expDomainID, domainID)
				assert.Equal(t, tc.expChanID, chanID)
				assert.Equal(t, tc.expSubtopic, subtopic)
			}
		})
	}
}
