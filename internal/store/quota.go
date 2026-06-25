package store

import (
	"sync"
	"time"
)

type QuotaManager interface {
	Allow(now time.Time, cost int) bool
	Remaining(now time.Time) int
	NextAllowedAt(now time.Time, cost int) time.Time
	Limit() int
}

type fixedWindowQuota struct {
	mu          sync.Mutex
	limit       int
	window      time.Duration
	windowStart time.Time
	used        int
}

func NewFixedWindowQuota(limit int, window time.Duration) QuotaManager {
	if limit <= 0 {
		limit = 1
	}
	if window <= 0 {
		window = time.Minute
	}
	return &fixedWindowQuota{
		limit:  limit,
		window: window,
	}
}

func (q *fixedWindowQuota) Allow(now time.Time, cost int) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.resetIfNeeded(now)

	if q.used+cost > q.limit {
		return false
	}
	q.used += cost
	return true
}

func (q *fixedWindowQuota) Remaining(now time.Time) int {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.resetIfNeeded(now)

	remaining := q.limit - q.used
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (q *fixedWindowQuota) NextAllowedAt(now time.Time, cost int) time.Time {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.resetIfNeeded(now)

	if q.used+cost <= q.limit {
		return time.Time{}
	}
	return q.windowStart.Add(q.window)
}

func (q *fixedWindowQuota) Limit() int {
	return q.limit
}

func (q *fixedWindowQuota) resetIfNeeded(now time.Time) {
	if q.windowExpired(now) {
		q.windowStart = now
		q.used = 0
	}
}

func (q *fixedWindowQuota) windowExpired(now time.Time) bool {
	return q.windowStart.IsZero() || !now.Before(q.windowStart.Add(q.window))
}
