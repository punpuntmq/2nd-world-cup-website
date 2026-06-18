package store

import (
	"time"

	"worldcup-realtime/internal/football"
)

func MapViewState(raw football.RawState, refreshInterval time.Duration, lastError string) ViewState {
	return BuildViewState(raw, refreshInterval, lastError)
}
