package ratelimit

import (
	"context"
	"time"
)

type Limiter struct {
	tokens <-chan time.Time
}

func NewLimiter(rps int) *Limiter {
	if rps < 1 {
		rps = 1
	}
	interval := time.Second / time.Duration(rps)
	ticker := time.NewTicker(interval)
	return &Limiter{tokens: ticker.C}
}

func (l *Limiter) Take(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-l.tokens:
		return nil
	}
}
