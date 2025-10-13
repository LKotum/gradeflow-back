package utils

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisTokenStore struct{ rdb *redis.Client }

func NewRedisTokenStore(rdb *redis.Client) *RedisTokenStore { return &RedisTokenStore{rdb: rdb} }

func (s *RedisTokenStore) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return s.rdb.Set(ctx, key, value, ttl).Err()
}
func (s *RedisTokenStore) Get(ctx context.Context, key string) (string, error) {
	v, err := s.rdb.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return v, nil
}
func (s *RedisTokenStore) Del(ctx context.Context, key string) error {
	return s.rdb.Del(ctx, key).Err()
}
