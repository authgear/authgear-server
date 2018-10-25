package handler

import (
	"time"
)

var timeNow = func() time.Time { return time.Now().UTC() }
