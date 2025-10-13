package utils

import (
	"context"
	"testing"
	"time"
)

func TestMemoryTokenStore(t *testing.T) {
	s := NewMemoryTokenStore()
	ctx := context.Background()
	if _, err := s.Get(ctx, "k"); err == nil {
		t.Fatalf("expected not found")
	}
	if err := s.Set(ctx, "k", "v", 50*time.Millisecond); err != nil {
		t.Fatalf("set err: %v", err)
	}
	if got, err := s.Get(ctx, "k"); err != nil || got != "v" {
		t.Fatalf("get want v, got %q err %v", got, err)
	}
	time.Sleep(60 * time.Millisecond)
	if _, err := s.Get(ctx, "k"); err == nil {
		t.Fatalf("expected expired not found")
	}
	if err := s.Set(ctx, "k2", "v2", time.Second); err != nil {
		t.Fatal(err)
	}
	if err := s.Del(ctx, "k2"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get(ctx, "k2"); err == nil {
		t.Fatalf("expected deleted")
	}
}
