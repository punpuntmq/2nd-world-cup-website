package store

type ViewState struct {
	Meta           MetaView             `json:"meta"`
	Teams          []TeamView           `json:"teams"`
	CurrentMatch   *MatchView           `json:"currentMatch"`
	NextMatch      *MatchView           `json:"nextMatch"`
	Highlight      HighlightView        `json:"highlight"`
	UpcomingMatches []MatchView         `json:"upcomingMatches"`
	FinishedMatches []MatchView         `json:"finishedMatches"`
	SpecialMatches  []MatchView         `json:"specialMatches"`
	Standings       []StandingGroupView `json:"standings"`
}

type MetaView struct {
	CompetitionCode string `json:"competitionCode"`
	CompetitionName string `json:"competitionName"`
	Source          string `json:"source"`
	FetchedAt       string `json:"fetchedAt"`
	LastError       string `json:"lastError,omitempty"`
	RefreshSeconds  int    `json:"refreshSeconds"`
	LiveCount       int    `json:"liveCount"`
	UpcomingCount   int    `json:"upcomingCount"`
	FinishedCount   int    `json:"finishedCount"`
	SpecialCount    int    `json:"specialCount"`
}

type TeamView struct {
	ID          int          `json:"id"`
	Name        string       `json:"name"`
	ShortName   string       `json:"shortName"`
	TLA         string       `json:"tla"`
	Crest       string       `json:"crest"`
	Area        string       `json:"area"`
	Coach       string       `json:"coach"`
	Venue       string       `json:"venue"`
	Group       string       `json:"group"`
	Rank        int          `json:"rank"`
	Played      int          `json:"played"`
	Points      int          `json:"points"`
	GoalDiff    int          `json:"goalDiff"`
	Squad       []PlayerView `json:"squad"`
}

type PlayerView struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Position    string `json:"position"`
	Nationality string `json:"nationality"`
	ShirtNumber string `json:"shirtNumber"`
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
	Venue        string       `json:"venue"`
	Stage        string       `json:"stage"`
	Group        string       `json:"group"`
	Matchday     string       `json:"matchday"`
	SortTimeUnix int64        `json:"sortTimeUnix"`
}

type HighlightView struct {
	Title   string `json:"title"`
	Message string `json:"message"`
	Detail  string `json:"detail"`
}

type StandingGroupView struct {
	Group       string            `json:"group"`
	MemberCount int              `json:"memberCount"`
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
