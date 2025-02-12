// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"context"
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

func (pc *patCache) Save(ctx context.Context, pat auth.PAT) error {
	for _, sc := range pat.Scope {
		key := GenerateKey(pat.ID, sc.OptionalDomainId, sc.EntityType, sc.Operation, sc.EntityId)
		if err := pc.client.Set(ctx, key, true, pc.duration).Err(); err != nil {
			return errors.Wrap(repoerr.ErrCreateEntity, err)
		}
	}

	return nil
}

func (pc *patCache) CheckScope(ctx context.Context, key string) (bool, error) {
	var authorized bool
	err := pc.client.Get(ctx, key).Scan(&authorized)
	if err != nil {
		return false, errors.Wrap(repoerr.ErrNotFound, err)
	}

	return authorized, nil
}

func (dc *patCache) Remove(ctx context.Context, patID string) error {
	if patID == "" {
		return errors.Wrap(repoerr.ErrRemoveEntity, errors.New("pat ID is empty"))
	}
	if err := dc.client.Del(ctx, patID).Err(); err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}

	return nil
}

func GenerateKey(patID, optionalDomainId string, entityType auth.EntityType, operation auth.Operation, entityID string) string {
	return patID + Separator + optionalDomainId + Separator + entityType.String() + Separator + operation.String() + Separator + entityID
}
