// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package postgres_test

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/absmach/magistrala/activitylog"
	"github.com/absmach/magistrala/activitylog/postgres"
	"github.com/absmach/magistrala/pkg/errors"
	repoerr "github.com/absmach/magistrala/pkg/errors/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	operation = "user.create"
	payload   = map[string]interface{}{
		"temperature": rand.Float64(),
		"humidity":    rand.Float64(),
		"sensor_id":   rand.Intn(1000),
		"locations": []string{
			strings.Repeat("a", 1024),
			strings.Repeat("a", 1024),
			strings.Repeat("a", 1024),
		},
		"status":    rand.Intn(1000),
		"timestamp": time.Now().UnixNano(),
	}
)

func TestActivitySave(t *testing.T) {
	t.Cleanup(func() {
		_, err := db.Exec("DELETE FROM activities")
		require.Nil(t, err, fmt.Sprintf("clean activities unexpected error: %s", err))
	})
	repo := postgres.NewRepository(database)

	occurredAt := time.Now()

	cases := []struct {
		desc     string
		activity activitylog.Activity
		err      error
	}{
		{
			desc: "new activity successfully",
			activity: activitylog.Activity{
				Operation:  operation,
				OccurredAt: occurredAt,
				Payload:    payload,
			},
			err: nil,
		},
		{
			desc: "with duplicate activity",
			activity: activitylog.Activity{
				Operation:  operation,
				OccurredAt: occurredAt,
				Payload:    payload,
			},
			err: repoerr.ErrConflict,
		},
		{
			desc: "with massive activity payload",
			activity: activitylog.Activity{
				Operation:  operation,
				OccurredAt: time.Now(),
				Payload: map[string]interface{}{
					"metadata": map[string]interface{}{
						"metadata": map[string]interface{}{
							"metadata": map[string]interface{}{
								"metadata": map[string]interface{}{
									"metadata": map[string]interface{}{
										"data": payload,
									},
									"data": payload,
								},
								"data": payload,
							},
							"data": payload,
						},
						"data": payload,
					},
					"data": payload,
				},
			},
			err: repoerr.ErrConflict,
		},
		{
			desc: "with nil activity operation",
			activity: activitylog.Activity{
				OccurredAt: time.Now(),
				Payload:    payload,
			},
			err: repoerr.ErrCreateEntity,
		},
		{
			desc: "with empty activity operation",
			activity: activitylog.Activity{
				Operation:  "",
				OccurredAt: time.Now(),
				Payload:    payload,
			},
			err: repoerr.ErrConflict,
		},
		{
			desc: "with nil activity occurred_at",
			activity: activitylog.Activity{
				Operation: operation,
				Payload:   payload,
			},
			err: repoerr.ErrCreateEntity,
		},
		{
			desc: "with empty activity occurred_at",
			activity: activitylog.Activity{
				Operation:  operation,
				OccurredAt: time.Time{},
				Payload:    payload,
			},
			err: repoerr.ErrConflict,
		},
		{
			desc: "with nil activity payload",
			activity: activitylog.Activity{
				Operation:  operation,
				OccurredAt: time.Now(),
			},
			err: repoerr.ErrConflict,
		},
		{
			desc: "with invalid activity payload",
			activity: activitylog.Activity{
				Operation:  operation,
				OccurredAt: time.Now(),
				Payload:    map[string]interface{}{"invalid": make(chan struct{})},
			},
			err: repoerr.ErrCreateEntity,
		},
		{
			desc: "with empty activity payload",
			activity: activitylog.Activity{
				Operation:  operation,
				OccurredAt: time.Now(),
				Payload:    map[string]interface{}{},
			},
			err: nil,
		},
		{
			desc:     "with empty activity",
			activity: activitylog.Activity{},
			err:      repoerr.ErrMalformedEntity,
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			switch err := repo.Save(context.Background(), tc.activity); {
			case err == nil:
				assert.Nil(t, err)
			default:
				assert.True(t, errors.Contains(err, tc.err), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
			}
		})
	}
}

