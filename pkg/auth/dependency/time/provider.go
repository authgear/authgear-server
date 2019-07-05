package time

import "time"

type Provider interface {
	Now() time.Time
}

type providerImpl struct{}

func NewProvider() Provider {
	return providerImpl{}
}

func (provider providerImpl) Now() time.Time {
	return time.Now()
}
