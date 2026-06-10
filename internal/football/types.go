package football

import "time"

type RawState struct {
	Source       string
	FetchedAt    time.Time
	Matches      MatchesResponse
	Teams        TeamsResponse
	Standings    StandingsResponse
}

type Area struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
	Flag string `json:"flag"`
}

type Competition struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Code   string `json:"code"`
	Type   string `json:"type"`
	Emblem string `json:"emblem"`
}

type Season struct {
	ID              int      `json:"id"`
	StartDate       string   `json:"startDate"`
	EndDate         string   `json:"endDate"`
	CurrentMatchday int      `json:"currentMatchday"`
	Stages          []string `json:"stages"`
}

type Coach struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Nationality string `json:"nationality"`
}

type Player struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Position    string `json:"position"`
	Nationality string `json:"nationality"`
	DateOfBirth string `json:"dateOfBirth"`
	ShirtNumber *int   `json:"shirtNumber"`
}

type TeamRef struct {
	ID         int      `json:"id"`
	Name       string   `json:"name"`
	ShortName  string   `json:"shortName"`
	TLA        string   `json:"tla"`
	Crest      string   `json:"crest"`
	Coach      *Coach   `json:"coach"`
	LeagueRank *int     `json:"leagueRank"`
	Formation  string   `json:"formation"`
	Lineup     []Player `json:"lineup"`
	Bench      []Player `json:"bench"`
}

type Team struct {
	ID         int      `json:"id"`
	Area       Area     `json:"area"`
	Name       string   `json:"name"`
	ShortName  string   `json:"shortName"`
	TLA        string   `json:"tla"`
	Crest      string   `json:"crest"`
	Address    string   `json:"address"`
	Website    string   `json:"website"`
	Founded    int      `json:"founded"`
	ClubColors string   `json:"clubColors"`
	Venue      string   `json:"venue"`
	Coach      *Coach   `json:"coach"`
	Squad      []Player `json:"squad"`
	LastUpdated string   `json:"lastUpdated"`
}

type TeamsResponse struct {
	Count       int         `json:"count"`
	Filters     interface{} `json:"filters"`
	Competition Competition `json:"competition"`
	Season      Season      `json:"season"`
	Teams       []Team      `json:"teams"`
}

type ScoreTime struct {
	Home *int `json:"home"`
	Away *int `json:"away"`
}

type Score struct {
	Winner   string    `json:"winner"`
	Duration string    `json:"duration"`
	FullTime ScoreTime `json:"fullTime"`
	HalfTime ScoreTime `json:"halfTime"`
}

type Match struct {
	Area        Area        `json:"area"`
	Competition Competition `json:"competition"`
	Season      Season      `json:"season"`
	ID          int         `json:"id"`
	UTCDate     string      `json:"utcDate"`
	Status      string      `json:"status"`
	Minute      *int        `json:"minute"`
	InjuryTime  *int        `json:"injuryTime"`
	Attendance  *int        `json:"attendance"`
	Venue       string      `json:"venue"`
	Matchday    *int        `json:"matchday"`
	Stage       string      `json:"stage"`
	Group       *string     `json:"group"`
	LastUpdated string      `json:"lastUpdated"`
	HomeTeam    TeamRef     `json:"homeTeam"`
	AwayTeam    TeamRef     `json:"awayTeam"`
	Score       Score       `json:"score"`
}

type MatchesResponse struct {
	Filters     interface{} `json:"filters"`
	ResultSet   interface{} `json:"resultSet"`
	Competition Competition `json:"competition"`
	Matches     []Match     `json:"matches"`
}

type StandingsResponse struct {
	Filters     interface{}     `json:"filters"`
	Area        Area            `json:"area"`
	Competition Competition     `json:"competition"`
	Season      Season          `json:"season"`
	Standings   []StandingGroup `json:"standings"`
}

type StandingGroup struct {
	Stage string          `json:"stage"`
	Type  string          `json:"type"`
	Group string          `json:"group"`
	Table []StandingEntry `json:"table"`
}

type StandingEntry struct {
	Position       int     `json:"position"`
	Team           TeamRef `json:"team"`
	PlayedGames    int     `json:"playedGames"`
	Form           string  `json:"form"`
	Won            int     `json:"won"`
	Draw           int     `json:"draw"`
	Lost           int     `json:"lost"`
	Points         int     `json:"points"`
	GoalsFor       int     `json:"goalsFor"`
	GoalsAgainst   int     `json:"goalsAgainst"`
	GoalDifference int     `json:"goalDifference"`
}
