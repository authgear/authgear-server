package hook

import (
	"github.com/skygeario/skygear-server/pkg/core/handler"
)

type hookHandler struct {
	handler.APIHandler

	provider Provider
}

func (handler hookHandler) WillCommitTx() error {
	return handler.provider.WillCommitTx()
}

func (handler hookHandler) DidCommitTx() {
	handler.provider.DidCommitTx()
}