func TestActivityRetrieveAll(t *testing.T) {
	t.Cleanup(func() {
		_, err := db.Exec("DELETE FROM activities")
		require.Nil(t, err, fmt.Sprintf("clean activities unexpected error: %s", err))
	})
	repo := postgres.NewRepository(database)

	num := 200

	var items []activitylog.Activity
	for i := 0; i < num; i++ {
		activity := activitylog.Activity{
			Operation:  fmt.Sprintf("%s-%d", operation, i),
			OccurredAt: time.Now().UTC().Truncate(time.Millisecond),
			Payload:    payload,
		}
		err := repo.Save(context.Background(), activity)
		require.Nil(t, err, fmt.Sprintf("create activity unexpected error: %s", err))
		activity.Payload = nil
		items = append(items, activity)
	}

	reversedItems := make([]activitylog.Activity, len(items))
	copy(reversedItems, items)
	sort.Slice(reversedItems, func(i, j int) bool {
		return reversedItems[i].OccurredAt.After(reversedItems[j].OccurredAt)
	})

	cases := []struct {
		desc     string
		page     activitylog.Page
		response activitylog.ActivitiesPage
		err      error
	}{
		{
			desc: "successfully",
			page: activitylog.Page{
				Offset: 0,
				Limit:  1,
			},
			response: activitylog.ActivitiesPage{
				Total:      uint64(num),
				Offset:     0,
				Limit:      1,
				Activities: items[:1],
			},
			err: nil,
		},
		{
			desc: "with offset and empty limit",
			page: activitylog.Page{
				Offset: 10,
			},
			response: activitylog.ActivitiesPage{
				Total:      uint64(num),
				Offset:     10,
				Limit:      0,
				Activities: []activitylog.Activity(nil),
			},
		},
		{
			desc: "with limit and empty offset",
			page: activitylog.Page{
				Limit: 50,
			},
			response: activitylog.ActivitiesPage{
				Total:      uint64(num),
				Offset:     0,
				Limit:      50,
				Activities: items[:50],
			},
		},
		{
			desc: "with offset and limit",
			page: activitylog.Page{
				Offset: 10,
				Limit:  50,
			},
			response: activitylog.ActivitiesPage{
				Total:      uint64(num),
				Offset:     10,
				Limit:      50,
				Activities: items[10:60],
			},
		},
		{
			desc: "with offset out of range",
			page: activitylog.Page{
				Offset: 1000,
				Limit:  50,
			},
			response: activitylog.ActivitiesPage{
				Total:      uint64(num),
				Offset:     1000,
				Limit:      50,
				Activities: []activitylog.Activity(nil),
			},
		},
		{
			desc: "with offset and limit out of range",
			page: activitylog.Page{
				Offset: 170,
				Limit:  50,
			},
			response: activitylog.ActivitiesPage{
				Total:      uint64(num),
				Offset:     170,
				Limit:      50,
				Activities: items[170:200],
			},
		},
		{
			desc: "with limit out of range",
			page: activitylog.Page{
				Offset: 0,
				Limit:  1000,
			},
			response: activitylog.ActivitiesPage{
				Total:      uint64(num),
				Offset:     0,
				Limit:      1000,
				Activities: items,
			},
		},
		{
			desc: "with empty page",
			page: activitylog.Page{},
			response: activitylog.ActivitiesPage{
				Total:      uint64(num),
				Offset:     0,
				Limit:      0,
				Activities: []activitylog.Activity(nil),
			},
		},
		{
			desc: "with operation",
			page: activitylog.Page{
				Operation: items[0].Operation,
				Offset:    0,
				Limit:     10,
			},
			response: activitylog.ActivitiesPage{
				Total:      1,
				Offset:     0,
				Limit:      10,
				Activities: []activitylog.Activity{items[0]},
			},
		},
		{
			desc: "with payload",
			page: activitylog.Page{
				WithPayload: true,
				Offset:      0,
				Limit:       10,
			},
			response: activitylog.ActivitiesPage{
				Total:      uint64(num),
				Offset:     0,
				Limit:      10,
				Activities: items[:10],
			},
		},
		{
			desc: "with from",
			page: activitylog.Page{
				From:   items[0].OccurredAt,
				Offset: 0,
				Limit:  10,
			},
			response: activitylog.ActivitiesPage{
				Total:      uint64(num),
				Offset:     0,
				Limit:      10,
				Activities: items[:10],
			},
		},
		{
			desc: "with to",
			page: activitylog.Page{
				To:     items[num-1].OccurredAt,
				Offset: 0,
				Limit:  10,
			},
			response: activitylog.ActivitiesPage{
				Total:      uint64(num),
				Offset:     0,
				Limit:      10,
				Activities: items[:10],
			},
		},
		{
			desc: "with from and to",
			page: activitylog.Page{
				From:   items[0].OccurredAt,
				To:     items[num-1].OccurredAt,
				Offset: 0,
				Limit:  10,
			},
			response: activitylog.ActivitiesPage{
				Total:      uint64(num),
				Offset:     0,
				Limit:      10,
				Activities: items[:10],
			},
		},
		{
			desc: "with asc direction",
			page: activitylog.Page{
				Direction: "asc",
				Offset:    0,
				Limit:     10,
			},
			response: activitylog.ActivitiesPage{
				Total:      uint64(num),
				Offset:     0,
				Limit:      10,
				Activities: items[:10],
			},
		},
		{
			desc: "with desc direction",
			page: activitylog.Page{
				Direction: "desc",
				Offset:    0,
				Limit:     10,
			},
			response: activitylog.ActivitiesPage{
				Total:      uint64(num),
				Offset:     0,
				Limit:      10,
				Activities: reversedItems[:10],
			},
		},
		{
			desc: "with all filters",
			page: activitylog.Page{
				Operation:   items[0].Operation,
				From:        items[0].OccurredAt,
				To:          items[num-1].OccurredAt,
				WithPayload: true,
				Direction:   "asc",
				Offset:      0,
				Limit:       10,
			},
			response: activitylog.ActivitiesPage{
				Total:      1,
				Offset:     0,
				Limit:      10,
				Activities: []activitylog.Activity{items[0]},
			},
		},
		{
			desc: "with invalid operation",
			page: activitylog.Page{
				Operation: strings.Repeat("a", 37),
				Offset:    0,
				Limit:     10,
			},
			response: activitylog.ActivitiesPage{
				Total:      0,
				Offset:     0,
				Limit:      10,
				Activities: []activitylog.Activity(nil),
			},
		},
		{
			desc: "with invalid from",
			page: activitylog.Page{
				From:   time.Now().UTC().Truncate(time.Millisecond),
				Offset: 0,
				Limit:  10,
			},
			response: activitylog.ActivitiesPage{
				Total:      0,
				Offset:     0,
				Limit:      10,
				Activities: []activitylog.Activity(nil),
			},
		},
		{
			desc: "with invalid to",
			page: activitylog.Page{
				To:     time.Now().UTC().Truncate(time.Millisecond).Add(-time.Hour),
				Offset: 0,
				Limit:  10,
			},
			response: activitylog.ActivitiesPage{
				Total:      0,
				Offset:     0,
				Limit:      10,
				Activities: []activitylog.Activity(nil),
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			page, err := repo.RetrieveAll(context.Background(), tc.page)
			assert.Equal(t, tc.response.Total, page.Total)
			assert.Equal(t, tc.response.Offset, page.Offset)
			assert.Equal(t, tc.response.Limit, page.Limit)
			if !tc.page.WithPayload {
				assert.ElementsMatch(t, page.Activities, tc.response.Activities)
			}
			assert.Equal(t, tc.err, err)
		})
	}
}
