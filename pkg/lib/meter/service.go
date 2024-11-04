package meter

import (
	"context"
)

type CounterStore interface {
	TrackActiveUser(ctx context.Context, userID string) error
	TrackPageView(ctx context.Context, visitorID string, pageType PageType) error
}

// Service provides methods for the app to record analytic count
type Service struct {
	Counter CounterStore
}

func (s *Service) TrackActiveUser(ctx context.Context, userID string) error {
	return s.Counter.TrackActiveUser(ctx, userID)
}

func (s *Service) TrackPageView(ctx context.Context, visitorID string, pageType PageType) error {
	return s.Counter.TrackPageView(ctx, visitorID, pageType)
}
