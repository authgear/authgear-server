package hook

import (
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/db"
)

type Provider interface {
	db.TransactionHook
	DispatchEvent(payload event.Payload, user *model.User) error
}
