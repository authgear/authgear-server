package messaging

import "github.com/authgear/authgear-server/pkg/lib/ratelimit"

type message struct {
	logger       Logger
	limiter      RateLimiter
	reservations []*ratelimit.Reservation
	isSent       bool
}

func (m *message) Close() {
	if m.isSent {
		return
	}

	for _, r := range m.reservations {
		err := m.limiter.Cancel(r)
		if err != nil {
			m.logger.WithError(err).Warn("failed to return reserved token")
		}
	}
	m.reservations = nil
}
