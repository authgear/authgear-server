package hook

import (
	"github.com/skygeario/skygear-server/pkg/auth/event"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/handler"
)

type Provider interface {
	WillCommitTx() error
	DidCommitTx()
	DispatchEvent(payload event.Payload, user *model.User) error
}

func WrapHandler(provider Provider, handler handler.APIHandler) handler.APIHandler {
	return hookHandler{
		APIHandler: handler,
		provider:   provider,
	}
}
