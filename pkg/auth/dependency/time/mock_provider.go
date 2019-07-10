package time

import "time"

type MockProvider struct {
	TimeNowUTC time.Time
}

func (provider MockProvider) NowUTC() time.Time {
	return provider.TimeNowUTC
}

var _ Provider = MockProvider{}
