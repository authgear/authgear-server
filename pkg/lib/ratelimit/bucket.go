package ratelimit

import (
	"fmt"
	"time"
)

type Kind string

const (
	KindDefault = ""
	KindRequest = "request"
	KindUsage   = "usage"
)

type Bucket struct {
	Kind        Kind
	Key         string
	Size        int
	ResetPeriod time.Duration
}

func (b Bucket) BucketError() error {
	kind := b.Kind
	if kind == KindDefault {
		kind = KindRequest
	}
	switch kind {
	case KindRequest:
		return ErrTooManyRequests
	case KindUsage:
		return ErrUsageLimitExceeded
	default:
		panic(fmt.Errorf("ratelimit: unknown kind: %v", kind))
	}
}
