package ratelimit

import (
	"errors"
	"sync"
	"time"
)

var ErrRateLimited = errors.New("rate_limit_exceeded")

type bucket struct {
	windowStart time.Time
	count       int
}

type Limiter struct {
	enabled bool
	window  time.Duration
	max     int
	mu      sync.Mutex
	state   map[string]bucket
}

func NewLimiter(enabled bool, window time.Duration, max int) *Limiter {
	return &Limiter{
		enabled: enabled,
		window:  window,
		max:     max,
		state:   make(map[string]bucket),
	}
}

func (l *Limiter) Allow(key string) error {
	if l == nil || !l.enabled || key == "" {
		return nil
	}

	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	for existingKey, existingBucket := range l.state {
		if !existingBucket.windowStart.IsZero() && now.After(existingBucket.windowStart.Add(l.window)) {
			delete(l.state, existingKey)
		}
	}

	current, ok := l.state[key]
	if !ok || current.windowStart.IsZero() || now.After(current.windowStart.Add(l.window)) {
		l.state[key] = bucket{windowStart: now, count: 1}
		return nil
	}

	current.count++
	l.state[key] = current
	if current.count > l.max {
		return ErrRateLimited
	}
	return nil
}
