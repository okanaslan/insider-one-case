package idempotency

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client *redis.Client
	log    *slog.Logger
}

func NewRedisStore(client *redis.Client, log *slog.Logger) *RedisStore {
	return &RedisStore{client: client, log: log}
}

func (s *RedisStore) ReserveEvent(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if key == "" {
		return false, errors.New("idempotency key is required")
	}

	if s.client == nil {
		s.log.Debug("redis unavailable, skipping idempotency reserve", "key", key)
		return true, nil
	}

	reserved, err := s.client.SetNX(ctx, key, "1", ttl).Result()
	if err != nil {
		return false, err
	}

	return reserved, nil
}
