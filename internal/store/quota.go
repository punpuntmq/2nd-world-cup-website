package store

import "time"

type quotaWindow struct {
	limit       int
	window      time.Duration
	windowStart time.Time
	used        int
}

func newQuotaWindow(limit int, window time.Duration) *quotaWindow {
	return &quotaWindow{
		limit:  limit,
		window: window,
	}
}

func (q *quotaWindow) allow(now time.Time, cost int) bool {
	if cost <= 0 {
		return true
	}
	q.resetIfNeeded(now)
	if q.used+cost > q.limit {
		return false
	}
	q.used += cost
	return true
}

func (q *quotaWindow) remaining(now time.Time) int {
	q.resetIfNeeded(now)
	remaining := q.limit - q.used
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (q *quotaWindow) nextAllowedAt(now time.Time, cost int) time.Time {
	q.resetIfNeeded(now)
	if cost <= 0 || q.used+cost <= q.limit {
		return time.Time{}
	}
	return q.windowStart.Add(q.window)
}

func (q *quotaWindow) resetIfNeeded(now time.Time) {
	if q.windowStart.IsZero() || !now.Before(q.windowStart.Add(q.window)) {
		q.windowStart = now
		q.used = 0
	}
}
