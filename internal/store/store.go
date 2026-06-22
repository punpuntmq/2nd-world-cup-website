package store

import (
	"context"
	"log"
	"sync"
	"time"

	"worldcup-realtime/internal/football"
)

const (
	refreshStatusIdle        = "idle"
	refreshStatusRefreshing  = "refreshing"
	refreshStatusRateLimited = "rate_limited"
	refreshStatusStale       = "stale"
	refreshStatusError       = "error"

	upstreamQuotaLimit     = 10
	upstreamQuotaWindow    = time.Minute
	fullSnapshotCallCost   = 2
	refreshSkippedNoChange = "refresh skipped"
)

type FootballClient interface {
	Fetch(context.Context, []football.Request) (football.RawState, error)
}

type MemoryStore struct {
	mu              sync.RWMutex
	client          FootballClient
	refreshInterval time.Duration
	quota           *quotaWindow
	raw             football.RawState
	view            ViewState
	refreshing      bool
	refreshStatus   string
	lastRefreshAt   time.Time
	nextAllowedAt   time.Time
	isStale         bool
	lastError       string
}

func NewMemoryStore(client FootballClient, refreshInterval time.Duration) *MemoryStore {
	s := &MemoryStore{
		client:          client,
		refreshInterval: refreshInterval,
		quota:           newQuotaWindow(upstreamQuotaLimit, upstreamQuotaWindow),
		refreshStatus:   refreshStatusIdle,
	}
	s.view = MapViewState(s.raw, s.refreshInterval, s.refreshMetaLocked(time.Now().UTC()))
	return s
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
	now := time.Now().UTC()
	s.mu.Lock()
	if s.refreshing {
		s.refreshStatus = refreshStatusRefreshing
		s.isStale = hasRawData(s.raw)
		s.nextAllowedAt = s.quota.nextAllowedAt(now, fullSnapshotCallCost)
		s.view = MapViewState(s.raw, s.refreshInterval, s.refreshMetaLocked(now))
		s.mu.Unlock()
		log.Printf("%s because in progress", refreshSkippedNoChange)
		return nil
	}

	if !s.quota.allow(now, fullSnapshotCallCost) {
		s.refreshStatus = refreshStatusRateLimited
		s.isStale = hasRawData(s.raw)
		s.nextAllowedAt = s.quota.nextAllowedAt(now, fullSnapshotCallCost)
		nextAllowedAt := s.nextAllowedAt
		s.view = MapViewState(s.raw, s.refreshInterval, s.refreshMetaLocked(now))
		s.mu.Unlock()
		log.Printf("refresh skipped because rate limited; next allowed at %s", nextAllowedAt.Format(time.RFC3339))
		return nil
	}

	s.refreshing = true
	s.refreshStatus = refreshStatusRefreshing
	s.lastError = ""
	s.isStale = hasRawData(s.raw)
	s.nextAllowedAt = s.quota.nextAllowedAt(now, fullSnapshotCallCost)
	s.view = MapViewState(s.raw, s.refreshInterval, s.refreshMetaLocked(now))
	s.mu.Unlock()

	log.Printf("refresh started: fetching full competition snapshot")
	raw, err := s.client.Fetch(ctx, StaticCompetitionRequests())

	now = time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()

	s.refreshing = false
	s.nextAllowedAt = s.quota.nextAllowedAt(now, fullSnapshotCallCost)
	if err != nil {
		s.lastError = err.Error()
		s.refreshStatus = refreshStatusError
		s.isStale = hasRawData(s.raw)
		s.view = MapViewState(s.raw, s.refreshInterval, s.refreshMetaLocked(now))
		log.Printf("refresh failed: %v", err)
		return err
	}

	s.raw = raw
	s.lastError = ""
	s.refreshStatus = refreshStatusIdle
	s.lastRefreshAt = raw.FetchedAt
	if s.lastRefreshAt.IsZero() {
		s.lastRefreshAt = now
	}
	s.isStale = false
	s.view = MapViewState(s.raw, s.refreshInterval, s.refreshMetaLocked(now))
	log.Printf("refresh succeeded: %d matches, %d teams", len(s.raw.Matches), len(s.raw.Teams))
	return nil
}

func (s *MemoryStore) Snapshot() ViewState {
	now := time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()

	s.normalizeRefreshStatusLocked(now)
	view := s.view
	s.applyRefreshMetaLocked(&view, now)
	return view
}

func (s *MemoryStore) normalizeRefreshStatusLocked(now time.Time) {
	if s.refreshStatus != refreshStatusRateLimited || s.quota.remaining(now) < fullSnapshotCallCost {
		return
	}
	s.nextAllowedAt = time.Time{}
	if s.isStale {
		s.refreshStatus = refreshStatusStale
		return
	}
	s.refreshStatus = refreshStatusIdle
}

func (s *MemoryStore) refreshMetaLocked(now time.Time) RefreshMeta {
	status := s.refreshStatus
	if status == "" {
		status = refreshStatusIdle
	}
	if status == refreshStatusIdle && s.isStale {
		status = refreshStatusStale
	}
	return RefreshMeta{
		Status:               status,
		RemainingCalls:       s.quota.remaining(now),
		NextAllowedRefreshAt: s.nextAllowedAt,
		LastRefreshAt:        s.lastRefreshAt,
		LastError:            s.lastError,
		IsStale:              s.isStale,
	}
}

func (s *MemoryStore) applyRefreshMetaLocked(view *ViewState, now time.Time) {
	refreshMeta := s.refreshMetaLocked(now)
	view.Meta.RefreshStatus = refreshMeta.Status
	view.Meta.RemainingCalls = refreshMeta.RemainingCalls
	view.Meta.NextAllowedRefreshAt = formatTime(refreshMeta.NextAllowedRefreshAt)
	view.Meta.LastRefreshAt = formatTime(refreshMeta.LastRefreshAt)
	view.Meta.LastError = refreshMeta.LastError
	view.Meta.IsStale = refreshMeta.IsStale
}

func hasRawData(raw football.RawState) bool {
	return len(raw.Matches) > 0 || len(raw.Teams) > 0
}
