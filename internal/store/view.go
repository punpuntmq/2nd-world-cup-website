package store

import "time"

type ViewState struct {
	Meta  MetaView   `json:"meta"`
	Teams []TeamView `json:"teams"`

	CurrentMatch *MatchView `json:"currentMatch"`
	NextMatch    *MatchView `json:"nextMatch"`

	Highlight HighlightView `json:"highlight"`

	LiveMatches     []MatchView `json:"liveMatches"`
	UpcomingMatches []MatchView `json:"upcomingMatches"`
	FinishedMatches []MatchView `json:"finishedMatches"`
	SpecialMatches  []MatchView `json:"specialMatches"`

	Standings []StandingGroupView `json:"standings"`
}

type MetaView struct {
	CompetitionName      string `json:"competitionName"`
	CompetitionEmblem    string `json:"competitionEmblem"`
	Source               string `json:"source"`
	FetchedAt            string `json:"fetchedAt"`
	LastError            string `json:"lastError,omitempty"`
	RefreshSeconds       int    `json:"refreshSeconds"`
	LiveCount            int    `json:"liveCount"`
	RefreshStatus        string `json:"refreshStatus"`
	RemainingCalls       int    `json:"remainingCalls"`
	NextAllowedRefreshAt string `json:"nextAllowedRefreshAt,omitempty"`
	LastRefreshAt        string `json:"lastRefreshAt,omitempty"`
	IsStale              bool   `json:"isStale"`
}

type RefreshMeta struct {
	Status               string
	RemainingCalls       int
	NextAllowedRefreshAt time.Time
	LastRefreshAt        time.Time
	LastError            string
	IsStale              bool
}

type TeamView struct {
	ID        int          `json:"id"`
	Name      string       `json:"name"`
	ShortName string       `json:"shortName"`
	TLA       string       `json:"tla"`
	Crest     string       `json:"crest"`
	Coach     string       `json:"coach"`
	Group     string       `json:"group"`
	Rank      int          `json:"rank"`
	Played    int          `json:"played"`
	Points    int          `json:"points"`
	GoalDiff  int          `json:"goalDiff"`
	Squad     []PlayerView `json:"squad"`
}

type PlayerView struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Position    string `json:"position"`
	Nationality string `json:"nationality"`
}

type TeamMiniView struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
	TLA       string `json:"tla"`
	Crest     string `json:"crest"`
}

type MatchView struct {
	ID           int          `json:"id"`
	HomeTeam     TeamMiniView `json:"homeTeam"`
	AwayTeam     TeamMiniView `json:"awayTeam"`
	HomeScore    string       `json:"homeScore"`
	AwayScore    string       `json:"awayScore"`
	Status       string       `json:"status"`
	StatusLabel  string       `json:"statusLabel"`
	StatusGroup  string       `json:"statusGroup"`
	Minute       string       `json:"minute"`
	KickoffUTC   string       `json:"kickoffUTC"`
	Stage        string       `json:"stage"`
	Group        string       `json:"group"`
	Matchday     string       `json:"matchday"`
	SortTimeUnix int64        `json:"-"`
}

type HighlightView struct {
	Title   string `json:"title"`
	Message string `json:"message"`
	Detail  string `json:"detail"`
}

type StandingGroupView struct {
	Group       string            `json:"group"`
	MemberCount int               `json:"memberCount"`
	Rows        []StandingRowView `json:"rows"`
}

type StandingRowView struct {
	Position       int          `json:"position"`
	Team           TeamMiniView `json:"team"`
	PlayedGames    int          `json:"playedGames"`
	Won            int          `json:"won"`
	Draw           int          `json:"draw"`
	Lost           int          `json:"lost"`
	GoalsFor       int          `json:"goalsFor"`
	GoalsAgainst   int          `json:"goalsAgainst"`
	GoalDifference int          `json:"goalDifference"`
	Points         int          `json:"points"`
	Form           string       `json:"form"`
}
