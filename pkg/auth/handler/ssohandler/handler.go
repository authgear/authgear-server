package ssohandler

import (
	"time"
)

// nolint: deadcode
var (
	zeroTime time.Time
	timeNow  = func() time.Time { return time.Now().UTC() }
)
