package football

import (
	"time"
)

type WCStanding struct {
	GeneratedAt time.Time
	Groups      map[string]*GroupStanding
}

type GroupStanding struct {
	Group string
	Table []StandingRow
}

type StandingRow struct {
	Position int

	TeamID int
	Name   string
	TLA    string
	Crest  string

	Played int
	Won    int
	Draw   int
	Lost   int

	GoalsFor       int
	GoalsAgainst   int
	GoalDifference int

	Points int
}
