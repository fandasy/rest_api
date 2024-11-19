package rdb

import (
	"context"
	"errors"
	"log/slog"
	"restApi/internal/cache"
	"restApi/internal/config"
	"time"

	"restApi/pkg/e"

	"github.com/go-redis/redis/v8"
)

type Cache struct {
	rdb *redis.Client
	log *slog.Logger
	ttl time.Duration
}

func New(ctx context.Context, cfg config.Redis, log *slog.Logger) (*Cache, error) {
	const op = "cache.redis.New"

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, e.Wrap(op, err)
	}

	return &Cache{
		rdb: rdb,
		log: log,
		ttl: cfg.TTL,
	}, nil
}

func (c *Cache) Set(ctx context.Context, k string, v string) error {
	const op = "cache.redis.Set"

	err := c.rdb.Set(ctx, k, v, c.ttl).Err()
	if err != nil {
		return e.Wrap(op, err)
	}

	return nil
}

func (c *Cache) Get(ctx context.Context, k string) (string, error) {
	const op = "cache.redis.Get"

	val, err := c.rdb.Get(ctx, k).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", e.Wrap(op, cache.ErrKeyNotFound)
		}

		return "", e.Wrap(op, err)
	}

	return val, nil
}
