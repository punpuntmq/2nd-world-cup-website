package scheduler

import (
	"context"
	"log"
	"time"

	"worldcup-realtime/internal/service"
	"worldcup-realtime/internal/sse"
)

type Refresher interface {
	Refresh(context.Context) (bool, error)
	State() service.ViewState
}

// minRefreshTimeout là sàn cứng để tránh config lỗi (buffer > interval)
// làm timeout tính ra âm hoặc gần 0.
const minRefreshTimeout = 1 * time.Second

type Scheduler struct {
	service        Refresher
	hub            *sse.Hub
	liveInterval   time.Duration
	idleInterval   time.Duration
	refreshTimeout time.Duration
	timeoutBuffer  time.Duration
}

func New(svc Refresher, hub *sse.Hub, liveInterval, idleInterval, refreshTimeout, timeoutBuffer time.Duration) *Scheduler {
	if liveInterval <= 0 {
		liveInterval = 10 * time.Second
	}
	if idleInterval <= 0 {
		idleInterval = 120 * time.Second
	}
	if refreshTimeout <= 0 {
		refreshTimeout = 30 * time.Second
	}
	if timeoutBuffer < 0 {
		timeoutBuffer = 500 * time.Millisecond
	}
	return &Scheduler{
		service:        svc,
		hub:            hub,
		liveInterval:   liveInterval,
		idleInterval:   idleInterval,
		refreshTimeout: refreshTimeout,
		timeoutBuffer:  timeoutBuffer,
	}
}

func (s *Scheduler) Run(ctx context.Context) {
	interval := s.idleInterval // start conservatively
	timer := time.NewTimer(interval)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			changed, err := s.refreshOnce(ctx, interval)
			if err != nil {
				log.Printf("scheduled refresh skipped or failed: %v", err)
			}
			if changed {
				s.hub.BroadcastState(s.service.State())
			}

			// Pick next interval based on live match count
			interval = s.nextInterval()
			timer.Reset(interval)
		}
	}
}

// refreshOnce runs exactly one refresh, bounded by a timeout that always
// leaves at least timeoutBuffer of headroom before the next scheduled tick.
// This guarantees a single hanging/slow upstream call can never block the
// scheduler forever (previously: refresh used the bare app ctx with no
// deadline, so one stuck request froze all future refreshes) and can never
// eat meaningfully into the following cycle either.
func (s *Scheduler) refreshOnce(ctx context.Context, interval time.Duration) (bool, error) {
	timeout := s.refreshTimeout
	if headroom := interval - s.timeoutBuffer; headroom < timeout {
		timeout = headroom
	}
	if timeout < minRefreshTimeout {
		timeout = minRefreshTimeout
	}

	fetchCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return s.service.Refresh(fetchCtx)
}

// nextInterval returns liveInterval when matches are live, idleInterval otherwise.
func (s *Scheduler) nextInterval() time.Duration {
	state := s.service.State()
	if state.Meta.LiveCount > 0 {
		return s.liveInterval
	}
	return s.idleInterval
}
