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

func LiveMatchRequests(matchID int) []football.Request {
	return []football.Request{
		{
			Endpoint: football.Live{TeamID: strconv.Itoa(matchID)},
			Target: func(raw *football.RawState) any {
				return &liveMatchResourceTarget{raw: raw}
			},
		},
	}
}

func MapViewState(raw football.RawState, refreshInterval time.Duration, refreshMeta RefreshMeta) ViewState {
	return BuildViewState(raw, refreshInterval, refreshMeta)
}

type matchesResourceTarget struct {
	raw *football.RawState
}

func (t *matchesResourceTarget) UnmarshalJSON(data []byte) error {
	var payload struct {
		Area        football.Area        `json:"area"`
		Competition football.Competition `json:"competition"`
		Season      football.Season      `json:"season"`
		Matches     []football.Match     `json:"matches"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	if payload.Matches == nil {
		var direct []football.Match
		if err := json.Unmarshal(data, &direct); err != nil {
			return err
		}
		payload.Matches = direct
	}

	t.raw.Matches = football.MatchesByID(payload.Matches)
	if payload.Area.ID != 0 {
		t.raw.Area = payload.Area
	}
	if payload.Competition.ID != 0 {
		t.raw.Competition = payload.Competition
	}
	if payload.Season.ID != 0 {
		t.raw.Season = payload.Season
	}
	return nil
}

type teamsResourceTarget struct {
	raw *football.RawState
}

func (t *teamsResourceTarget) UnmarshalJSON(data []byte) error {
	var payload struct {
		Area        football.Area        `json:"area"`
		Competition football.Competition `json:"competition"`
		Season      football.Season      `json:"season"`
		Teams       []football.Team      `json:"teams"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	if payload.Teams == nil {
		var direct []football.Team
		if err := json.Unmarshal(data, &direct); err != nil {
			return err
		}
		payload.Teams = direct
	}

	t.raw.Teams = payload.Teams
	if payload.Area.ID != 0 {
		t.raw.Area = payload.Area
	}
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
