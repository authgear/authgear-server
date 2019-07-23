package hook

import (
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
)

type providerImpl struct {
}

func NewProvider() Provider {
	return &providerImpl{}
}

func (providerImpl) DispatchEvent(eventType event.Type, payload event.Payload, user *model.User) error {
	// TODO(webhook): real impl
	return nil
}
