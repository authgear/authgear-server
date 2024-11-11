package messaging

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/usage"
)

type message struct {
	logger       Logger
	rateLimiter  RateLimiter
	usageLimiter UsageLimiter
	rateLimits   []*ratelimit.Reservation
	usageLimit   *usage.Reservation

	isSent bool
}

func (m *message) Close(ctx context.Context) {
	if m.isSent {
		return
	}

	for _, r := range m.rateLimits {
		m.rateLimiter.Cancel(ctx, r)
	}
	m.rateLimits = nil

	if m.usageLimit != nil {
		m.usageLimiter.Cancel(ctx, m.usageLimit)
	}
	m.usageLimit = nil
}
