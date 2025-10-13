package utils

import (
	"context"
	"time"
)

type TokenStore interface {
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
}

// Memory token store for tests/dev
type MemoryTokenStore struct {
	m map[string]struct {
		v   string
		exp time.Time
	}
}

func NewMemoryTokenStore() *MemoryTokenStore {
	return &MemoryTokenStore{m: make(map[string]struct {
		v   string
		exp time.Time
	})}
}
func (s *MemoryTokenStore) Set(_ context.Context, k, v string, ttl time.Duration) error {
	s.m[k] = struct {
		v   string
		exp time.Time
	}{v, time.Now().Add(ttl)}
	return nil
}
func (s *MemoryTokenStore) Get(_ context.Context, k string) (string, error) {
	it, ok := s.m[k]
	if !ok {
		return "", ErrNotFound
	}
	if time.Now().After(it.exp) {
		delete(s.m, k)
		return "", ErrNotFound
	}
	return it.v, nil
}
func (s *MemoryTokenStore) Del(_ context.Context, k string) error { delete(s.m, k); return nil }

// Keys returns current non-expired keys; for tests only
func (s *MemoryTokenStore) Keys() []string {
	out := make([]string, 0, len(s.m))
	now := time.Now()
	for k, it := range s.m {
		if now.After(it.exp) { continue }
		out = append(out, k)
	}
	return out
}

var ErrNotFound = context.DeadlineExceeded // sentinel for not found
