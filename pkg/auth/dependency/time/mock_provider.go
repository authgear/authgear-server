package time

import "time"

type MockProvider struct {
	TimeNowUTC time.Time
	TimeNow    time.Time
}

func (provider MockProvider) NowUTC() time.Time {
	return provider.TimeNowUTC
}

func (provider MockProvider) Now() time.Time {
	return provider.TimeNow
}

var _ Provider = MockProvider{}
