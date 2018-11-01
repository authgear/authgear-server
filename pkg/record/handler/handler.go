package handler

import (
	"time"
)

var zeroTime time.Time
var timeNow = func() time.Time { return time.Now().UTC() }
