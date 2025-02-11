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
		key := generateKey(pat.ID, sc.OptionalDomainId, sc.EntityType.String(), sc.Operation.String(), sc.EntityId)
		if err := pc.client.Set(ctx, key, true, pc.duration).Err(); err != nil {
			return errors.Wrap(repoerr.ErrCreateEntity, err)
		}
	}

	return nil
}

func (pc *patCache) ID(ctx context.Context, patID string) (auth.PAT, error) {
	var pat auth.PAT
	err := pc.client.Get(ctx, patID).Scan(&pat)
	if err != nil {
		return auth.PAT{}, errors.Wrap(repoerr.ErrNotFound, err)
	}

	return pat, nil
}

func (pc *patCache) Check(ctx context.Context, key string) (bool, error) {
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

func generateKey(patID, optionalDomainId, entityType, operation, entityID string) string {
	return patID + optionalDomainId + entityType + operation + entityID
}
