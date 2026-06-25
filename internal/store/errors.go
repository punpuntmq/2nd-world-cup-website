package store

import "errors"

var (
	ErrRefreshInProgress = errors.New("refresh in progress")
	ErrRateLimited       = errors.New("rate limited")
)
