// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package activitylog_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/absmach/magistrala/activitylog"
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
						ID:         "123",
						Operation:  "123",
						OccurredAt: occurredAt,
						Payload:    map[string]interface{}{"123": "123"},
					},
				},
			},
			res: fmt.Sprintf(`{"total":1,"offset":0,"limit":0,"activities":[{"id":"123","operation":"123","occurred_at":"%s","payload":{"123":"123"}}]}`, occurredAt.Format(time.RFC3339Nano)),
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
