package limiter

import (
	"context"
	"testing"
	"time"
)

func TestLimiterBlocksAndHonorsCancellation(t *testing.T) {
	l, err := New(1)
	if err != nil {
		t.Fatal(err)
	}
	if err := l.Acquire(context.Background()); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := l.Acquire(ctx); err == nil {
		t.Fatal("expected cancellation while at capacity")
	}
	l.Release()
	if err := l.Acquire(context.Background()); err != nil {
		t.Fatalf("permit not released: %v", err)
	}
	l.Release()
}
