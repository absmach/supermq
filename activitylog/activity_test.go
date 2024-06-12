// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package activitylog_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/absmach/magistrala/activitylog"
	"github.com/absmach/magistrala/internal/apiutil"
	"github.com/stretchr/testify/assert"
)

func TestActivitiesPage_MarshalJSON(t *testing.T) {
	occurredAt := time.Now()

	cases := []struct {
		desc string
		page activitylog.ActivitiesPage
		res  string
	}{
		{
			desc: "empty page",
			page: activitylog.ActivitiesPage{
				Activities: []activitylog.Activity(nil),
			},
			res: `{"total":0,"offset":0,"limit":0,"activities":[]}`,
		},
		{
			desc: "page with activities",
			page: activitylog.ActivitiesPage{
				Total:  1,
				Offset: 0,
				Limit:  0,
				Activities: []activitylog.Activity{
					{
						Operation:  "123",
						OccurredAt: occurredAt,
						Attributes: map[string]interface{}{"123": "123"},
						Metadata:   map[string]interface{}{"123": "123"},
					},
				},
			},
			res: fmt.Sprintf(`{"total":1,"offset":0,"limit":0,"activities":[{"operation":"123","occurred_at":"%s","attributes":{"123":"123"},"metadata":{"123":"123"}}]}`, occurredAt.Format(time.RFC3339Nano)),
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			data, err := tc.page.MarshalJSON()
			assert.NoError(t, err, "Unexpected error: %v", err)
			assert.Equal(t, tc.res, string(data))
		})
	}
}

func TestEntityType(t *testing.T) {
	cases := []struct {
		desc        string
		e           activitylog.EntityType
		str         string
		authString  string
		queryString string
	}{
		{
			desc:       "UserEntity",
			e:          activitylog.UserEntity,
			str:        "user",
			authString: "user",
		},
		{
			desc:       "ThingEntity",
			e:          activitylog.ThingEntity,
			str:        "thing",
			authString: "thing",
		},
		{
			desc:       "GroupEntity",
			e:          activitylog.GroupEntity,
			str:        "group",
			authString: "group",
		},
		{
			desc:       "ChannelEntity",
			e:          activitylog.ChannelEntity,
			str:        "channel",
			authString: "group",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			assert.Equal(t, tc.str, tc.e.String())
			assert.Equal(t, tc.authString, tc.e.AuthString())
			assert.NotEmpty(t, tc.e.Query())
		})
	}
}

func TestToEntityType(t *testing.T) {
	cases := []struct {
		desc        string
		entityType  string
		expected    activitylog.EntityType
		expectedErr error
	}{
		{
			desc:       "UserEntity",
			entityType: "user",
			expected:   activitylog.UserEntity,
		},
		{
			desc:       "ThingEntity",
			entityType: "thing",
			expected:   activitylog.ThingEntity,
		},
		{
			desc:       "GroupEntity",
			entityType: "group",
			expected:   activitylog.GroupEntity,
		},
		{
			desc:       "ChannelEntity",
			entityType: "channel",
			expected:   activitylog.ChannelEntity,
		},
		{
			desc:        "Invalid entity type",
			entityType:  "invalid",
			expectedErr: apiutil.ErrInvalidEntityType,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			entityType, err := activitylog.ToEntityType(tc.entityType)
			assert.Equal(t, tc.expected, entityType)
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}
