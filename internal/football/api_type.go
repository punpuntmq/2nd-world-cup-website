package football

import (
	"time"
)

type Season struct {
	ID              int          `json:"id"`
	StartDate       string       `json:"startDate"`
	EndDate         string       `json:"endDate"`
	CurrentMatchday *int         `json:"currentMatchday"`
	Winner          *TeamSummary `json:"winner"`
}

type Area struct {
	ID   int     `json:"id"`
	Name string  `json:"name"`
	Code string  `json:"code"`
	Flag *string `json:"flag"`
}

type Competition struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Code   string `json:"code"`
	Type   string `json:"type"`
	Emblem string `json:"emblem"`
}

type RawState struct {
	Source    string
	FetchedAt time.Time

	Matches map[int]Match
	Teams   []Team

	Area        Area
	Season      Season
	Competition Competition
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
}

type TeamSummary struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
	TLA       string `json:"tla"`
	Crest     string `json:"crest"`
}

type Team struct {
	TeamSummary

	Coach *Coach   `json:"coach"`
	Squad []Player `json:"squad"`
}

type ScoreTime struct {
	Home *int `json:"home"`
	Away *int `json:"away"`
}

type Score struct {
	Winner   *string   `json:"winner"`
	Duration string    `json:"duration"`
	FullTime ScoreTime `json:"fullTime"`
	HalfTime ScoreTime `json:"halfTime"`
}

type Match struct {
	ID       int     `json:"id"`
	UTCDate  string  `json:"utcDate"`
	Status   string  `json:"status"`
	Matchday *int    `json:"matchday"`
	Stage    string  `json:"stage"`
	Group    *string `json:"group"`

	HomeTeam TeamSummary `json:"homeTeam"`
	AwayTeam TeamSummary `json:"awayTeam"`

	Score Score `json:"score"`
}

/*
type RawState struct {
	Source    string
	FetchedAt time.Time
	Matches   MatchesResponse
	Teams     TeamsResponse
}

type Area struct {
	ID   int     `json:"id"`
	Name string  `json:"name"`
	Code string  `json:"code"`
	Flag *string `json:"flag"`
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
	CurrentMatchday *int     `json:"currentMatchday"`
	Winner          *TeamSummary `json:"winner"`
}

type Coach struct {
	ID          int        `json:"id"`
	FirstName   string     `json:"firstName"`
	LastName    string     `json:"lastName"`
	Name        string     `json:"name"`
	DateOfBirth string     `json:"dateOfBirth"`
	Nationality string     `json:"nationality"`
	Contract    Contract   `json:"contract"`
}

type Contract struct {
	Start *string `json:"start"`
	Until *string `json:"until"`
}

type Player struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Position    string `json:"position"`
	DateOfBirth string `json:"dateOfBirth"`
	Nationality string `json:"nationality"`
}

type TeamSummary struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
	TLA       string `json:"tla"`
	Crest     string `json:"crest"`
}

type Team struct {
	ID                  int          `json:"id"`
	Area                Area         `json:"area"`
	Name                string       `json:"name"`
	ShortName           string       `json:"shortName"`
	TLA                 string       `json:"tla"`
	Crest               string       `json:"crest"`
	Address             string       `json:"address"`
	Website             string       `json:"website"`
	Founded             int          `json:"founded"`
	ClubColors          string       `json:"clubColors"`
	Venue               *string      `json:"venue"`
	RunningCompetitions []Competition `json:"runningCompetitions"`
	Coach               *Coach       `json:"coach"`
	Squad               []Player     `json:"squad"`
	Staff               []StaffMember `json:"staff"`
	LastUpdated         string       `json:"lastUpdated"`
}

type StaffMember struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Nationality string `json:"nationality"`
}

type MatchesResponse struct {
	Filters     json.RawMessage `json:"filters"`
	ResultSet   json.RawMessage `json:"resultSet"`
	Competition Competition     `json:"competition"`
	Matches     []Match         `json:"matches"`
}

type TeamsResponse struct {
	Count       int           `json:"count"`
	Filters     json.RawMessage `json:"filters"`
	Competition Competition   `json:"competition"`
	Season      Season        `json:"season"`
	Teams       []Team        `json:"teams"`
}

type ScoreTime struct {
	Home *int `json:"home"`
	Away *int `json:"away"`
}

type Score struct {
	Winner   *string   `json:"winner"`
	Duration string    `json:"duration"`
	FullTime ScoreTime `json:"fullTime"`
	HalfTime ScoreTime `json:"halfTime"`
}

type Referee struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Nationality string `json:"nationality"`
}

type Match struct {
	Area        Area         `json:"area"`
	Competition Competition  `json:"competition"`
	Season      Season       `json:"season"`
	ID          int          `json:"id"`
	UTCDate     string       `json:"utcDate"`
	Status      string       `json:"status"`
	Matchday    *int         `json:"matchday"`
	Stage       string       `json:"stage"`
	Group       *string      `json:"group"`
	LastUpdated string       `json:"lastUpdated"`

	HomeTeam TeamSummary `json:"homeTeam"`
	AwayTeam TeamSummary `json:"awayTeam"`

	Score    Score     `json:"score"`
	Referees []Referee `json:"referees,omitempty"`
}
*/
