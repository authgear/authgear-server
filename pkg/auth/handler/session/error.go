package session

import (
	skyerr "github.com/skygeario/skygear-server/pkg/core/xskyerr"
)

var SessionNotFound = skyerr.NotFound.WithReason("SessionNotFound")

var errSessionNotFound = SessionNotFound.New("session not found")
