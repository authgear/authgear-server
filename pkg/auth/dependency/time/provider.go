package time

import "time"

type Provider interface {
	NowUTC() time.Time
}

type providerImpl struct{}

func NewProvider() Provider {
	return providerImpl{}
}

func (provider providerImpl) NowUTC() time.Time {
	return time.Now().UTC()
}
