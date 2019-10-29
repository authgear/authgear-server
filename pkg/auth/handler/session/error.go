package session

import (
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

var SessionNotFound = skyerr.NotFound.WithReason("SessionNotFound")

var errSessionNotFound = SessionNotFound.New("session not found")
