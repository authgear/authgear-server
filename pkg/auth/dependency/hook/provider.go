package hook

import (
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
)

type Provider interface {
	WillCommitTx() error
	DidCommitTx()
	DispatchEvent(payload event.Payload, user *model.User) error
}
