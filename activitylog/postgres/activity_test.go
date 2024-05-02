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
	"github.com/absmach/magistrala/internal/testsutil"
	"github.com/absmach/magistrala/pkg/errors"
	repoerr "github.com/absmach/magistrala/pkg/errors/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	operation = "user.create"
	payload   = map[string]interface{}{
		"temperature": rand.Float64(),
		"humidity":    float64(rand.Intn(1000)),
		"locations": []interface{}{
			strings.Repeat("a", 100),
			strings.Repeat("a", 100),
		},
		"status": "active",
		"nested": map[string]interface{}{
			"nested": map[string]interface{}{
				"nested": map[string]interface{}{
					"nested": map[string]interface{}{
						"key": "value",
					},
				},
			},
		},
	}

	entityID          = testsutil.GenerateUUID(&testing.T{})
	thingOperation    = "thing.create"
	thingAttributesV1 = map[string]interface{}{
		"id":         entityID,
		"status":     "enabled",
		"created_at": time.Now().Add(-time.Hour),
		"name":       "thing",
		"tags":       []interface{}{"tag1", "tag2"},
		"domain":     testsutil.GenerateUUID(&testing.T{}),
		"metadata":   payload,
		"identity":   testsutil.GenerateUUID(&testing.T{}),
	}
	thingAttributesV2 = map[string]interface{}{
		"thing_id": entityID,
		"metadata": payload,
	}
	userAttributesV1 = map[string]interface{}{
		"id":         entityID,
		"status":     "enabled",
		"created_at": time.Now().Add(-time.Hour),
		"name":       "user",
		"tags":       []interface{}{"tag1", "tag2"},
		"domain":     testsutil.GenerateUUID(&testing.T{}),
		"metadata":   payload,
		"identity":   testsutil.GenerateUUID(&testing.T{}),
	}
	userAttributesV2 = map[string]interface{}{
		"user_id":  entityID,
		"metadata": payload,
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
				Attributes: payload,
				Metadata:   payload,
			},
			err: nil,
		},
		{
			desc: "with duplicate activity",
			activity: activitylog.Activity{
				Operation:  operation,
				OccurredAt: occurredAt,
				Attributes: payload,
				Metadata:   payload,
			},
			err: repoerr.ErrConflict,
		},
		{
			desc: "with massive activity metadata and attributes",
			activity: activitylog.Activity{
				Operation:  operation,
				OccurredAt: time.Now(),
				Attributes: map[string]interface{}{
					"attributes": map[string]interface{}{
						"attributes": map[string]interface{}{
							"attributes": map[string]interface{}{
								"attributes": map[string]interface{}{
									"attributes": map[string]interface{}{
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
				Metadata: map[string]interface{}{
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
			err: nil,
		},
		{
			desc: "with nil activity operation",
			activity: activitylog.Activity{
				OccurredAt: time.Now(),
				Attributes: payload,
				Metadata:   payload,
			},
			err: repoerr.ErrCreateEntity,
		},
		{
			desc: "with empty activity operation",
			activity: activitylog.Activity{
				Operation:  "",
				OccurredAt: time.Now(),
				Attributes: payload,
				Metadata:   payload,
			},
			err: repoerr.ErrCreateEntity,
		},
		{
			desc: "with nil activity occurred_at",
			activity: activitylog.Activity{
				Operation:  operation,
				Attributes: payload,
				Metadata:   payload,
			},
			err: repoerr.ErrCreateEntity,
		},
		{
			desc: "with empty activity occurred_at",
			activity: activitylog.Activity{
				Operation:  operation,
				OccurredAt: time.Time{},
				Attributes: payload,
				Metadata:   payload,
			},
			err: repoerr.ErrCreateEntity,
		},
		{
			desc: "with nil activity attributes",
			activity: activitylog.Activity{
				Operation:  operation + ".with.nil.attributes",
				OccurredAt: time.Now(),
				Metadata:   payload,
			},
			err: nil,
		},
		{
			desc: "with invalid activity attributes",
			activity: activitylog.Activity{
				Operation:  operation,
				OccurredAt: time.Now(),
				Attributes: map[string]interface{}{"invalid": make(chan struct{})},
				Metadata:   payload,
			},
			err: repoerr.ErrCreateEntity,
		},
		{
			desc: "with empty activity attributes",
			activity: activitylog.Activity{
				Operation:  operation + ".with.empty.attributes",
				OccurredAt: time.Now(),
				Attributes: map[string]interface{}{},
				Metadata:   payload,
			},
			err: nil,
		},
		{
			desc: "with nil activity metadata",
			activity: activitylog.Activity{
				Operation:  operation + ".with.nil.metadata",
				OccurredAt: time.Now(),
				Attributes: payload,
			},
			err: nil,
		},
		{
			desc: "with invalid activity metadata",
			activity: activitylog.Activity{
				Operation:  operation,
				OccurredAt: time.Now(),
				Metadata:   map[string]interface{}{"invalid": make(chan struct{})},
				Attributes: payload,
			},
			err: repoerr.ErrCreateEntity,
		},
		{
			desc: "with empty activity metadata",
			activity: activitylog.Activity{
				Operation:  operation + ".with.empty.metadata",
				OccurredAt: time.Now(),
				Metadata:   map[string]interface{}{},
				Attributes: payload,
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
			Attributes: userAttributesV1,
			Metadata:   payload,
		}
		if i%2 == 0 {
			activity.Operation = fmt.Sprintf("%s-%d", thingOperation, i)
			activity.Attributes = thingAttributesV1
		}
		if i%3 == 0 {
			activity.Attributes = userAttributesV2
		}
		if i%5 == 0 {
			activity.Attributes = thingAttributesV2
		}
		err := repo.Save(context.Background(), activity)
		require.Nil(t, err, fmt.Sprintf("create activity unexpected error: %s", err))
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
			desc: "with attributes",
			page: activitylog.Page{
				WithAttributes: true,
				Offset:         0,
				Limit:          10,
			},
			response: activitylog.ActivitiesPage{
				Total:      uint64(num),
				Offset:     0,
				Limit:      10,
				Activities: items[:10],
			},
		},
		{
			desc: "with metadata",
			page: activitylog.Page{
				WithMetadata: true,
				Offset:       0,
				Limit:        10,
			},
			response: activitylog.ActivitiesPage{
				Total:      uint64(num),
				Offset:     0,
				Limit:      10,
				Activities: items[:10],
			},
		},
		{
			desc: "with attributes and Metadata",
			page: activitylog.Page{
				WithAttributes: true,
				WithMetadata:   true,
				Offset:         0,
				Limit:          10,
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
			desc: "with user entity type",
			page: activitylog.Page{
				Offset:     0,
				Limit:      10,
				EntityID:   entityID,
				EntityType: activitylog.UserEntity,
			},
			response: activitylog.ActivitiesPage{
				Total:      uint64(len(extractEntities(items, activitylog.UserEntity, entityID))),
				Offset:     0,
				Limit:      10,
				Activities: extractEntities(items, activitylog.UserEntity, entityID)[:10],
			},
		},
		{
			desc: "with user entity type, attributes and metadata",
			page: activitylog.Page{
				Offset:         0,
				Limit:          10,
				EntityID:       entityID,
				EntityType:     activitylog.UserEntity,
				WithAttributes: true,
				WithMetadata:   true,
			},
			response: activitylog.ActivitiesPage{
				Total:      uint64(len(extractEntities(items, activitylog.UserEntity, entityID))),
				Offset:     0,
				Limit:      10,
				Activities: extractEntities(items, activitylog.UserEntity, entityID)[:10],
			},
		},
		{
			desc: "with thing entity type",
			page: activitylog.Page{
				Offset:     0,
				Limit:      10,
				EntityID:   entityID,
				EntityType: activitylog.ThingEntity,
			},
			response: activitylog.ActivitiesPage{
				Total:      uint64(len(extractEntities(items, activitylog.ThingEntity, entityID))),
				Offset:     0,
				Limit:      10,
				Activities: extractEntities(items, activitylog.ThingEntity, entityID)[:10],
			},
		},
		{
			desc: "with invalid entity id",
			page: activitylog.Page{
				Offset:     0,
				Limit:      10,
				EntityID:   testsutil.GenerateUUID(&testing.T{}),
				EntityType: activitylog.ChannelEntity,
			},
			response: activitylog.ActivitiesPage{
				Total:      0,
				Offset:     0,
				Limit:      10,
				Activities: []activitylog.Activity(nil),
			},
		},
		{
			desc: "with all filters",
			page: activitylog.Page{
				Offset:         0,
				Limit:          10,
				Operation:      items[0].Operation,
				From:           items[0].OccurredAt,
				To:             items[num-1].OccurredAt,
				WithAttributes: true,
				WithMetadata:   true,
				Direction:      "asc",
			},
			response: activitylog.ActivitiesPage{
				Total:      1,
				Offset:     0,
				Limit:      10,
				Activities: []activitylog.Activity{items[0]},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			page, err := repo.RetrieveAll(context.Background(), tc.page)
			assert.Equal(t, tc.response.Total, page.Total)
			assert.Equal(t, tc.response.Offset, page.Offset)
			assert.Equal(t, tc.response.Limit, page.Limit)
			for i := range page.Activities {
				assert.Equal(t, tc.response.Activities[i].Operation, page.Activities[i].Operation)
				assert.Equal(t, tc.response.Activities[i].OccurredAt, page.Activities[i].OccurredAt)
				if tc.page.WithAttributes {
					if _, ok := page.Activities[i].Attributes["created_at"].(string); ok {
						delete(page.Activities[i].Attributes, "created_at")
						delete(tc.response.Activities[i].Attributes, "created_at")
					}
					assert.Equal(t, page.Activities[i].Attributes, tc.response.Activities[i].Attributes)
				}
				if tc.page.WithMetadata {
					assert.Equal(t, page.Activities[i].Metadata, tc.response.Activities[i].Metadata)
				}
			}
			assert.Equal(t, tc.err, err)
		})
	}
}

func extractEntities(activities []activitylog.Activity, entityType activitylog.EntityType, entityID string) []activitylog.Activity {
	var entities []activitylog.Activity
	for _, activity := range activities {
		switch entityType {
		case activitylog.UserEntity:
			if strings.HasPrefix(activity.Operation, "user.") && activity.Attributes["id"] == entityID || activity.Attributes["user_id"] == entityID {
				entities = append(entities, activity)
			}
		case activitylog.GroupEntity:
			if strings.HasPrefix(activity.Operation, "group.") && activity.Attributes["id"] == entityID || activity.Attributes["group_id"] == entityID {
				entities = append(entities, activity)
			}
		case activitylog.ThingEntity:
			if strings.HasPrefix(activity.Operation, "thing.") && activity.Attributes["id"] == entityID || activity.Attributes["thing_id"] == entityID {
				entities = append(entities, activity)
			}
		case activitylog.ChannelEntity:
			if strings.HasPrefix(activity.Operation, "channel.") && activity.Attributes["id"] == entityID || activity.Attributes["group_id"] == entityID {
				entities = append(entities, activity)
			}
		}
	}

	return entities
}
