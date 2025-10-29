package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// Store defines cache interactions used by services.
type Store interface {
	Get(ctx context.Context, key string, dest any) (bool, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	InvalidatePrefix(ctx context.Context, prefix string) error
}

// NoopStore is used when cache is disabled.
type NoopStore struct{}

// NewNoop returns a disabled cache implementation.
func NewNoop() *NoopStore {
	return &NoopStore{}
}

func (NoopStore) Get(context.Context, string, any) (bool, error) {
	return false, nil
}

func (NoopStore) Set(context.Context, string, any, time.Duration) error {
	return nil
}

func (NoopStore) Delete(context.Context, string) error {
	return nil
}

func (NoopStore) InvalidatePrefix(context.Context, string) error {
	return nil
}

// RedisStore implements Store backed by Redis.
type RedisStore struct {
	client     *redis.Client
	defaultTTL time.Duration
}

// NewRedis creates a redis-backed cache store.
func NewRedis(client *redis.Client, defaultTTL time.Duration) *RedisStore {
	if defaultTTL <= 0 {
		defaultTTL = time.Minute
	}
	return &RedisStore{client: client, defaultTTL: defaultTTL}
}

// Get reads cached value into destination. Returns false if key missing.
func (s *RedisStore) Get(ctx context.Context, key string, dest any) (bool, error) {
	if s == nil || s.client == nil {
		return false, errors.New("redis cache not configured")
	}
	res, err := s.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if dest == nil {
		return true, nil
	}
	if err := json.Unmarshal(res, dest); err != nil {
		return false, err
	}
	return true, nil
}

// Set stores value encoded as JSON with optional TTL.
func (s *RedisStore) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if s == nil || s.client == nil {
		return errors.New("redis cache not configured")
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if ttl <= 0 {
		ttl = s.defaultTTL
	}
	return s.client.Set(ctx, key, data, ttl).Err()
}

// Delete removes a single cache entry.
func (s *RedisStore) Delete(ctx context.Context, key string) error {
	if s == nil || s.client == nil {
		return errors.New("redis cache not configured")
	}
	return s.client.Del(ctx, key).Err()
}

// InvalidatePrefix removes entries that share common prefix.
func (s *RedisStore) InvalidatePrefix(ctx context.Context, prefix string) error {
	if s == nil || s.client == nil {
		return errors.New("redis cache not configured")
	}
	var cursor uint64
	for {
		keys, next, err := s.client.Scan(ctx, cursor, prefix+"*", 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := s.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return nil
}
