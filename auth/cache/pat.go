// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/absmach/supermq/auth"
	"github.com/absmach/supermq/pkg/errors"
	repoerr "github.com/absmach/supermq/pkg/errors/repository"
	"github.com/redis/go-redis/v9"
)

const Separator = "_"

type patCache struct {
	client   *redis.Client
	duration time.Duration
}

func NewPatsCache(client *redis.Client, duration time.Duration) auth.Cache {
	return &patCache{
		client:   client,
		duration: duration,
	}
}

func (pc *patCache) Save(ctx context.Context, scopes []auth.Scope) error {
	for _, sc := range scopes {
		key := generateKey(sc.PatID, sc.OptionalDomainID, sc.EntityType, sc.Operation, sc.EntityID)
		if err := pc.client.Set(ctx, key, true, pc.duration).Err(); err != nil {
			return errors.Wrap(repoerr.ErrCreateEntity, err)
		}
	}

	return nil
}

func (pc *patCache) CheckScope(patID, optionalDomainID string, entityType auth.EntityType, operation auth.Operation, entityID string) bool {
	ctx := context.Background()
	exactKey := fmt.Sprintf("pat:%s:%s:%s:%s:%s", patID, entityType, optionalDomainID, operation, entityID)
	wildcardKey := fmt.Sprintf("pat:%s:%s:%s:%s:*", patID, entityType, operation, operation)

	res, err := pc.client.Exists(ctx, exactKey, wildcardKey).Result()
	if err != nil {
		log.Println("Error checking PAT:", err)
		return false
	}

	return res > 0
}

func (dc *patCache) Remove(ctx context.Context, scopes []auth.Scope) error {
	for _, sc := range scopes {
		key := generateKey(sc.PatID, sc.OptionalDomainID, sc.EntityType, sc.Operation, sc.EntityID)
		if err := dc.client.Del(ctx, key).Err(); err != nil {
			return errors.Wrap(repoerr.ErrRemoveEntity, err)
		}
	}

	return nil
}

func (pc *patCache) RemoveAllScope(ctx context.Context, patID string) error {
	pattern := fmt.Sprintf("pat:%s", patID)

	iter := pc.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := pc.client.Del(ctx, iter.Val()).Err(); err != nil {
			return errors.Wrap(repoerr.ErrRemoveEntity, err)
		}
	}

	if err := iter.Err(); err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}

	return nil
}

func generateKey(patID, optionalDomainId string, entityType auth.EntityType, operation auth.Operation, entityID string) string {
	return fmt.Sprintf("pat:%s:%s:%s:%s:%s", patID, entityType, optionalDomainId, operation, entityID)
}
