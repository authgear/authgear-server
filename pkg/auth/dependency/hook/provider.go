package hook

import (
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
)

type Provider interface {
	DispatchEvent(eventType event.Type, payload event.Payload, user *model.User) error
}
