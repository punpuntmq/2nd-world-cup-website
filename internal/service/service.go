package service

import (
	"context"
	"log"
	"sync"
	"time"

	"worldcup-realtime/internal/football"
	"worldcup-realtime/internal/store"
)

type FootballClient interface {
	Fetch(context.Context, []football.Request) (football.RawState, error)
}

type Options struct {
	LiveRefreshInterval time.Duration
	IdleRefreshInterval time.Duration
	Quota               store.QuotaManager
}

type WorldCupService struct {
	client          FootballClient
	store               *store.MemoryStore
	liveRefreshInterval time.Duration
	idleRefreshInterval time.Duration
	quota               store.QuotaManager

	mu         sync.Mutex
	refreshing bool

	cachedView   ViewState
	cachedViewAt time.Time
	cachedViewMu sync.RWMutex
}

func NewWorldCupService(client FootballClient, snapshotStore *store.MemoryStore, options Options) *WorldCupService {
	liveInterval := options.LiveRefreshInterval
	if liveInterval <= 0 {
		liveInterval = 10 * time.Second
	}
	idleInterval := options.IdleRefreshInterval
	if idleInterval <= 0 {
		idleInterval = 120 * time.Second
	}
	if options.Quota == nil {
		panic("service quota manager is required")
	}
	return &WorldCupService{
		client:              client,
		store:               snapshotStore,
		liveRefreshInterval: liveInterval,
		idleRefreshInterval: idleInterval,
		quota:               options.Quota,
	}
}

func (s *WorldCupService) InitialRefresh(ctx context.Context) (bool, error) {
	return s.refresh(ctx, InitialSnapshot)
}

func (s *WorldCupService) Refresh(ctx context.Context) (bool, error) {
	kind := RefreshSnapshot
	if s.store.IsEmpty() {
		kind = InitialSnapshot
	}
	return s.refresh(ctx, kind)
}

func (s *WorldCupService) State() ViewState {
	now := time.Now().UTC()
	snapshot := s.store.Snapshot(now)
	
	s.cachedViewMu.RLock()
	cached := s.cachedView
	cachedAt := s.cachedViewAt
	s.cachedViewMu.RUnlock()

	var view ViewState
	// If cached view is from the same refresh cycle (fetchedAt matches), use it
	if !cachedAt.IsZero() && cachedAt.Equal(snapshot.Raw.FetchedAt) {
		view = cached
	} else {
		view = MapViewState(snapshot.Raw, s.idleRefreshInterval)
		s.cachedViewMu.Lock()
		s.cachedView = view
		s.cachedViewAt = snapshot.Raw.FetchedAt
		s.cachedViewMu.Unlock()
	}

	s.applyRefreshMeta(&view, snapshot.Refresh, now)
	return view
}

func (s *WorldCupService) Team(id int) (TeamView, bool) {
	for _, team := range s.State().Teams {
		if team.ID == id {
			return team, true
		}
	}
	return TeamView{}, false
}

func (s *WorldCupService) Match(id int) (MatchView, bool) {
	return findMatch(s.State(), id)
}

func (s *WorldCupService) Matches() MatchesView {
	view := s.State()
	all := make([]MatchView, 0, len(view.LiveMatches)+len(view.UpcomingMatches)+len(view.FinishedMatches)+len(view.SpecialMatches))
	all = append(all, view.LiveMatches...)
	all = append(all, view.UpcomingMatches...)
	all = append(all, view.FinishedMatches...)
	all = append(all, view.SpecialMatches...)
	return MatchesView{
		Live:     view.LiveMatches,
		Upcoming: view.UpcomingMatches,
		Finished: view.FinishedMatches,
		Special:  view.SpecialMatches,
		All:      all,
	}
}

func (s *WorldCupService) RefreshIntervals() (live, idle time.Duration) {
	return s.liveRefreshInterval, s.idleRefreshInterval
}

func (s *WorldCupService) refresh(ctx context.Context, kind RequestKind) (bool, error) {
	if !s.beginRefresh() {
		return false, store.ErrRefreshInProgress
	}
	defer s.finishRefresh()

	now := time.Now().UTC()
	plan := BuildRequestPlan(kind, 0)

	if !s.quota.Allow(now, plan.Cost) {
		nextAllowedAt := s.quota.NextAllowedAt(now, plan.Cost)
		s.store.MarkRateLimited(nextAllowedAt)
		log.Printf("refresh skipped because rate limited; next allowed at %s", nextAllowedAt.Format(time.RFC3339))
		return false, store.ErrRateLimited
	}

	s.store.MarkRefreshing(s.quota.NextAllowedAt(now, plan.Cost))
	before := s.store.Snapshot(now).Raw

	raw, err := s.client.Fetch(ctx, plan.Requests)
	now = time.Now().UTC()
	if err != nil {
		s.store.MarkError(err.Error())
		log.Printf("refresh failed: %v", err)
		return false, err
	}

	if raw.Teams == nil {
		raw.Teams = before.Teams
	}
	if raw.Matches == nil {
		raw.Matches = before.Matches
	}

	saved := s.store.Save(raw, now)
	changed := rawChanged(before, saved)
	
	s.cachedViewMu.RLock()
	firstRun := s.cachedViewAt.IsZero()
	s.cachedViewMu.RUnlock()

	if firstRun || changed {
		view := MapViewState(saved, s.idleRefreshInterval)
		s.cachedViewMu.Lock()
		s.cachedView = view
		s.cachedViewAt = saved.FetchedAt
		s.cachedViewMu.Unlock()
		changed = true
	}

	log.Printf("refresh succeeded: %d matches, %d teams, changed=%t", len(saved.Matches), len(saved.Teams), changed)
	return changed, nil
}

func (s *WorldCupService) beginRefresh() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.refreshing {
		return false
	}
	s.refreshing = true
	return true
}

func (s *WorldCupService) finishRefresh() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.refreshing = false
}

func (s *WorldCupService) applyRefreshMeta(view *ViewState, meta store.RefreshMeta, now time.Time) {
	view.Meta.RefreshStatus = meta.Status
	view.Meta.RemainingCalls = s.quota.Remaining(now)
	view.Meta.QuotaLimit = s.quota.Limit()
	view.Meta.NextAllowedRefreshAt = formatTime(meta.NextAllowedRefreshAt)
	view.Meta.LastRefreshAt = formatTime(meta.LastRefreshAt)
	view.Meta.LastError = meta.LastError
	view.Meta.IsStale = isStale(meta.LastRefreshAt, now, s.idleRefreshInterval)
}

func isStale(lastRefreshAt time.Time, now time.Time, interval time.Duration) bool {
	if lastRefreshAt.IsZero() || interval <= 0 {
		return false
	}
	return now.After(lastRefreshAt.Add(interval * 3))
}

func findMatch(view ViewState, id int) (MatchView, bool) {
	if view.CurrentMatch != nil && view.CurrentMatch.ID == id {
		return *view.CurrentMatch, true
	}
	if view.NextMatch != nil && view.NextMatch.ID == id {
		return *view.NextMatch, true
	}
	for _, matches := range [][]MatchView{
		view.LiveMatches,
		view.UpcomingMatches,
		view.FinishedMatches,
		view.SpecialMatches,
	} {
		for _, match := range matches {
			if match.ID == id {
				return match, true
			}
		}
	}
	return MatchView{}, false
}
