package store

import (
	"sync"
	"time"

	"worldcup-realtime/internal/football"
)

const (
	RefreshStatusIdle        = "idle"
	RefreshStatusRefreshing  = "refreshing"
	RefreshStatusRateLimited = "rate_limited"
	RefreshStatusError       = "error"
)

type Snapshot struct {
	Raw     football.RawState
	Refresh RefreshMeta
}

type RefreshMeta struct {
	Status               string
	NextAllowedRefreshAt time.Time
	LastRefreshAt        time.Time
	LastError            string
	IsStale              bool
}

type MemoryStore struct {
	mu      sync.RWMutex
	raw     football.RawState
	refresh RefreshMeta
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		refresh: RefreshMeta{Status: RefreshStatusIdle},
	}
}

// Snapshot returns a shallow clone of the current state.
// Callers MUST NOT mutate pointer fields (e.g. Match.Group, Match.Matchday)
// or slice elements (e.g. Team.Squad) of the returned data.
func (s *MemoryStore) Snapshot(now time.Time) Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return Snapshot{
		Raw:     cloneRawState(s.raw),
		Refresh: s.effectiveRefreshMetaLocked(now),
	}
}

func (s *MemoryStore) IsEmpty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.raw.Matches) == 0 && len(s.raw.Teams) == 0
}

func (s *MemoryStore) MarkRefreshing(nextAllowedAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.refresh.Status = RefreshStatusRefreshing
	s.refresh.LastError = ""
	s.refresh.NextAllowedRefreshAt = nextAllowedAt
}

func (s *MemoryStore) MarkRateLimited(nextAllowedAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.refresh.Status = RefreshStatusRateLimited
	s.refresh.NextAllowedRefreshAt = nextAllowedAt
}

func (s *MemoryStore) MarkError(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.refresh.Status = RefreshStatusError
	s.refresh.LastError = message
}

// Save updates the store and returns a shallow clone of the new state.
// Callers MUST NOT mutate pointer fields or slice elements of the returned data.
func (s *MemoryStore) Save(raw football.RawState, now time.Time) football.RawState {
	s.mu.Lock()
	defer s.mu.Unlock()

	if raw.Teams == nil {
		raw.Teams = s.raw.Teams
	}
	if raw.Matches == nil {
		raw.Matches = s.raw.Matches
	}

	s.raw = cloneRawState(raw)
	s.refresh.Status = RefreshStatusIdle
	s.refresh.LastError = ""
	s.refresh.LastRefreshAt = raw.FetchedAt
	if s.refresh.LastRefreshAt.IsZero() {
		s.refresh.LastRefreshAt = now
	}
	return cloneRawState(s.raw)
}

func (s *MemoryStore) effectiveRefreshMetaLocked(now time.Time) RefreshMeta {
	meta := s.refresh
	if meta.Status == "" {
		meta.Status = RefreshStatusIdle
	}
	if meta.Status == RefreshStatusRateLimited {
		if !meta.NextAllowedRefreshAt.IsZero() && !now.Before(meta.NextAllowedRefreshAt) {
			meta.Status = RefreshStatusIdle
		}
	}
	return meta
}

func cloneRawState(raw football.RawState) football.RawState {
	cloned := raw

	if raw.Matches != nil {
		cloned.Matches = make(map[int]football.Match, len(raw.Matches))
		for id, match := range raw.Matches {
			cloned.Matches[id] = match
		}
	}

	if raw.Teams != nil {
		cloned.Teams = append([]football.Team(nil), raw.Teams...)
	}

	return cloned
}
