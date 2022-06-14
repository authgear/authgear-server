package meter

type CounterStore interface {
	TrackActiveUser(userID string) error
	TrackPageView(visitorID string, pageType PageType) error
}

// Service provides methods for the app to record analytic count
type Service struct {
	Counter CounterStore
}

func (s *Service) TrackActiveUser(userID string) error {
	return s.Counter.TrackActiveUser(userID)
}

func (s *Service) TrackPageView(visitorID string, pageType PageType) error {
	return s.Counter.TrackPageView(visitorID, pageType)
}
