package service

import (
	"fmt"

	"worldcup-realtime/internal/football"
)

func matchFingerprint(m football.Match) string {
	homeScore := -1
	if m.Score.FullTime.Home != nil {
		homeScore = *m.Score.FullTime.Home
	}
	awayScore := -1
	if m.Score.FullTime.Away != nil {
		awayScore = *m.Score.FullTime.Away
	}

	return fmt.Sprintf("%d|%s|%d|%d|%s", m.ID, m.Status, homeScore, awayScore, m.LastUpdated)
}

func rawChanged(before, after football.RawState) bool {
	if len(before.Matches) != len(after.Matches) {
		return true
	}
	for id, bMatch := range before.Matches {
		aMatch, ok := after.Matches[id]
		if !ok {
			return true
		}
		if matchFingerprint(bMatch) != matchFingerprint(aMatch) {
			return true
		}
	}

	// For teams, usually checking length is enough as they rarely change during a tournament.
	// But we can check IDs just to be safe.
	if len(before.Teams) != len(after.Teams) {
		return true
	}
	bTeamIDs := make(map[int]struct{}, len(before.Teams))
	for _, t := range before.Teams {
		bTeamIDs[t.ID] = struct{}{}
	}
	for _, t := range after.Teams {
		if _, ok := bTeamIDs[t.ID]; !ok {
			return true
		}
	}

	return false
}
