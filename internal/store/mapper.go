package store

import (
	"encoding/json"
	"strconv"
	"time"

	"worldcup-realtime/internal/football"
)

func StaticCompetitionRequests() []football.Request {
	return []football.Request{
		{
			Endpoint: football.CompetitionEndpoint{Resource: football.MatchesResource},
			Target: func(raw *football.RawState) any {
				return &matchesResourceTarget{raw: raw}
			},
		},
		{
			Endpoint: football.CompetitionEndpoint{Resource: football.TeamsResource},
			Target: func(raw *football.RawState) any {
				return &teamsResourceTarget{raw: raw}
			},
		},
	}
}

func LiveMatchRequests(matches []football.Match) []football.Request {
	requests := make([]football.Request, 0, len(matches))
	for _, match := range matches {
		if match.ID == 0 {
			continue
		}
		requests = append(requests, football.Request{
			Endpoint: football.Live{MatchID: strconv.Itoa(match.ID)},
			Target: func(raw *football.RawState) any {
				return &liveMatchResourceTarget{raw: raw}
			},
		})
	}
	return requests
}

func MapViewState(raw football.RawState, refreshInterval time.Duration, refreshMeta RefreshMeta) ViewState {
	return BuildViewState(raw, refreshInterval, refreshMeta)
}

type matchesResourceTarget struct {
	raw *football.RawState
}

func (t *matchesResourceTarget) UnmarshalJSON(data []byte) error {
	var payload struct {
		Competition football.Competition `json:"competition"`
		Matches     []football.Match     `json:"matches"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	t.raw.Matches = football.MatchesByID(payload.Matches)
	if payload.Competition.ID != 0 {
		t.raw.Competition = payload.Competition
	}
	return nil
}

type teamsResourceTarget struct {
	raw *football.RawState
}

func (t *teamsResourceTarget) UnmarshalJSON(data []byte) error {
	var payload struct {
		Competition football.Competition `json:"competition"`
		Season      football.Season      `json:"season"`
		Teams       []football.Team      `json:"teams"`
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	t.raw.Teams = payload.Teams
	if payload.Competition.ID != 0 {
		t.raw.Competition = payload.Competition
	}

	if payload.Season.ID != 0 {
		t.raw.Season = payload.Season
	}

	return nil
}

type liveMatchResourceTarget struct {
	raw *football.RawState
}

func (t *liveMatchResourceTarget) UnmarshalJSON(data []byte) error {
	var metadata struct {
		Area        football.Area        `json:"area"`
		Competition football.Competition `json:"competition"`
		Season      football.Season      `json:"season"`
	}

	if err := json.Unmarshal(data, &metadata); err != nil {
		return err
	}

	var match football.Match
	if err := json.Unmarshal(data, &match); err != nil {
		return err
	}
	if match.ID != 0 {
		t.raw.Matches = map[int]football.Match{match.ID: match}
	}

	if metadata.Area.ID != 0 {
		t.raw.Area = metadata.Area
	}
	if metadata.Competition.ID != 0 {
		t.raw.Competition = metadata.Competition
	}
	if metadata.Season.ID != 0 {
		t.raw.Season = metadata.Season
	}

	return nil
}
