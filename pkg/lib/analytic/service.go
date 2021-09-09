package analytic

type CounterStore interface {
	TrackActiveUser(userID string) error
}

type Service struct {
	Counter CounterStore
}

func (s *Service) TrackActiveUser(userID string) error {
	return s.Counter.TrackActiveUser(userID)
}
