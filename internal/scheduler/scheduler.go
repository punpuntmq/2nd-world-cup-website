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

type Scheduler struct {
	service      Refresher
	hub          *sse.Hub
	liveInterval time.Duration
	idleInterval time.Duration
}

func New(svc Refresher, hub *sse.Hub, liveInterval, idleInterval time.Duration) *Scheduler {
	if liveInterval <= 0 {
		liveInterval = 10 * time.Second
	}
	if idleInterval <= 0 {
		idleInterval = 120 * time.Second
	}
	return &Scheduler{
		service:      svc,
		hub:          hub,
		liveInterval: liveInterval,
		idleInterval: idleInterval,
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
			changed, err := s.service.Refresh(ctx)
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

// nextInterval returns liveInterval when matches are live, idleInterval otherwise.
func (s *Scheduler) nextInterval() time.Duration {
	state := s.service.State()
	if state.Meta.LiveCount > 0 {
		return s.liveInterval
	}
	return s.idleInterval
}
