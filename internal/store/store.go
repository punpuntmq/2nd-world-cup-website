package store

import (
	"context"
	"sync"
	"time"

	"worldcup-realtime/internal/football"
)

type FootballClient interface {
	Fetch(context.Context) (football.RawState, error)
}

type MemoryStore struct {
	mu              sync.RWMutex
	client          FootballClient
	refreshInterval time.Duration
	raw             football.RawState
	view            ViewState
	lastError       string
}

func NewMemoryStore(client FootballClient, refreshInterval time.Duration) *MemoryStore {

	return &MemoryStore{
		client:          client,
		refreshInterval: refreshInterval,
	}
}

func (s *MemoryStore) Run(ctx context.Context) {
	ticker := time.NewTicker(s.refreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = s.Refresh(ctx)
		}
	}
}

func (s *MemoryStore) Refresh(ctx context.Context) error {

	raw, err := s.client.Fetch(ctx)
	lastError := ""
	if err != nil {
		lastError = err.Error()
		if len(raw.Matches) == 0 && len(raw.Teams) == 0 {
			return err
		}
	}

	view := buildView(raw, s.refreshInterval, lastError)

	s.mu.Lock()
	s.raw = raw
	s.view = view
	s.lastError = lastError
	s.mu.Unlock()
	return err
}

func (s *MemoryStore) Snapshot() ViewState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.view
}
