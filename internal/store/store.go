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
	refreshStatusError       = "error"
)

type FootballClient interface {
	Fetch(context.Context, []football.Request) (football.RawState, error)
}

type StoreOptions struct {
	RefreshInterval time.Duration
	Quota           QuotaManager
}

type MemoryStore struct {
	mu              sync.RWMutex
	client          FootballClient
	refreshInterval time.Duration
	quota           QuotaManager
	raw             football.RawState
	refreshing      bool
	refreshStatus   string
	lastRefreshAt   time.Time
	nextAllowedAt   time.Time
	lastError       string
}

func NewMemoryStore(client FootballClient, options StoreOptions) *MemoryStore {
	refreshInterval := options.RefreshInterval
	if refreshInterval <= 0 {
		refreshInterval = 30 * time.Second
	}
	if options.Quota == nil {
		panic("store quota manager is required")
	}
	s := &MemoryStore{
		client:          client,
		refreshInterval: refreshInterval,
		quota:           options.Quota,
		refreshStatus:   refreshStatusIdle,
	}
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

func (s *MemoryStore) InitialRefresh(ctx context.Context) error {
	return s.refresh(ctx, InitialSnapshot)
}

func (s *MemoryStore) Refresh(ctx context.Context) error {
	request := RefreshSnapshot

	isEmpty := len(s.raw.Matches) == 0 && len(s.raw.Teams) == 0

	if isEmpty {
		request = InitialSnapshot
	}
	return s.refresh(ctx, request)
}

func (s *MemoryStore) refresh(ctx context.Context, kind RequestKind) error {
	// Kiểm tra refresh
	now := time.Now().UTC()

	s.mu.Lock()
	if s.refreshing {
		s.mu.Unlock()
		return ErrRefreshInProgress
	}
	// Kiểm tra hết api call
	plan := BuildRequestPlan(kind, 0)

	if !s.quota.Allow(now, plan.Cost) {
		s.refreshStatus = refreshStatusRateLimited
		s.nextAllowedAt = s.quota.NextAllowedAt(now, plan.Cost)
		nextAllowedAt := s.nextAllowedAt
		s.mu.Unlock()
		log.Printf("refresh skipped because rate limited; next allowed at %s", nextAllowedAt.Format(time.RFC3339))
		return ErrRateLimited
	}

	s.refreshing = true
	s.refreshStatus = refreshStatusRefreshing
	s.lastError = ""
	s.nextAllowedAt = s.quota.NextAllowedAt(now, plan.Cost)
	s.mu.Unlock()

	raw, err := s.client.Fetch(ctx, plan.Requests)

	now = time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()

	s.refreshing = false

	if err != nil {
		s.lastError = err.Error()
		s.refreshStatus = refreshStatusError
		log.Printf("refresh failed: %v", err)
		return err
	}
	if raw.Teams == nil {
		raw.Teams = s.raw.Teams
	}
	s.raw = raw
	s.lastError = ""
	s.refreshStatus = refreshStatusIdle

	s.lastRefreshAt = raw.FetchedAt
	if s.lastRefreshAt.IsZero() {
		s.lastRefreshAt = now
	}

	log.Printf("refresh succeeded: %d matches, %d teams", len(s.raw.Matches), len(s.raw.Teams))
	return nil
}

func (s *MemoryStore) Snapshot() ViewState {
	now := time.Now().UTC()

	// Copy raw data và refresh meta dưới sự bảo vệ của RLock.
	// Nhả khóa ngay sau khi copy để không block refresh goroutine.
	s.mu.RLock()
	raw := s.raw
	refreshInterval := s.refreshInterval
	refreshMeta := s.refreshMetaLocked(now)
	s.mu.RUnlock()

	// Tính toán ViewState ĐỘNG dựa trên raw data + thời gian hiện tại.
	// MapViewState là pure function, an toàn khi chạy ngoài lock.
	view := MapViewState(raw, refreshInterval)

	// Đắp refresh meta vào view
	view.Meta.RefreshStatus = refreshMeta.Status
	view.Meta.RemainingCalls = refreshMeta.RemainingCalls
	view.Meta.QuotaLimit = refreshMeta.QuotaLimit
	view.Meta.NextAllowedRefreshAt = formatTime(refreshMeta.NextAllowedRefreshAt)
	view.Meta.LastRefreshAt = formatTime(refreshMeta.LastRefreshAt)
	view.Meta.LastError = refreshMeta.LastError
	view.Meta.IsStale = refreshMeta.IsStale

	return view
}

func (s *MemoryStore) effectiveRefreshStatusLocked(now time.Time) string {
	status := s.refreshStatus
	if status == "" {
		status = refreshStatusIdle
	}

	if status == refreshStatusRateLimited {
		if !s.nextAllowedAt.IsZero() && !now.Before(s.nextAllowedAt) {
			return refreshStatusIdle
		}
		return refreshStatusRateLimited
	}
	return status
}

func (s *MemoryStore) refreshMetaLocked(now time.Time) RefreshMeta {
	return RefreshMeta{
		Status:               s.effectiveRefreshStatusLocked(now),
		RemainingCalls:       s.quota.Remaining(now),
		QuotaLimit:           s.quota.Limit(),
		NextAllowedRefreshAt: s.nextAllowedAt,
		LastRefreshAt:        s.lastRefreshAt,
		LastError:            s.lastError,
	}
}
