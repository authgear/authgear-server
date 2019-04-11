package auth

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/auth/response"

	"github.com/skygeario/skygear-server/pkg/core/db"
	"github.com/skygeario/skygear-server/pkg/core/handler"
)

type HookHandler interface {
	DecodeRequest(request *http.Request) (handler.RequestPayload, error)
	WithTx() bool

	ExecBeforeHooks(payload interface{}, user *response.User) error
	HandleRequest(payload interface{}, user *response.User) (interface{}, error)
	ExecAfterHooks(payload interface{}, user response.User) error
}

type hookExecutor struct {
	handler HookHandler
}

func HookHandlerToAPIHandler(h HookHandler, txContext db.TxContext) http.Handler {
	executor := hookExecutor{
		handler: h,
	}
	return handler.APIHandlerToHandler(executor, txContext)
}

func (h hookExecutor) WithTx() bool {
	return h.handler.WithTx()
}

func (h hookExecutor) DecodeRequest(request *http.Request) (handler.RequestPayload, error) {
	return h.handler.DecodeRequest(request)
}

func (h hookExecutor) Handle(req interface{}) (interface{}, error) {
	var user response.User
	err := h.handler.ExecBeforeHooks(req, &user)
	if err != nil {
		return nil, err
	}
	resp, err := h.handler.HandleRequest(req, &user)
	if err != nil {
		return nil, err
	}
	err = h.handler.ExecAfterHooks(req, user)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
