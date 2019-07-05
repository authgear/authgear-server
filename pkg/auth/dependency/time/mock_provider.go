package time

import "time"

type MockProvider struct {
	TimeNow time.Time
}

func (provider MockProvider) Now() time.Time {
	return provider.TimeNow
}

var _ Provider = MockProvider{}
