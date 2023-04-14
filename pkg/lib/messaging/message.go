package messaging

import (
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

func (m *message) Close() {
	if m.isSent {
		return
	}

	for _, r := range m.rateLimits {
		err := m.rateLimiter.Cancel(r)
		if err != nil {
			m.logger.WithError(err).Warn("failed to return reserved token")
		}
	}
	m.rateLimits = nil

	if m.usageLimit != nil {
		err := m.usageLimiter.Cancel(m.usageLimit)
		if err != nil {
			m.logger.WithError(err).Warn("failed to return reserved token")
		}
	}
	m.usageLimit = nil
}
