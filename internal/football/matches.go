package football

import (
	"sort"
	"strings"
	"time"
)

func MatchesByID(matches []Match) map[int]Match {
	byID := make(map[int]Match, len(matches))
	for _, match := range matches {
		if match.ID == 0 {
			continue
		}
		byID[match.ID] = match
	}
	return byID
}

func SortedMatches(matches map[int]Match) []Match {
	ids := make([]int, 0, len(matches))
	for id := range matches {
		ids = append(ids, id)
	}

	sort.SliceStable(ids, func(i, j int) bool {
		a := matches[ids[i]]
		b := matches[ids[j]]

		aPriority := matchSortPriority(a.Status)
		bPriority := matchSortPriority(b.Status)
		if aPriority != bPriority {
			return aPriority < bPriority
		}

		aKickoff := matchKickoffTime(a)
		bKickoff := matchKickoffTime(b)
		switch {
		case aKickoff.IsZero() && bKickoff.IsZero():
			return false
		case aKickoff.IsZero():
			return false
		case bKickoff.IsZero():
			return true
		default:
			return aKickoff.Before(bKickoff)
		}
	})

	sorted := make([]Match, 0, len(ids))
	for _, id := range ids {
		sorted = append(sorted, matches[id])
	}
	return sorted
}

func matchSortPriority(status string) int {
	switch strings.ToUpper(status) {
	case "IN_PLAY", "LIVE", "PAUSED", "BREAK", "EXTRA_TIME", "PENALTY_SHOOTOUT":
		return 0
	case "TIMED", "SCHEDULED", "SUSPENDED", "POSTPONED", "CANCELLED":
		return 1
	case "FINISHED", "AWARDED":
		return 2
	default:
		return 1
	}
}

func matchKickoffTime(match Match) time.Time {
	if strings.TrimSpace(match.UTCDate) == "" {
		return time.Time{}
	}
	kickoff, err := time.Parse(time.RFC3339, match.UTCDate)
	if err != nil {
		return time.Time{}
	}
	return kickoff.UTC()
}
