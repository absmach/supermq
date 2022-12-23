// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package redis

import (
	"strconv"

	r "github.com/go-redis/redis/v8"
)

// Config of RedisDB
type Config struct {
	URL  string `env:"ES_URL"    envDefault:"localhost:6379"`
	Pass string `env:"ES_PASS"   envDefault:""`
	DB   string `env:"ES_DB"     envDefault:"0"`
}

// Connect create connection to RedisDB
func Connect(cfg Config) (*r.Client, error) {
	db, err := strconv.Atoi(cfg.DB)
	if err != nil {
		return nil, err
	}

	return r.NewClient(&r.Options{
		Addr:     cfg.URL,
		Password: cfg.Pass,
		DB:       db,
	}), nil
}
