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
	service  Refresher
	hub      *sse.Hub
	interval time.Duration
}

func New(service Refresher, hub *sse.Hub, interval time.Duration) *Scheduler {
	if interval <= 0 {
		interval = 30 * time.Second
	}
	return &Scheduler{
		service:  service,
		hub:      hub,
		interval: interval,
	}
}

func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			changed, err := s.service.Refresh(ctx)
			if err != nil {
				log.Printf("scheduled refresh skipped or failed: %v", err)
				continue
			}
			if changed {
				s.hub.BroadcastState(s.service.State())
			}
		}
	}
}
