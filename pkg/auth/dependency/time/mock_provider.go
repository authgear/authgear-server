package time

import "time"

type MockProvider struct {
	TimeNowUTC time.Time
	TimeNow    time.Time
}

func (provider *MockProvider) NowUTC() time.Time {
	return provider.TimeNowUTC
}

func (provider *MockProvider) Now() time.Time {
	return provider.TimeNow
}

func (provider *MockProvider) AdvanceSeconds(seconds int) {
	provider.TimeNowUTC = provider.TimeNowUTC.Add(time.Duration(seconds) * time.Second)
	provider.TimeNow = provider.TimeNow.Add(time.Duration(seconds) * time.Second)
}

var _ Provider = &MockProvider{}
