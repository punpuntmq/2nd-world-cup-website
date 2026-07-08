package service

import (
	"encoding/json"
	"time"

	"worldcup-realtime/internal/football"
)

type RequestKind string

const (
	InitialSnapshot RequestKind = "initial_snapshot"
	RefreshSnapshot RequestKind = "refresh_snapshot"
)

type RequestPlan struct {
	Kind     RequestKind
	Requests []football.Request
	Cost     int
}

func BuildRequestPlan(kind RequestKind, matchID int) RequestPlan {
	var requests []football.Request
	switch kind {
	case InitialSnapshot:
		requests = StaticCompetitionRequests()
	case RefreshSnapshot:
		requests = RefreshCompetitionRequest()
	}
	return RequestPlan{
		Kind:     kind,
		Requests: requests,
		Cost:     len(requests),
	}
}

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

func RefreshCompetitionRequest() []football.Request {
	return []football.Request{
		{
			Endpoint: football.CompetitionEndpoint{Resource: football.MatchesResource},
			Target: func(raw *football.RawState) any {
				return &matchesResourceTarget{raw: raw}
			},
		},
	}
}

func MapViewState(raw football.RawState, refreshInterval time.Duration) ViewState {
	return BuildViewState(raw, refreshInterval)
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
