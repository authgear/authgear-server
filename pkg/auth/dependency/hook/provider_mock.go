package hook

import (
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
)

type MockProvider struct {
}

func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (MockProvider) DispatchEvent(payload event.Payload, user *model.User) error {
	// TODO(webhook): test impl
	return nil
}

var _ Provider = &MockProvider{}
