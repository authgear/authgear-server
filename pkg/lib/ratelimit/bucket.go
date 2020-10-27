package ratelimit

import "time"

type Bucket struct {
	Key         string
	Size        int
	ResetPeriod time.Duration
}
