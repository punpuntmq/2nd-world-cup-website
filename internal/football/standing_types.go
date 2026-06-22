package football

import (
	"sort"
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

func BuildWCStanding(raw RawState) *WCStanding {
	s := &WCStanding{
		GeneratedAt: time.Now().UTC(),
		Groups:      make(map[string]*GroupStanding),
	}

	for _, match := range SortedMatches(raw.Matches) {
		s.ApplyMatch(match)
	}

	s.Sort()
	return s
}

func (s *WCStanding) ApplyMatch(m Match) {
	if m.Stage != "GROUP_STAGE" {
		return
	}
	if m.Status != "FINISHED" {
		return
	}
	if m.Score.FullTime.Home == nil || m.Score.FullTime.Away == nil {
		return
	}

	groupKey := *m.Group

	table, ok := s.Groups[groupKey]
	if !ok {
		table = &GroupStanding{Group: groupKey}
		s.Groups[groupKey] = table
	}

	home := table.getOrCreate(m.HomeTeam)
	away := table.getOrCreate(m.AwayTeam)

	homeGoals := *m.Score.FullTime.Home
	awayGoals := *m.Score.FullTime.Away

	home.Played++
	away.Played++

	home.GoalsFor += homeGoals
	home.GoalsAgainst += awayGoals

	away.GoalsFor += awayGoals
	away.GoalsAgainst += homeGoals

	switch {
	case homeGoals > awayGoals:
		home.Won++
		home.Points += 3
		away.Lost++
	case homeGoals < awayGoals:
		away.Won++
		away.Points += 3
		home.Lost++
	default:
		home.Draw++
		away.Draw++
		home.Points++
		away.Points++
	}

	home.GoalDifference = home.GoalsFor - home.GoalsAgainst
	away.GoalDifference = away.GoalsFor - away.GoalsAgainst
}

func (g *GroupStanding) getOrCreate(team TeamSummary) *StandingRow {
	for i := range g.Table {
		if g.Table[i].TeamID == team.ID {
			return &g.Table[i]
		}
	}

	g.Table = append(g.Table, StandingRow{
		TeamID: team.ID,
		Name:   team.Name,
		TLA:    team.TLA,
		Crest:  team.Crest,
	})

	return &g.Table[len(g.Table)-1]
}

func (s *WCStanding) Sort() {
	for _, group := range s.Groups {
		sort.SliceStable(group.Table, func(i, j int) bool {
			a := group.Table[i]
			b := group.Table[j]

			if a.Points != b.Points {
				return a.Points > b.Points
			}
			if a.GoalDifference != b.GoalDifference {
				return a.GoalDifference > b.GoalDifference
			}
			if a.GoalsFor != b.GoalsFor {
				return a.GoalsFor > b.GoalsFor
			}
			return a.Name < b.Name
		})

		for i := range group.Table {
			group.Table[i].Position = i + 1
		}
	}
}
