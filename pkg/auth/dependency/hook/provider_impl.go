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

func (provider *providerImpl) DispatchEvent(payload event.Payload, user *model.User) error {
	// TODO(webhook): real impl
	return nil
}

func (provider *providerImpl) WillCommitTx() error {
	// TODO(webhook): real impl
	return nil
}

func (provider *providerImpl) DidCommitTx() {

}
