package limiter

import (
	"context"
	"errors"
)

var ErrInvalidLimit = errors.New("limit must be positive")

type Limiter struct{ tokens chan struct{} }

func New(max int) (*Limiter, error) {
	if max < 1 {
		return nil, ErrInvalidLimit
	}
	return &Limiter{tokens: make(chan struct{}, max)}, nil
}

func (l *Limiter) Acquire(ctx context.Context) error {
	select {
	case l.tokens <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (l *Limiter) Release()      { <-l.tokens }
func (l *Limiter) Inflight() int { return len(l.tokens) }
