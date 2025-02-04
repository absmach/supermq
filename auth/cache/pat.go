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

func (pc *patCache) Save(ctx context.Context, patID string, pat auth.PAT) error {
	if err := pc.client.Set(ctx, patID, pat, pc.duration).Err(); err != nil {
		return errors.Wrap(repoerr.ErrCreateEntity, err)
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

func (dc *patCache) Remove(ctx context.Context, patID string) error {
	if patID == "" {
		return errors.Wrap(repoerr.ErrRemoveEntity, errors.New("pat ID is empty"))
	}
	if err := dc.client.Del(ctx, patID).Err(); err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}

	return nil
}
